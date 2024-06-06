import os
import time
from kafka import KafkaProducer
from clickhouse_driver import Client

from proto.event_pb2 import EventType
from proto.event_pb2 import Event

kafka_producer = None

def produce_to_kafka():
    kafka_producer.send('event-topic', Event(
        task_id=1,
        event_type=EventType.LIKE,
        author_id=1,
        user_id=1
    ))
    kafka_producer.flush()

def read_from_clickhouse():
    client = Client(host='clickhouse', database=os.getenv("CLICKHOUSE_DB"), user=os.getenv("CLICKHOUSE_USER"), password=os.getenv("CLICKHOUSE_PASSWORD"))
    result = client.execute('SELECT * FROM task_events')
    return result


if __name__ == "__main__":
    time.sleep(30)  # wait for start

    kafka_producer = KafkaProducer(bootstrap_servers=[os.getenv("KAFKA_SERVICE")], value_serializer=lambda m: m.SerializeToString())
    produce_to_kafka()

    # Подождать 10 секунд для записи данных в ClickHouse
    for i in range(10):
        print("wait ", i)
        time.sleep(1)

    result = read_from_clickhouse()
    print("GOT:", result)
    assert len(result) == 1
    assert result[0] == (1, 'LIKE', 1, 1)
