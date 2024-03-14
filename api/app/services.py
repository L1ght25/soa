import datetime
import jwt
from werkzeug.security import generate_password_hash, check_password_hash
from .models import db, User

def create_user(username, password, first_name, last_name, birth_date, email, phone_number):
    if db.session.query(User.id).filter_by(username=username).first() is not None:
        return False

    hashed_password = generate_password_hash(password, method='sha256')
    new_user = User(
        username=username,
        password=hashed_password,
        first_name=first_name,
        last_name=last_name,
        birth_date=birth_date,
        email=email,
        phone_number=phone_number
    )
    db.session.add(new_user)
    db.session.commit()
    return True


def update_user(user_id, first_name, last_name, birth_date, email, phone_number):
    user = User.query.get(user_id)
    if user:
        user.first_name = first_name
        user.last_name = last_name
        user.birth_date = birth_date
        user.email = email
        user.phone_number = phone_number
        db.session.commit()
        return True
    return False


def authenticate_user(username, password, secret_key):
    user = User.query.filter_by(username=username).first()
    if user and check_password_hash(user.password, password):
        token = jwt.encode({
            'username': user.username,
            'exp' : datetime.datetime.now() + datetime.timedelta(minutes = 30)
        }, secret_key)
        return token.decode('UTF-8')
    return False
