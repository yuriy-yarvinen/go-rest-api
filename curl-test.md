# curl тесты

Все команды проверены вручную против запущенного стека (`docker-compose up`).
Сервер слушает `localhost:8082`, все роуты под префиксом `/api/v1`.

```bash
BASE=http://localhost:8082/api/v1
```

## Аутентификация

Создание/изменение/удаление событий и изменение/удаление пользователей
требуют заголовок `Authorization: Bearer <token>`. Токен выдаётся при
логине. Middleware не просто проверяет подпись токена — она ещё раз идёт
в базу и проверяет, что пользователь из токена всё ещё существует, так
что токен удалённого пользователя перестаёт работать сразу же, не дожидаясь
истечения TTL.

```bash
curl -s -X POST "$BASE/register" \
  -H 'Content-Type: application/json' \
  -d '{"email": "alice@gmail.com", "password": "s3cret-pass"}'

TOKEN=$(curl -s -X POST "$BASE/login" \
  -H 'Content-Type: application/json' \
  -d '{"email": "alice@gmail.com", "password": "s3cret-pass"}' | jq -r .token)
echo "$TOKEN"
```

Без токена или с мусорным токеном — ожидаем `401`:

```bash
curl -i -X POST "$BASE/events" \
  -H 'Content-Type: application/json' \
  -d '{"name":"NoAuth","description":"x","location":"Remote","date_time":"2026-08-01T10:00:00Z","user_id":1}'

curl -i -X POST "$BASE/events" \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer garbage.token.here' \
  -d '{"name":"BadAuth","description":"x","location":"Remote","date_time":"2026-08-01T10:00:00Z","user_id":1}'
```

## Events

### Создать событие

Требует `Authorization: Bearer $TOKEN`. `user_id` в теле игнорируется —
владельцем всегда становится аутентифицированный пользователь из токена,
а не то, что прислал клиент.

```bash
curl -i -X POST "$BASE/events" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFsaWNlQGdtYWlsLmNvbSIsImV4cCI6MTc4NDA3NTI5MywidXNlcklEIjoxfQ.zKvUuBDFmow_t1vogOqxrAhJRZ4BdBacKY5EKJUWZ1Q" \
  -d '{
        "name": "Launch",
        "description": "Test event",
        "location": "Remote",
        "date_time": "2026-08-01T10:00:00Z"
      }'
```

### Список событий

Публичный, токен не нужен.

```bash
curl -i "$BASE/events"
```

### Событие по id

Публичный, токен не нужен.

```bash
curl -i "$BASE/events/1"
```

### Обновить событие

Требует `Authorization: Bearer $TOKEN` **и** что событие принадлежит этому
пользователю — иначе `403`. `user_id` в теле игнорируется, владелец через
update не меняется.

```bash
curl -i -X PUT "$BASE/events/1" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
        "name": "Launch v2",
        "description": "Updated",
        "location": "Remote",
        "date_time": "2026-08-02T10:00:00Z"
      }'
```

### Удалить событие

Требует `Authorization: Bearer $TOKEN` **и** владение событием — иначе `403`.

```bash
curl -i -X DELETE "$BASE/events/1" -H "Authorization: Bearer $TOKEN"
```

### Чужое событие — ожидаем `403`

`TOKEN_OTHER` — токен другого пользователя, не владеющего событием `1`.

```bash
curl -i -X PUT "$BASE/events/1" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN_OTHER" \
  -d '{"name": "Hijack", "description": "x", "location": "Remote", "date_time": "2026-08-01T10:00:00Z"}'

curl -i -X DELETE "$BASE/events/1" -H "Authorization: Bearer $TOKEN_OTHER"
```

## Users

### Регистрация

Email проверяется по DNS (MX/A запись) — домен должен реально существовать.
Публичный, токен не нужен (иначе им было бы невозможно воспользоваться).

```bash
curl -i -X POST "$BASE/register" \
  -H 'Content-Type: application/json' \
  -d '{"email": "alice@gmail.com", "password": "s3cret-pass"}'
```

Регистрация с несуществующим доменом — ожидаем `400`:

```bash
curl -i -X POST "$BASE/register" \
  -H 'Content-Type: application/json' \
  -d '{"email": "alice@no-such-domain-xyz123456789.com", "password": "s3cret-pass"}'
```

### Логин

Публичный. Возвращает `token`, который нужен для всех мутирующих запросов.

```bash
curl -i -X POST "$BASE/login" \
  -H 'Content-Type: application/json' \
  -d '{"email": "alice@gmail.com", "password": "s3cret-pass"}'
```

Неверный пароль — ожидаем `401`:

```bash
curl -i -X POST "$BASE/login" \
  -H 'Content-Type: application/json' \
  -d '{"email": "alice@gmail.com", "password": "wrong-pass"}'
```

### Пользователь по id

Публичный, токен не нужен.

```bash
curl -i "$BASE/users/1"
```

### Обновить пользователя

Требует `Authorization: Bearer $TOKEN`. `password` в теле обязателен по
валидации (`binding:"required"`), даже если сам update меняет только email.

```bash
curl -i -X PUT "$BASE/users/1" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"email": "alice2@gmail.com", "password": "s3cret-pass"}'
```

### Удалить пользователя

Требует `Authorization: Bearer $TOKEN`.

```bash
curl -i -X DELETE "$BASE/users/1" -H "Authorization: Bearer $TOKEN"
```

## Сквозной прогон одним куском

Регистрирует пользователя, логинится за токеном, создаёт событие, использует
id по цепочке, чистит за собой (событие удаляется раньше пользователя из-за
FOREIGN KEY на events.user_id). Требует `jq`.

```bash
BASE=http://localhost:8082/api/v1

USER_ID=$(curl -s -X POST "$BASE/register" \
  -H 'Content-Type: application/json' \
  -d '{"email": "bob@gmail.com", "password": "s3cret-pass"}' | jq -r .id)
echo "created user $USER_ID"

TOKEN=$(curl -s -X POST "$BASE/login" \
  -H 'Content-Type: application/json' \
  -d '{"email": "bob@gmail.com", "password": "s3cret-pass"}' | jq -r .token)

EVENT_ID=$(curl -s -X POST "$BASE/events" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Launch","description":"Test","location":"Remote","date_time":"2026-08-01T10:00:00Z"}' \
  | jq -r .id)
echo "created event $EVENT_ID"

curl -s "$BASE/events/$EVENT_ID"

curl -s -X DELETE "$BASE/events/$EVENT_ID" -H "Authorization: Bearer $TOKEN"
curl -s -X DELETE "$BASE/users/$USER_ID" -H "Authorization: Bearer $TOKEN"

echo "reusing token after user deletion (expect 401):"
curl -s -o /dev/null -w '%{http_code}\n' -X POST "$BASE/events" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"ShouldFail","description":"x","location":"Remote","date_time":"2026-08-01T10:00:00Z"}'
```
