from __future__ import annotations

import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Config:
    kafka_brokers: str
    kafka_in_topic: str
    kafka_out_topic: str
    kafka_group_id: str
    service_name: str
    env: str
    otlp_endpoint: str


def _getenv(key: str, default: str) -> str:
    value = os.getenv(key)
    return value if value else default


def load_config() -> Config:
    return Config(
        kafka_brokers=_getenv("KAFKA_BROKERS", ""),
        kafka_in_topic=_getenv("KAFKA_IN_TOPIC", _getenv("KAFKA_TOPIC", "dagruntime-events")),
        kafka_out_topic=_getenv("KAFKA_OUT_TOPIC", _getenv("KAFKA_TOPIC", "dagruntime-events")),
        kafka_group_id=_getenv("KAFKA_GROUP_ID", "dagruntime-worker-py"),
        service_name=_getenv("SERVICE_NAME", "dag-observatory-worker-py"),
        env=_getenv("ENV", "dev"),
        otlp_endpoint=_getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4318"),
    )
