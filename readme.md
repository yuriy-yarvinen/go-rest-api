# Gin install

go get github.com/gin-gonic/gin

go get github.com/mattn/go-sqlite3

go get github.com/lib/pq

go get golang.org/x/crypto

## Database

The database driver is selected with the `DB_DRIVER` env var: `sqlite` (default) or `postgres`.

- sqlite: `DB_PATH` (default `./api.db`)
- postgres: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` (defaults: `localhost`, `5432`, `postgres`, `postgres`, `go_rest_api`, `disable`)

For local (non-Docker) development, copy `.env.example` to `.env` and source it:

```sh
cp .env.example .env
set -a && source .env && set +a
go run main.go
```

## Тесты

```sh
go test ./events/... ./users/... ./utils/...
```

Не гоняй `go test ./...` целиком: `go build`/`go test` тогда пытается
пройти и `volumes/postgres/`, а там после `docker-compose up` лежат файлы
Postgres (сокеты и т.п.), созданные от имени контейнера — обход упадёт с
`permission denied`. Пакеты с `postgres`/`sqlite`/`redis` в названии тестов
не содержат (нужна реальная БД/Redis) — покрыты только `events`, `users` и
`utils`.

Флаги, которые полезны по ходу:

```sh
go test ./events/... ./users/... ./utils/... -v          # подробный вывод по каждому тесту
go test ./events/... ./users/... ./utils/... -run JWT     # только тесты, чьё имя матчит regexp
go test ./events ./users ./users/rest ./utils -cover      # % покрытия (без /... — иначе тулчейн
                                                            # падает с "no such tool covdata" на
                                                            # cgo-пакетах без тестов, postgres/sqlite)
```

## Docker

Run the API with a Postgres database in containers:

```sh
docker compose up --build
```

The API is served on `http://localhost:8082`.

Project layout for the container setup:

- `Dockerfile` — multi-stage build for the app binary.
- `docker-compose.yml` — app + Postgres services.
- `.env` — env vars injected into the containers.
- `volumes/postgres/` — bind-mounted Postgres data directory (gitignored).
