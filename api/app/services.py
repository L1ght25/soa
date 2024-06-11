import datetime
import os

import grpc
import kafka
import jwt
from werkzeug.security import generate_password_hash, check_password_hash
from .models import db, User

from .proto.task_service_pb2_grpc import TaskServiceStub
from .proto import task_service_pb2

from .proto.statistics_service_pb2_grpc import StatsServiceStub
from .proto import statistics_service_pb2
from .proto.event_pb2 import EventType

from .proto.event_pb2 import Event


kafka_producer = kafka.KafkaProducer(bootstrap_servers=[os.getenv("KAFKA_SERVICE")], value_serializer=lambda m: m.SerializeToString())

def tasks_service_connect():
    channel = grpc.insecure_channel(os.getenv("GRPC_SERVICE"))
    return TaskServiceStub(channel)


def stats_service_connect():
    channel = grpc.insecure_channel(os.getenv("STATS_SERVICE"))
    return StatsServiceStub(channel)


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
            'userID': user.id,
            'exp' : datetime.datetime.now() + datetime.timedelta(minutes = 30)
        }, secret_key)
        return token.decode('UTF-8')
    return False


def create_task(title, content, token):
    client = tasks_service_connect()
    return client.CreateTask(
        task_service_pb2.CreateTaskRequest(title=title, content=content),
        metadata=(('x-access-token', f'Bearer {token}'),)
    )


def get_task(task_id, token):
    client = tasks_service_connect()
    return client.GetTaskById(
        task_service_pb2.GetTaskByIdRequest(task_id=int(task_id)),
        metadata=(('x-access-token', f'Bearer {token}'),)
    )


def update_task(task_id, title, content, token):
    client = tasks_service_connect()
    return client.UpdateTask(
        task_service_pb2.UpdateTaskRequest(task_id=int(task_id), title=title, content=content),
        metadata=(('x-access-token', f'Bearer {token}'),)
    )


def delete_task(task_id, token):
    client = tasks_service_connect()
    return client.DeleteTask(
        task_service_pb2.DeleteTaskRequest(task_id=int(task_id)),
        metadata=(('x-access-token', f'Bearer {token}'),)
    )


def get_pag(page_number, page_size, token):
    client = tasks_service_connect()
    return client.GetTaskListWithPagination(
        task_service_pb2.GetTaskListRequest(page_number=int(page_number), page_size=int(page_size)),
        metadata=(('x-access-token', f'Bearer {token}'),)
    )


def send_event(task_id, event_type, token, user_id):
    task_info = get_task(task_id, token)
    kafka_producer.send('event-topic', Event(
        task_id=int(task_id),
        event_type=event_type,
        author_id=int(task_info.createdByUserID),
        user_id=int(user_id)
    ))


def get_task_stats(task_id, token):
    client = stats_service_connect()
    return client.GetTaskStats(
        statistics_service_pb2.TaskRequest(task_id=int(task_id)),
        metadata=(('x-access-token', f'Bearer {token}'),)
    )


def get_top_tasks(sort_by, token):
    client = stats_service_connect()
    return client.GetTopTasks(
        statistics_service_pb2.TopTasksRequest(metric=EventType.LIKE if sort_by == 'likes' else EventType.VIEW),
        metadata=(('x-access-token', f'Bearer {token}'),)
    )


def get_top_users(token):
    client = stats_service_connect()
    return client.GetTopUsers(
        statistics_service_pb2.TopUsersRequest(),
        metadata=(('x-access-token', f'Bearer {token}'),)
    )


def get_username_by_id(id):
    return db.session.query(User.username).filter_by(id=id).first()[0]
