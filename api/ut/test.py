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
def update_user(token, user_id, first_name, last_name, birth_date, email, phone_number):
    url = f'{BASE_URL}/users/update'
    data = {
        'user_id': user_id,
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
    url = f'{BASE_URL}/tasks/page'
    response = requests.get(url, json={'page_number': page_number, 'page_size': page_size}, headers={'x-access-token': token})
    print(response.json(), response.status_code)


if __name__ == '__main__':
    register_user('testuser', 'password123', 'John', 'Doe', '2000-01-01', 'test@example.com', '1234567890')
    token = login_user('testuser', 'password123')
    update_user(token, 1, 'John', 'Doe', '2000-01-01', 'test@example.com', '1234567890')

    task_id = create_task(token, "test task", "this is test task")
    get_task(token, task_id)
    get_task("Test with invalid token", task_id)

    get_pag(token, 1, 5)
