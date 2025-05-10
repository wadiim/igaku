# Igaku

## Quick Start

```console
$ ln -s .env.example .env
$ docker compose up
```

Visit:
- <http://localhost:8080/swagger/index.html>
- <http://localhost:8081/swagger/index.html>
- <http://localhost:8082/swagger/index.html>

## Testing

```console
$ go -C user-service test ./tests/
$ go -C auth-service test ./tests/
$ go -C encounter-service test ./tests/
$ go -C user-service test -tags=integration ./tests/ -v
$ go -C auth-service test -tags=integration ./tests/ -v
```

## Swagger documentation generation

```console
$ swag init -g main.go -o user-service/docs -d user-service,commons/dtos,commons/errors,commons/models,commons/utils
$ swag init -g main.go -o auth-service/docs -d auth-service,commons/dtos,commons/errors,commons/models,commons/utils
$ swag init -g main.go -o encounter-service/docs -d encounter-service,commons/dtos,commons/errors,commons/models,commons/utils
```
