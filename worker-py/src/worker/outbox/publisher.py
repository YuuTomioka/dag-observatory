from __future__ import annotations

import base64
import logging
import time
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from typing import Any

from worker.contracts.events import Event, PayloadEnvelope, PayloadKey
from worker.db.tsdb import TSDB
from worker.kafka.producer import KafkaProducer


@dataclass(frozen=True)
class PublisherConfig:
    poll_interval_sec: float
    batch_size: int


def _decode_payload(payload: list[dict[str, Any]]) -> list[PayloadEnvelope]:
    envelopes: list[PayloadEnvelope] = []
    for envelope in payload:
        key = envelope.get("Key") or envelope.get("key") or {}
        name = key.get("Name") or key.get("name") or ""
        stable_id = key.get("StableID") or key.get("stable_id") or ""
        raw = envelope.get("Data") or envelope.get("data") or ""
        try:
            data = base64.b64decode(raw) if isinstance(raw, str) else bytes(raw)
        except Exception:
            data = b""
        envelopes.append(PayloadEnvelope(key=PayloadKey(name=name, stable_id=stable_id), data=data))
    return envelopes


def _build_event(event_type: str, partition_key: str, payload_json: dict[str, Any]) -> Event:
    event_id = payload_json.get("event_id", "")
    raw_time = payload_json.get("event_time", "")
    event_time: datetime | None = None
    if raw_time:
        try:
            event_time = datetime.fromisoformat(raw_time.replace("Z", "+00:00"))
        except ValueError:
            event_time = None
    payload = _decode_payload(payload_json.get("payload", []))
    return Event(
        event_id=event_id,
        event_time=event_time,
        partition=payload_json.get("partition", "") or partition_key,
        type=payload_json.get("type", "") or event_type,
        payload=payload,
    )


def _next_publish_time(attempts: int) -> datetime:
    delay = min(60, max(1, 2 ** attempts))
    return datetime.now(timezone.utc) + timedelta(seconds=delay)


class OutboxPublisher:
    def __init__(self, tsdb: TSDB, producer: KafkaProducer, cfg: PublisherConfig) -> None:
        self._tsdb = tsdb
        self._producer = producer
        self._cfg = cfg

    def run_forever(self) -> None:
        logging.info("outbox publisher started")
        while True:
            rows = self._tsdb.list_pending_outbox_events(self._cfg.batch_size)
            if not rows:
                time.sleep(self._cfg.poll_interval_sec)
                continue
            for row in rows:
                self._publish_row(row)

    def _publish_row(self, row: dict[str, Any]) -> None:
        try:
            event = _build_event(row["event_type"], row["partition_key"], row["payload_json"])
            self._producer.send(event, headers=None)
            published_at = datetime.now(timezone.utc).isoformat()
            self._tsdb.mark_outbox_event_published(row["id"], published_at)
        except Exception as exc:  # noqa: BLE001 - report and retry
            attempts = int(row.get("attempts") or 0)
            next_time = _next_publish_time(attempts).isoformat()
            self._tsdb.mark_outbox_event_failed(row["id"], str(exc), next_time)
