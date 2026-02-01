from __future__ import annotations

import json
import logging
from datetime import datetime, timezone
from typing import Any

from worker.config import load_config
from worker.contracts.events import Event, PayloadEnvelope, PayloadKey
from worker.kafka.consumer import KafkaConsumer
from worker.kafka.producer import KafkaProducer
from worker.runtime.dispatcher import Dispatcher
from worker.tasks.registry import build_handlers


def _encode_payload(name: str, value: Any) -> PayloadEnvelope:
    return PayloadEnvelope(
        key=PayloadKey(name=name),
        data=json.dumps(value).encode("utf-8"),
    )


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


def _build_event_id(event: Event, run_id: str | None, task_id: str | None, attempt: Any) -> str:
    if run_id and task_id is not None and attempt is not None:
        return f"{run_id}:{task_id}:{attempt}"
    return event.event_id or run_id or ""


def _build_common_payload(event: Event) -> list[PayloadEnvelope]:
    payload: list[PayloadEnvelope] = []
    for key in ("run_id", "task_id", "attempt", "task_name"):
        value = _payload_lookup(event, key)
        if value is not None:
            payload.append(_encode_payload(key, value))
    return payload


def _build_completed(event: Event, output: dict) -> Event:
    run_id = _payload_lookup(event, "run_id") or event.partition
    task_id = _payload_lookup(event, "task_id")
    attempt = _payload_lookup(event, "attempt")
    payload = _build_common_payload(event)
    payload.append(_encode_payload("output", output))
    return Event(
        event_id=_build_event_id(event, run_id, task_id, attempt),
        event_time=datetime.now(timezone.utc),
        partition=run_id or event.partition,
        type="task.completed",
        payload=payload,
    )


def _build_failed(event: Event, error: dict) -> Event:
    run_id = _payload_lookup(event, "run_id") or event.partition
    task_id = _payload_lookup(event, "task_id")
    attempt = _payload_lookup(event, "attempt")
    payload = _build_common_payload(event)
    payload.append(_encode_payload("error", error))
    return Event(
        event_id=_build_event_id(event, run_id, task_id, attempt),
        event_time=datetime.now(timezone.utc),
        partition=run_id or event.partition,
        type="task.failed",
        payload=payload,
    )


def main() -> None:
    logging.basicConfig(level=logging.INFO, format="[worker-py] %(message)s")
    cfg = load_config()
    consumer = KafkaConsumer(cfg)
    producer = KafkaProducer(cfg)
    dispatcher = Dispatcher(build_handlers())

    try:
        for event in consumer.iter_events():
            try:
                output = dispatcher.dispatch(event)
                producer.send(_build_completed(event, output))
            except Exception as exc:  # noqa: BLE001 - report error via task.failed
                error = {"message": str(exc), "code": "handler_error"}
                producer.send(_build_failed(event, error))
    finally:
        producer.flush()
        producer.close()
        consumer.close()


if __name__ == "__main__":
    main()
