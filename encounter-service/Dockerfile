FROM golang:alpine
RUN apk add --no-cache git
WORKDIR /src
COPY ./encounter-service ./encounter-service
COPY ./commons ./commons
RUN go build -C ./encounter-service
RUN chmod +x ./encounter-service/encounter-service
CMD ["./encounter-service/encounter-service"]
EXPOSE 8080
