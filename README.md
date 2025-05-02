# Igaku

## Quick Start

```console
$ ln -s .env.example .env
$ docker compose up
```

Visit:
- <http://localhost:8080/swagger/index.html>
- <http://localhost:8081/swagger/index.html>

## Testing

```console
$ cd api
$ go test ./api-service/tests/
$ go test ./auth-service/tests/
$ go test -tags=integration ./api-service/tests/ -v
$ go test -tags=integration ./auth-service/tests/ -v
```

## Swagger documentation generation

```console
cd api-service
swag init -d .,../commons/dtos/,../commons/errors/,../commons/models/,../commons/utils/
cd ../auth-service
swag init -d .,../commons/dtos/,../commons/errors/,../commons/models/,../commons/utils/
```
