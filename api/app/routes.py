from functools import wraps
from flask import Blueprint, jsonify, request
import jwt
import os

from .models import db, User
from .services import create_user, update_user, authenticate_user

users_bp = Blueprint('users', __name__)

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
            user_id = db.session.query(User.id).filter_by(username=data['username']).first()
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
