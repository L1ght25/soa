FROM golang:latest

COPY /statistics_service /statistics_service
COPY /proto /proto
COPY /common /common


WORKDIR /statistics_service

RUN apt-get update && apt-get install --no-install-recommends --assume-yes protobuf-compiler

RUN go mod download && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    ./proto/generate.sh

RUN go build -o statistics_service .

EXPOSE 50052

CMD ["./statistics_service"]
