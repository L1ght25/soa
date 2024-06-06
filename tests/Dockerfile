FROM python:3.8

ENV PYTHONUNBUFFERED 1

COPY /tests /tests
COPY /proto /proto

RUN pip install -r /tests/requirements.txt

RUN /tests/proto/generate.sh

CMD ["python3", "/tests/kafka_test.py"]
