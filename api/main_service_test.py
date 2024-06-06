# test_app.py
import unittest
from unittest.mock import patch

from flask import Flask
from app.routes import users_bp, tasks_bp, stats_bp


app = Flask(__name__)
# Регистрируем Blueprint с роутами пользователей
app.register_blueprint(users_bp, url_prefix='/users')

# Регистрируем Blueprint с роутами тасок
app.register_blueprint(tasks_bp, url_prefix='/tasks')

# Регистрируем Blueprint с роутами статистики
app.register_blueprint(stats_bp, url_prefix='/stats')


class TaskServiceTests(unittest.TestCase):
    def setUp(self):
        self.app = app.test_client()
        self.app.testing = True

    @patch('app.services.create_user')
    def test_register_user_success(self, mock_create_user):
        mock_create_user.return_value = True
        response = self.app.post('/users/register', json={
            'username': 'testuser',
            'password': 'password123',
            'first_name': 'Test',
            'last_name': 'User',
            'birth_date': '1990-01-01',
            'email': 'testuser@example.com',
            'phone_number': '+1234567890'
        })
        self.assertEqual(response.status_code, 201)
        self.assertEqual(response.json, {'message': 'User registered successfully'})
        mock_create_user.assert_called_once()

    @patch('app.services.create_user')
    def test_register_user_missing_field(self, mock_create_user):
        mock_create_user.return_value = True
        response = self.app.post('/users/register', json={
            'username': 'testuser',
            'password': 'password123',
            'first_name': 'Test',
            'last_name': 'User',
            'email': 'testuser@example.com',
            'phone_number': '+1234567890'
        })
        self.assertEqual(response.status_code, 400)
        self.assertIn('Bad request, missing or invalid parameters', response.json['message'])
        mock_create_user.assert_not_called()

    @patch('app.services.create_user')
    def test_register_user_already_exists(self, mock_create_user):
        mock_create_user.return_value = False
        response = self.app.post('/users/register', json={
            'username': 'testuser',
            'password': 'password123',
            'first_name': 'Test',
            'last_name': 'User',
            'birth_date': '1990-01-01',
            'email': 'testuser@example.com',
            'phone_number': '+1234567890'
        })
        self.assertEqual(response.status_code, 401)
        self.assertEqual(response.json, {'message': 'User already exists'})


if __name__ == '__main__':
    unittest.main()
