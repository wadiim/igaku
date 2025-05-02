FROM golang:latest

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download
COPY ./api-service ./api-service
COPY ./auth-service ./auth-service
COPY ./commons ./commons
RUN go build -buildvcs=false -o ./bin/api ./api-service

EXPOSE 8080
ENTRYPOINT ["./bin/api"]
