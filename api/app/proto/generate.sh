#!/bin/bash
# needs to be run in root folder
python3 -m grpc_tools.protoc -I/proto --python_out=/api/app/proto/ --pyi_out=/api/app/proto/ --grpc_python_out=/api/app/proto/ /proto/task_service.proto
