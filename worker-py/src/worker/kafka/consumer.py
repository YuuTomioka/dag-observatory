from __future__ import annotations

from typing import Iterable

from worker.config import Config
from worker.kafka.serde import decode_event


class KafkaConsumer:
    def __init__(self, cfg: Config) -> None:
        try:
            from kafka import KafkaConsumer as _KafkaConsumer
        except ImportError as exc:  # pragma: no cover - runtime dependency
            raise RuntimeError("kafka-python is required to run the worker") from exc

        if not cfg.kafka_brokers:
            raise ValueError("KAFKA_BROKERS is required")

        self._consumer = _KafkaConsumer(
            cfg.kafka_topic,
            bootstrap_servers=cfg.kafka_brokers.split(","),
            group_id=cfg.kafka_group_id,
            enable_auto_commit=True,
            auto_offset_reset="earliest",
        )

    def iter_events(self) -> Iterable:
        for msg in self._consumer:
            yield decode_event(msg.value)

    def close(self) -> None:
        self._consumer.close()
