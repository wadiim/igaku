# Igaku

## Quick Start

```console
$ ln -s .env.example .env
$ docker compose up
```

Visit <http://localhost:8080/swagger/index.html>

## Testing

```console
$ cd api
$ go test ./test/
$ go test -tags=integration ./test/ -v
```
