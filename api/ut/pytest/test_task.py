import requests
import pytest

@pytest.fixture
def url():
    return "http://localhost:5100"

def test_task_service(url):
    url_reg = f'{url}/users/register'
    requests.post(url_reg, json={
        'username': "first_user",
        'password': "123",
        'first_name': "Hello",
        'last_name': "World",
        'birth_date': "12.12.1212",
        'email': "efe@fefe.ee",
        'phone_number': "123456789"
    })
    requests.post(url_reg, json={
        'username': "second_user",
        'password': "123",
        'first_name': "Hello1",
        'last_name': "World1",
        'birth_date': "12.12.1212",
        'email': "efefefe@feffefe.ee",
        'phone_number': "123456789"
    })

    token1 = requests.post(f'{url}/users/login', json={
        'username': 'first_user',
        'password': '123'
    }).headers.get('x-access-token')
    token2 = requests.post(f'{url}/users/login', json={
        'username': 'second_user',
        'password': '123'
    }).headers.get('x-access-token')

    task_id_first = requests.post(f'{url}/tasks/create', json={
        'title': "first",
        'content': "FIRST CONTENT",
    }, headers={'x-access-token': token1}).json().get('id')

    task_id_second = requests.post(f'{url}/tasks/create', json={
        'title': "second",
        'content': "SECOND CONTENT",
    }, headers={'x-access-token': token2}).json().get('id')

    tasks = requests.get(f'{url}/tasks/page/0/2', headers={'x-access-token': token1}).json().get('tasks')

    assert len(tasks) == 2
    assert tasks[0]['id'] == 1
    assert tasks[0]['title'] == "first"
    assert tasks[0]['content'] == "FIRST CONTENT"

    assert tasks[1]['id'] == 2
    assert tasks[1]['title'] == "second"
    assert tasks[1]['content'] == "SECOND CONTENT"