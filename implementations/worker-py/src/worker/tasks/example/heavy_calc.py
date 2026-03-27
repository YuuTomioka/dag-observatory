from __future__ import annotations

import json
import logging
from typing import Any

from worker.contracts.events import Event, PayloadEnvelope


def _decode_payload(envelope: PayloadEnvelope) -> Any:
    try:
        return json.loads(envelope.data.decode("utf-8"))
    except json.JSONDecodeError:
        return None


def _payload_lookup(event: Event, name: str) -> Any:
    for envelope in event.payload:
        if envelope.key.name == name:
            return _decode_payload(envelope)
    return None


def heavy_calc(event: Event) -> dict:
    task_name = _payload_lookup(event, "task_name")
    if task_name and task_name != "heavy_calc":
        raise ValueError(f"unsupported task_name={task_name}")

    logging.info("received task.requested for heavy_calc")

    payload = _payload_lookup(event, "input") or {}
    x = payload.get("x", 0)
    y = payload.get("y", 0)
    return {"sum": x + y}
