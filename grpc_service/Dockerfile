FROM golang:latest

COPY /grpc_service /grpc_service
COPY /proto /proto

WORKDIR /grpc_service


RUN apt-get update && apt-get install --no-install-recommends --assume-yes protobuf-compiler

RUN go mod download && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    ./proto/generate.sh


RUN go build -o taskservice .

EXPOSE 50051

CMD ["./taskservice"]
