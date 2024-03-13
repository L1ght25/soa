import requests

BASE_URL = 'http://localhost:5100/users'

# Регистрация нового пользователя
def register_user(username, password, first_name, last_name, birth_date, email, phone_number):
    url = f'{BASE_URL}/register'
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
    url = f'{BASE_URL}/login'
    data = {
        'username': username,
        'password': password
    }
    response = requests.post(url, json=data)
    print(response.json(), response.status_code)

# Обновление данных пользователя
def update_user(user_id, first_name, last_name, birth_date, email, phone_number):
    url = f'{BASE_URL}/update'
    data = {
        'user_id': user_id,
        'first_name': first_name,
        'last_name': last_name,
        'birth_date': birth_date,
        'email': email,
        'phone_number': phone_number
    }
    response = requests.put(url, json=data)
    print(response.json(), response.status_code)

if __name__ == '__main__':
    register_user('testuser', 'password123', 'John', 'Doe', '2000-01-01', 'test@example.com', '1234567890')
    login_user('testuser', 'password123')
    update_user(1, 'John', 'Doe', '2000-01-01', 'test@example.com', '1234567890')
