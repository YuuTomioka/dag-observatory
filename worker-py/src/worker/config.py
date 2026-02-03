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
    tsdb_url: str
    minio_endpoint: str
    minio_access_key: str
    minio_secret_key: str
    minio_secure: bool
    minio_bucket: str
    outbox_poll_interval_sec: float
    outbox_batch_size: int


def _getenv(key: str, default: str) -> str:
    value = os.getenv(key)
    return value if value else default


def _getenv_bool(key: str, default: bool) -> bool:
    value = os.getenv(key)
    if value is None or value == "":
        return default
    return value.lower() in {"1", "true", "yes", "on"}


def load_config() -> Config:
    return Config(
        kafka_brokers=_getenv("KAFKA_BROKERS", ""),
        kafka_in_topic=_getenv("KAFKA_IN_TOPIC", _getenv("KAFKA_TOPIC", "dagruntime-events")),
        kafka_out_topic=_getenv("KAFKA_OUT_TOPIC", _getenv("KAFKA_TOPIC", "dagruntime-events")),
        kafka_group_id=_getenv("KAFKA_GROUP_ID", "dagruntime-worker-py"),
        service_name=_getenv("SERVICE_NAME", "dag-observatory-worker-py"),
        env=_getenv("ENV", "dev"),
        otlp_endpoint=_getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4318"),
        tsdb_url=_getenv("TSDB_URL", ""),
        minio_endpoint=_getenv("MINIO_ENDPOINT", ""),
        minio_access_key=_getenv("MINIO_ACCESS_KEY", ""),
        minio_secret_key=_getenv("MINIO_SECRET_KEY", ""),
        minio_secure=_getenv_bool("MINIO_SECURE", False),
        minio_bucket=_getenv("MINIO_BUCKET", "worker-results"),
        outbox_poll_interval_sec=float(_getenv("OUTBOX_POLL_INTERVAL_SEC", "1.0")),
        outbox_batch_size=int(_getenv("OUTBOX_BATCH_SIZE", "100")),
    )
