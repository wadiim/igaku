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

### Unit Testing

```console
$ go -C user-service test ./tests/
$ go -C auth-service test ./tests/
$ go -C encounter-service test ./tests/
```

### Integration Testing

```console
$ go -C user-service test -tags=integration ./tests/ -v
$ go -C auth-service test -tags=integration ./tests/ -v
```

### Load Testing

```console
$ ./tools/run-load-tests.sh
```

## Swagger documentation generation

```console
$ ./tools/gen-docs.sh
```
