#!/bin/bash
# needs to be run in root folder

task_prefix=""
if [[ "$1" == "--local" ]]; then
    task_prefix="."
fi
python3 -m grpc_tools.protoc -I${task_prefix}/proto --python_out=${task_prefix}/api/app/proto/ --pyi_out=${task_prefix}/api/app/proto/ --grpc_python_out=${task_prefix}/api/app/proto/ ${task_prefix}/proto/task_service.proto
python3 -m grpc_tools.protoc -I${task_prefix}/proto --python_out=${task_prefix}/api/app/proto/ --pyi_out=${task_prefix}/api/app/proto/ ${task_prefix}/proto/event.proto
