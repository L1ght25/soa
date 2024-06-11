USE statistics;

CREATE TABLE IF NOT EXISTS task_events (
    task_id Int32,
    event_type Enum8('VIEW' = 0, 'LIKE' = 1),
    author_id Int32,
    user_id Int32
) ENGINE = ReplacingMergeTree()
ORDER BY (event_type, task_id, user_id);

CREATE TABLE IF NOT EXISTS events (
    task_id Int32,
    event_type Enum8('VIEW' = 0, 'LIKE' = 1),
    author_id Int32,
    user_id Int32
)
ENGINE = Kafka
SETTINGS kafka_broker_list = 'kafka:9092',
       kafka_topic_list = 'event-topic',
       kafka_group_name = 'readings_consumer_group1',
       kafka_format = 'ProtobufSingle',
       kafka_schema = 'event.proto:Event',
       kafka_max_block_size = 1048576;

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_unique_events TO task_events AS
SELECT
    task_id,
    event_type,
    author_id,
    user_id
FROM events;
