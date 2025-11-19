# Igaku

<p align="center">
  <img src="https://github.com/user-attachments/assets/1859108e-11f0-4f9b-a0a4-0331697153eb" style="width: 32%; height: auto;" />
</p>

## Quick Start

```console
$ ln -s .env.example .env
$ docker compose up
```

Visit:
- [Swagger UI](http://localhost:8090/)
- [RabbitMQ Management UI](http://localhost:15672)

## Service Replication

```console
$ docker compose up --scale auth=2 --scale user=2
```

## Enabling the `mail` service

By default, the `mail` service is disabled in order to minimize the number of
steps needed to run the project for the first time. Enabling the `mail`
service requires setting the `MAIL_ENABLED` variable in the `.env` file to
a non-empty value. Additionally, the `SMTP_*` variables must be set to
values appropriate to the SMTP (Simple Mail Transfer Protocol) server being
used.

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

### Client code type checking

```console
$ cd client
$ npx tsc --project tsconfig.app.json --noEmit --watch
```

## Swagger documentation generation

```console
$ ./tools/gen-docs.sh
```

## License

[MIT](https://github.com/wadiim/igaku/blob/main/LICENSE)
