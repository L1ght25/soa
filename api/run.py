from flask import Flask
from app.models import db
from app.routes import users_bp, tasks_bp
from connexion import App
import os


if __name__ == '__main__':

    # Инициализируем приложение Connexion
    connexion_app = App(__name__, specification_dir='./')
    connexion_app.add_api('openapi.yml')

    app = connexion_app.app

    # Регистрируем спецификацию OpenAPI
    app.config['SWAGGER_UI_JSONEDITOR'] = True
    app.config['SWAGGER_JSON'] = 'openapi.yml'

    # Подключаемся к базе данных
    app.config['SQLALCHEMY_DATABASE_URI'] = os.environ.get('DATABASE_URI')
    app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False

    db.init_app(app)

    with app.app_context():
        db.create_all()

    # Регистрируем Blueprint с роутами пользователей
    app.register_blueprint(users_bp, url_prefix='/users')

    # Регистрируем Blueprint с роутами тасок
    app.register_blueprint(tasks_bp, url_prefix='/tasks')

    app.run(debug=True, host="0.0.0.0", port="5100")
