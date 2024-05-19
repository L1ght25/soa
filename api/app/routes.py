from functools import wraps

from flask import Blueprint, jsonify, request, g
import jwt
import os

from .models import db, User
from .proto.event_pb2 import EventType
from . import services

from google.protobuf.json_format import MessageToJson
import grpc


users_bp = Blueprint('users', __name__)
tasks_bp = Blueprint('tasks', __name__)

SECRET_KEY = os.getenv("SECRET_KEY")


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
            g.user_id = data['userID']
        except:
            return jsonify({
                'message' : 'Token is invalid'
            }), 401

        return f(*args, **kwargs)
  
    return decorated


@users_bp.route('/register', methods=['POST'])
def register_user_route():
    try:
        data = request.get_json()
        if services.create_user(
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
        if services.update_user(
            g.get('user_id'),
            data['first_name'],
            data['last_name'],
            data['birth_date'],
            data['email'],
            data['phone_number']
        ):
            return jsonify({"message": "User updated successfully"}), 200
        return jsonify({"message": "User not found"}), 404
    except:
        return jsonify({"message": "Bad request, missing or invalid parameters"}), 400


@users_bp.route('/login', methods=['POST'])
def login_user_route():
    data = request.get_json()
    token = services.authenticate_user(data['username'], data['password'], SECRET_KEY)
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
        response = services.create_task(title=data.get('title'), content=data.get('content'), token=token)
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
        response = services.get_task(task_id=task_id, token=token)
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
        response = services.update_task(
            task_id=task_id,
            title=data.get('title'),
            content=data.get('content'),
            token=token
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
        response = services.delete_task(task_id=task_id, token=token)
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
        response = services.get_pag(page_number=page_number, page_size=page_size, token=token)
        return MessageToJson(response), 200
    except grpc.RpcError as e:
        return jsonify({"message": f"rpc error: {e}"}), 500


@tasks_bp.route('/<task_id>/view', methods=['POST'])
@token_required
def send_view(task_id):
    token = request.headers['x-access-token']

    try:
        services.send_event(task_id, EventType.VIEW, token, g.get('user_id'))
        return jsonify({
            'message': 'view sent successfully'
        }), 201
    except Exception as e:
        return jsonify({"message": f"send to topic error: {e}"}), 500


@tasks_bp.route('/<task_id>/like', methods=['POST'])
@token_required
def send_like(task_id):
    token = request.headers['x-access-token']

    try:
        services.send_event(task_id, EventType.LIKE, token, g.get('user_id'))
        return jsonify({
            'message': 'like sent successfully'
        }), 201
    except Exception as e:
        return jsonify({"message": f"send to topic error: {e}"}), 500
