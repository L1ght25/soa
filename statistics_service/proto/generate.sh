#!/bin/bash
# needs to be run in root folder folder
task_prefix=""
if [[ "$1" == "--local" ]]; then
    task_prefix="."
fi

protoc --proto_path=${task_prefix}/proto --go_out=${task_prefix}/statistics_service/ ${task_prefix}/proto/event.proto
protoc --proto_path=${task_prefix}/proto --go_out=${task_prefix}/statistics_service/ --go-grpc_out=${task_prefix}/statistics_service/ ${task_prefix}/proto/statistics_service.proto