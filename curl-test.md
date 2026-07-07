# curl тесты

Все команды проверены вручную против запущенного стека (`docker-compose up`).
Сервер слушает `localhost:8082`, все роуты под префиксом `/api/v1`.

```bash
BASE=http://localhost:8082/api/v1
```

## Events

### Создать событие

```bash
curl -i -X POST "$BASE/events" \
  -H 'Content-Type: application/json' \
  -d '{
        "name": "Launch",
        "description": "Test event",
        "location": "Remote",
        "date_time": "2026-08-01T10:00:00Z",
        "user_id": 1
      }'
```

### Список событий

```bash
curl -i "$BASE/events"
```

### Событие по id

```bash
curl -i "$BASE/events/1"
```

### Обновить событие

```bash
curl -i -X PUT "$BASE/events/1" \
  -H 'Content-Type: application/json' \
  -d '{
        "name": "Launch v2",
        "description": "Updated",
        "location": "Remote",
        "date_time": "2026-08-02T10:00:00Z",
        "user_id": 1
      }'
```

### Удалить событие

```bash
curl -i -X DELETE "$BASE/events/1"
```

## Users

### Регистрация

Email проверяется по DNS (MX/A запись) — домен должен реально существовать.

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

```bash
curl -i "$BASE/users/1"
```

### Обновить пользователя

`password` в теле обязателен по валидации (`binding:"required"`), даже
если сам update меняет только email.

```bash
curl -i -X PUT "$BASE/users/1" \
  -H 'Content-Type: application/json' \
  -d '{"email": "alice2@gmail.com", "password": "s3cret-pass"}'
```

### Удалить пользователя

```bash
curl -i -X DELETE "$BASE/users/1"
```

## Сквозной прогон одним куском

Создаёт пользователя и событие, использует их id по цепочке, чистит за собой.
Требует `jq`.

```bash
BASE=http://localhost:8082/api/v1

USER_ID=$(curl -s -X POST "$BASE/register" \
  -H 'Content-Type: application/json' \
  -d '{"email": "bob@gmail.com", "password": "s3cret-pass"}' | jq -r .id)
echo "created user $USER_ID"

curl -s -X POST "$BASE/login" \
  -H 'Content-Type: application/json' \
  -d '{"email": "bob@gmail.com", "password": "s3cret-pass"}'

EVENT_ID=$(curl -s -X POST "$BASE/events" \
  -H 'Content-Type: application/json' \
  -d "{\"name\":\"Launch\",\"description\":\"Test\",\"location\":\"Remote\",\"date_time\":\"2026-08-01T10:00:00Z\",\"user_id\":$USER_ID}" \
  | jq -r .id)
echo "created event $EVENT_ID"

curl -s "$BASE/events/$EVENT_ID"
curl -s -X DELETE "$BASE/events/$EVENT_ID"
curl -s -X DELETE "$BASE/users/$USER_ID"
```
