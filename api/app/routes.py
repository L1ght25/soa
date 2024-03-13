from flask import Blueprint, jsonify, request
from .models import db, User
from .services import create_user, update_user, authenticate_user

users_bp = Blueprint('users', __name__)

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
    if authenticate_user(data['username'], data['password']):
        return jsonify({"message": "User authenticated successfully"}), 200
    else:
        return jsonify({"message": "Unauthorized, invalid credentials"}), 401
