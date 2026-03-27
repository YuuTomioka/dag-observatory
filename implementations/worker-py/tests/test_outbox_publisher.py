import unittest
from datetime import datetime, timezone

from worker.outbox.publisher import OutboxPublisher, PublisherConfig


class _FakeTSDB:
    def __init__(self, rows):
        self._rows = rows
        self.published = []
        self.failed = []

    def list_pending_outbox_events(self, limit: int):
        return self._rows[:limit]

    def mark_outbox_event_published(self, event_id: int, published_at: str) -> None:
        self.published.append((event_id, published_at))

    def mark_outbox_event_failed(self, event_id: int, error: str, next_publish_at: str) -> None:
        self.failed.append((event_id, error, next_publish_at))


class _FakeProducer:
    def __init__(self, fail: bool = False):
        self.fail = fail
        self.sent = []

    def send(self, event, headers=None):  # matches worker.kafka.producer.KafkaProducer.send
        if self.fail:
            raise RuntimeError("send failed")
        self.sent.append((event, headers))


class OutboxPublisherTests(unittest.TestCase):
    def test_publish_success_marks_published(self):
        rows = [
            {
                "id": 1,
                "event_type": "task.completed",
                "partition_key": "run-1",
                "payload_json": {
                    "event_id": "e1",
                    "event_time": datetime.now(timezone.utc).isoformat(),
                    "partition": "run-1",
                    "type": "task.completed",
                    "payload": [],
                },
                "created_at": datetime.now(timezone.utc),
                "attempts": 0,
            }
        ]
        tsdb = _FakeTSDB(rows)
        producer = _FakeProducer(fail=False)
        pub = OutboxPublisher(tsdb, producer, PublisherConfig(poll_interval_sec=0.0, batch_size=10))

        pub._publish_row(rows[0])  # noqa: SLF001 - unit test

        self.assertEqual(len(producer.sent), 1)
        self.assertEqual(len(tsdb.published), 1)
        self.assertEqual(tsdb.published[0][0], 1)
        self.assertEqual(len(tsdb.failed), 0)

    def test_publish_failure_marks_failed_and_sets_next_publish_at(self):
        rows = [
            {
                "id": 2,
                "event_type": "task.completed",
                "partition_key": "run-2",
                "payload_json": {
                    "event_id": "e2",
                    "event_time": datetime.now(timezone.utc).isoformat(),
                    "partition": "run-2",
                    "type": "task.completed",
                    "payload": [],
                },
                "created_at": datetime.now(timezone.utc),
                "attempts": 2,
            }
        ]
        tsdb = _FakeTSDB(rows)
        producer = _FakeProducer(fail=True)
        pub = OutboxPublisher(tsdb, producer, PublisherConfig(poll_interval_sec=0.0, batch_size=10))

        pub._publish_row(rows[0])  # noqa: SLF001 - unit test

        self.assertEqual(len(producer.sent), 0)
        self.assertEqual(len(tsdb.published), 0)
        self.assertEqual(len(tsdb.failed), 1)
        self.assertEqual(tsdb.failed[0][0], 2)
        self.assertTrue(tsdb.failed[0][2])  # next_publish_at is non-empty


if __name__ == "__main__":
    unittest.main()

