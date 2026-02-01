from __future__ import annotations

import base64
import json
from datetime import datetime, timezone
from typing import Any

from worker.contracts.events import Event, PayloadEnvelope, PayloadKey


_TIME_FORMAT = "%Y-%m-%dT%H:%M:%S.%f%z"


def _encode_time(value: datetime | None) -> str:
    if value is None:
        return ""
    if value.tzinfo is None:
        value = value.replace(tzinfo=timezone.utc)
    return value.strftime(_TIME_FORMAT)


def _decode_time(raw: str) -> datetime | None:
    if not raw:
        return None
    try:
        return datetime.strptime(raw, _TIME_FORMAT)
    except ValueError:
        return datetime.fromisoformat(raw)


def encode_event(event: Event) -> bytes:
    payload = []
    for envelope in event.payload:
        payload.append(
            {
                "Key": {"Name": envelope.key.name, "StableID": envelope.key.stable_id},
                "Data": base64.b64encode(envelope.data).decode("ascii"),
            }
        )
    msg = {
        "event_id": event.event_id,
        "event_time": _encode_time(event.event_time),
        "partition": event.partition,
        "type": event.type,
        "payload": payload,
    }
    return json.dumps(msg, separators=(",", ":")).encode("utf-8")


def decode_event(raw: bytes) -> Event:
    msg: dict[str, Any] = json.loads(raw.decode("utf-8"))
    payload: list[PayloadEnvelope] = []
    for envelope in msg.get("payload", []) or []:
        key = envelope.get("Key", {})
        payload.append(
            PayloadEnvelope(
                key=PayloadKey(name=key.get("Name", ""), stable_id=key.get("StableID", "")),
                data=base64.b64decode(envelope.get("Data", "") or b""),
            )
        )
    return Event(
        event_id=msg.get("event_id", ""),
        event_time=_decode_time(msg.get("event_time", "")),
        partition=msg.get("partition", ""),
        type=msg.get("type", ""),
        payload=payload,
    )
