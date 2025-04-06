# Igaku

## Quick Start

```console
$ ln -s .env.example .env
$ docker compose up
```

```console
$ curl http://localhost:8080/hello
```

## Testing

```console
$ cd api
$ go test ./test/
$ go test -tags=integration ./test/ -v
```
