#!/bin/bash
# needs to be run in root folder folder
task_prefix=""
if [[ "$1" == "--local" ]]; then
    task_prefix="."
fi

protoc --proto_path=${task_prefix}/proto --go_out=${task_prefix}/grpc_service/ --go-grpc_out=${task_prefix}/grpc_service/ ${task_prefix}/proto/task_service.proto