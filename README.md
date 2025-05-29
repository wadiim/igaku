# Igaku

## Quick Start

```console
$ ln -s .env.example .env
```

> [!WARNING]
> Most of the values in the `.env` file can remain unchanged but the SMTP
> (Simple Mail Transfer Protocol) credentials must be set.

```console
$ docker compose up
```

Visit:
- [Swagger UI](http://localhost:8090/)
- [RabbitMQ Management UI](http://localhost:15672)

## Service Replication

```console
$ docker compose up --scale auth=2 --scale user=2 --scale mail=2
```

## Starting the [ELK](https://www.elastic.co/elastic-stack/) stack

```console
$ docker compose --profile elk up
```

Visit:
- [Kibana](http://localhost:5601)

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
$ swag init -g main.go -o mail-service/docs -d mail-service,commons/dtos,commons/errors,commons/models,commons/utils
```
