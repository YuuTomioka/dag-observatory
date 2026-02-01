from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime
from typing import Any


@dataclass(frozen=True)
class PayloadKey:
    name: str
    stable_id: str = ""


@dataclass(frozen=True)
class PayloadEnvelope:
    key: PayloadKey
    data: bytes


@dataclass(frozen=True)
class Event:
    event_id: str
    event_time: datetime | None
    partition: str
    type: str
    payload: list[PayloadEnvelope]


EventPayload = list[PayloadEnvelope]
