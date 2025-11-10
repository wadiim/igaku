FROM golang:alpine
RUN apk add --no-cache git
WORKDIR /src
COPY ./mail-service ./mail-service
COPY ./commons ./commons
RUN go build -C ./mail-service
RUN chmod +x ./mail-service/mail-service
CMD ["./mail-service/mail-service"]
EXPOSE 8080
