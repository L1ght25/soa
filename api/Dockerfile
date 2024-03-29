FROM python:3.8

ENV PYTHONUNBUFFERED 1

COPY /api /api
COPY /proto /proto

RUN pip install -r /api/requirements.txt

RUN /api/app/proto/generate.sh

CMD ["python3", "/api/run.py"]
