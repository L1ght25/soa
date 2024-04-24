#!/bin/bash
# needs to be run in root folder folder
protoc --proto_path=/proto --go_out=/grpc_service/proto --go-grpc_out=/grpc_service/proto /proto/task_service.proto