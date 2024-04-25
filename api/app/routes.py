from functools import wraps
import json

import kafka.errors
from flask import Blueprint, jsonify, request
import jwt
import os
import kafka

from .models import db, User
from .services import create_user, update_user, authenticate_user

from google.protobuf.json_format import MessageToJson
import grpc

from .proto.task_service_pb2_grpc import TaskServiceStub
from .proto import task_service_pb2

from .proto.event_pb2 import Event, EventType


users_bp = Blueprint('users', __name__)
tasks_bp = Blueprint('tasks', __name__)

SECRET_KEY = os.getenv("SECRET_KEY")

kafka_producer = kafka.KafkaProducer(bootstrap_servers=[os.getenv("KAFKA_SERVICE")], value_serializer=lambda m: m.SerializeToString())

def grpc_connect():
    channel = grpc.insecure_channel(os.getenv("GRPC_SERVICE"))
    return TaskServiceStub(channel)


def token_required(f):
    @wraps(f)
    def decorated(*args, **kwargs):
        token = None
        if 'x-access-token' in request.headers:
            token = request.headers['x-access-token']
        if not token:
            return jsonify({'message' : 'Token is missing'}), 401

        try:
            data = jwt.decode(token, SECRET_KEY)
            user_id = db.session.query(User.id).filter_by(id=data['userID']).first()
            if not user_id:
                raise RuntimeError()
        except:
            return jsonify({
                'message' : 'Token is invalid'
            }), 401
        return  f(*args, **kwargs)
  
    return decorated


@users_bp.route('/register', methods=['POST'])
def register_user_route():
    try:
        data = request.get_json()
        if create_user(
            data['username'],
            data['password'],
            data['first_name'],
            data['last_name'],
            data['birth_date'],
            data['email'],
            data['phone_number']
        ):
            return jsonify({"message": "User registered successfully"}), 201
        return jsonify({"message": "User already exists"}), 401
    except:
        return jsonify({"message": "Bad request, missing or invalid parameters"}), 400


@users_bp.route('/update', methods=['PUT'])
@token_required
def update_user_route():
    try:
        data = request.get_json()
        if update_user(
            data['user_id'],
            data['first_name'],
            data['last_name'],
            data['birth_date'],
            data['email'],
            data['phone_number']
        ):
            return jsonify({"message": "User updated successfully"}), 200
        return jsonify({"message": "User not found"}), 404
    except:
        jsonify({"message": "Bad request, missing or invalid parameters"}), 400


@users_bp.route('/login', methods=['POST'])
def login_user_route():
    data = request.get_json()
    token = authenticate_user(data['username'], data['password'], SECRET_KEY)
    if token:
        resp = jsonify({"message": "User authenticated successfully"})
        resp.headers.add('x-access-token', token)
        return resp, 201
    else:
        return jsonify({"message": "Unauthorized, invalid credentials"}), 401


@tasks_bp.route('/create', methods=['POST'])
@token_required
def create_task():
    data = request.get_json()
    token = request.headers['x-access-token']

    try:
        client = grpc_connect()
        response = client.CreateTask(
            task_service_pb2.CreateTaskRequest(title=data.get('title'), content=data.get('content')),
            metadata=(('x-access-token', f'Bearer {token}'),)
        )
        return jsonify({
            'message': "Created task successfully",
            'id': response.id,
            'title': response.title,
            'content': response.content
        }), 201
    except grpc.RpcError as e:
        return jsonify({"message": f"rpc error: {e}"}), 500


@tasks_bp.route('/<task_id>', methods=['GET'])
@token_required
def select_task(task_id):
    token = request.headers['x-access-token']

    try:
        client = grpc_connect()
        response = client.GetTaskById(
            task_service_pb2.GetTaskByIdRequest(task_id=int(task_id)),
            metadata=(('x-access-token', f'Bearer {token}'),)
        )
        return jsonify({
            'id': response.id,
            'title': response.title,
            'content': response.content,
        }), 200
    except grpc.RpcError as e:
        return jsonify({"message": f"rpc error: {e}"}), 500


@tasks_bp.route('/<task_id>', methods=['PUT'])
@token_required
def update_task(task_id):
    data = request.get_json()
    token = request.headers['x-access-token']

    try:
        client = grpc_connect()
        response = client.UpdateTask(
            task_service_pb2.UpdateTaskRequest(task_id=int(task_id), title=data.get('title'), content=data.get('content')),
            metadata=(('x-access-token', f'Bearer {token}'),)
        )
        return jsonify({
            'id': response.id,
            'title': response.title,
            'content': response.content,
        }), 201
    except grpc.RpcError as e:
        return jsonify({"message": f"rpc error: {e}"}), 500


@tasks_bp.route('/<task_id>', methods=['DELETE'])
@token_required
def delete_task(task_id):
    token = request.headers['x-access-token']

    try:
        client = grpc_connect()
        response = client.DeleteTask(
            task_service_pb2.DeleteTaskRequest(task_id=int(task_id)),
            metadata=(('x-access-token', f'Bearer {token}'),)
        )
        if not response.success:
            return jsonify({
                'message': "Undefined error"
            }), 401
        return jsonify({
            'id': task_id,
        }), 201
    except grpc.RpcError as e:
        return jsonify({"message": f"rpc error: {e}"}), 500


@tasks_bp.route('/page/<page_number>/<page_size>', methods=['GET'])
@token_required
def get_pag(page_number, page_size):
    token = request.headers['x-access-token']

    try:
        client = grpc_connect()
        response = client.GetTaskListWithPagination(
            task_service_pb2.GetTaskListRequest(page_number=int(page_number), page_size=int(page_size)),
            metadata=(('x-access-token', f'Bearer {token}'),)
        )
        return MessageToJson(response), 200
    except grpc.RpcError as e:
        return jsonify({"message": f"rpc error: {e}"}), 500


@tasks_bp.route('/<task_id>/view', methods=['POST'])
@token_required
def send_view(task_id):
    try:
        kafka_producer.send('event-topic', Event(task_id=int(task_id), event_type=EventType.VIEW))
        return jsonify({
            'message': 'view sent successfully'
        }), 201
    except Exception as e:
        return jsonify({"message": f"send to topic error: {e}"}), 500


@tasks_bp.route('/<task_id>/like', methods=['POST'])
@token_required
def send_like(task_id):
    try:
        kafka_producer.send('event-topic', Event(task_id=int(task_id), event_type=EventType.LIKE))
        return jsonify({
            'message': 'like sent successfully'
        }), 201
    except Exception as e:
        return jsonify({"message": f"send to topic error: {e}"}), 500
