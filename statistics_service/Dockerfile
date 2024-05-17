FROM golang:latest

COPY /statistics_service /statistics_service


WORKDIR /statistics_service

RUN go build -o statistics_service .

EXPOSE 8080

CMD ["./statistics_service"]
