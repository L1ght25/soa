import requests

BASE_URL = 'http://localhost:5100'

# Регистрация нового пользователя
def register_user(username, password, first_name, last_name, birth_date, email, phone_number):
    url = f'{BASE_URL}/users/register'
    data = {
        'username': username,
        'password': password,
        'first_name': first_name,
        'last_name': last_name,
        'birth_date': birth_date,
        'email': email,
        'phone_number': phone_number
    }
    response = requests.post(url, json=data)
    print(response.json(), response.status_code)

# Авторизация пользователя
def login_user(username, password):
    url = f'{BASE_URL}/users/login'
    data = {
        'username': username,
        'password': password
    }
    response = requests.post(url, json=data)
    print(response.json(), response.status_code, response.headers)
    return response.headers.get('x-access-token')

# Обновление данных пользователя
def update_user(token, first_name, last_name, birth_date, email, phone_number):
    url = f'{BASE_URL}/users/update'
    data = {
        'first_name': first_name,
        'last_name': last_name,
        'birth_date': birth_date,
        'email': email,
        'phone_number': phone_number
    }
    response = requests.put(url, json=data, headers={'x-access-token': token})
    print(response.json(), response.status_code)


def create_task(token, title, content):
    url = f'{BASE_URL}/tasks/create'
    data = {
        'title': title,
        'content': content,
    }
    response = requests.post(url, json=data, headers={'x-access-token': token})
    print(response.json(), response.status_code)

    return response.json().get('id')


def get_task(token, task_id):
    url = f'{BASE_URL}/tasks/{task_id}'
    response = requests.get(url, headers={'x-access-token': token})
    print(response.json(), response.status_code)


def get_pag(token, page_number, page_size):
    url = f'{BASE_URL}/tasks/page/{page_number}/{page_size}'
    response = requests.get(url, headers={'x-access-token': token})
    print(response.json(), response.status_code)


def send_view(token, task_id):
    url = f'{BASE_URL}/tasks/{task_id}/view'
    response = requests.post(url, headers={'x-access-token': token})
    print(response.json(), response.status_code)


def send_like(token, task_id):
    url = f'{BASE_URL}/tasks/{task_id}/like'
    response = requests.post(url, headers={'x-access-token': token})
    print(response.json(), response.status_code)


def get_stats(token, task_id):
    url = f'{BASE_URL}/stats/tasks/{task_id}'
    response = requests.get(url, headers={'x-access-token': token})
    print(response.json(), response.status_code)


def get_top_tasks(token):
    url = f'{BASE_URL}/stats/top-tasks'
    response = requests.get(url, headers={'x-access-token': token}, params={'sort_by': 'likes'})
    print('by likes:', response.json(), response.status_code)

    response = requests.get(url, headers={'x-access-token': token}, params={'sort_by': 'views'})
    print('by views:', response.json(), response.status_code)


def get_top_users(token):
    url = f'{BASE_URL}/stats/top-users'
    response = requests.get(url, headers={'x-access-token': token})
    print(response.json(), response.status_code)


if __name__ == '__main__':
    register_user('testuser', 'password123', 'John', 'Doe', '2000-01-01', 'test@example.com', '1234567890')
    token = login_user('testuser', 'password123')
    update_user(token, 'John', 'Doe', '2000-01-01', 'test@example.com', '1234567890')

    task_id = create_task(token, "test task", "this is test task")
    get_task(token, task_id)
    get_task("Test with invalid token", task_id)

    get_pag(token, 0, 5)

    send_view(token, 1)
    send_like(token, 1)

    get_stats(token, 1)
    get_top_tasks(token)
    get_top_users(token)

    register_user('testuser2', 'password123', 'John', 'Doe', '2000-01-01', 'testwefweafwe@example.com', '1234567890')
    token = login_user('testuser2', 'password123')
    send_view(token, 1)
    send_like(token, 1)

    register_user('testuser3', 'password123', 'John', 'Doe', '2000-01-01', 'testwefweafefwefwe@example.com', '1234567890')
    token = login_user('testuser3', 'password123')
    send_view(token, 1)
    send_like(token, 1)
