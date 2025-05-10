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
$ go -C api-service test ./tests/
$ go -C auth-service test ./tests/
$ go -C api-service test -tags=integration ./tests/ -v
$ go -C auth-service test -tags=integration ./tests/ -v
```

## Swagger documentation generation

```console
$ swag init -g main.go -o api-service/docs -d api-service,commons/dtos,commons/errors,commons/models,commons/utils
$ swag init -g main.go -o auth-service/docs -d auth-service,commons/dtos,commons/errors,commons/models,commons/utils
```
