CREATE TABLE IF NOT EXISTS task_statistics (
    task_id Int32,
    likes UInt64 DEFAULT 0,
    views UInt64 DEFAULT 0
) ENGINE = SummingMergeTree()
ORDER BY task_id;

CREATE TABLE IF NOT EXISTS events (
    task_id Int32,
    event_type Enum8('VIEW' = 0, 'LIKE' = 1)
)
ENGINE = Kafka
SETTINGS kafka_broker_list = 'kafka:9092',
       kafka_topic_list = 'event-topic',
       kafka_group_name = 'readings_consumer_group1',
       kafka_format = 'ProtobufSingle',
       kafka_schema = 'event.proto:Event',
       kafka_max_block_size = 1048576;

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_likes_views TO task_statistics AS
SELECT
    task_id,
    if(event_type = 'LIKE', 1, 0) AS likes,
    if(event_type = 'VIEW', 1, 0) AS views
FROM
    events;