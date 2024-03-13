# Используем базовый образ Python
FROM python:3.8

# Устанавливаем переменную окружения PYTHONUNBUFFERED
# Это предотвращает вывод Python в буфер и обеспечивает вывод в реальном времени
ENV PYTHONUNBUFFERED 1

# Копируем файлы зависимостей
COPY requirements.txt /app/

# Устанавливаем зависимости
RUN pip install --no-cache-dir -r /app/requirements.txt

# Копируем файлы приложения в контейнер
COPY . /app/

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Команда для запуска приложения
CMD ["python3", "run.py"]
