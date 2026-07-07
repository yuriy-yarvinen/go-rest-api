FROM golang:1.25-alpine AS builder

# go-sqlite3 needs cgo + a C toolchain to build.
RUN apk add --no-cache gcc musl-dev

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /out/server .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/server ./server

EXPOSE 8082
ENTRYPOINT ["./server"]
