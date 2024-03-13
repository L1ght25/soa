FROM python:3.8

ENV PYTHONUNBUFFERED 1

COPY requirements.txt /app/

RUN pip install --no-cache-dir -r /app/requirements.txt

COPY . /app/

WORKDIR /app

CMD ["python3", "run.py"]
