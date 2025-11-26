FROM golang:alpine
RUN apk add --no-cache git
WORKDIR /src
COPY ./visit-service ./visit-service
COPY ./commons ./commons
RUN go build -C ./visit-service
RUN chmod +x ./visit-service/visit-service
CMD ["./visit-service/visit-service"]
EXPOSE 8080
