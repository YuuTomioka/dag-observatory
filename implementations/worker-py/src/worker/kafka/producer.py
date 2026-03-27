from __future__ import annotations

from worker.config import Config
from worker.contracts.events import Event
from worker.kafka.serde import encode_event


class KafkaProducer:
    def __init__(self, cfg: Config) -> None:
        try:
            from kafka import KafkaProducer as _KafkaProducer
        except ImportError as exc:  # pragma: no cover - runtime dependency
            raise RuntimeError("kafka-python is required to run the worker") from exc

        if not cfg.kafka_brokers:
            raise ValueError("KAFKA_BROKERS is required")

        self._topic = cfg.kafka_out_topic
        self._producer = _KafkaProducer(
            bootstrap_servers=cfg.kafka_brokers.split(","),
        )

    def send(self, event: Event, headers: dict[str, str] | None = None) -> None:
        payload = encode_event(event)
        key = event.partition.encode("utf-8") if event.partition else None
        header_list = None
        if headers:
            header_list = [(k, v.encode("utf-8")) for k, v in headers.items() if v]
        self._producer.send(self._topic, value=payload, key=key, headers=header_list)

    def flush(self) -> None:
        self._producer.flush()

    def close(self) -> None:
        self._producer.flush()
        self._producer.close()
