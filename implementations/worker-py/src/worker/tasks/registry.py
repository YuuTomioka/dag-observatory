from __future__ import annotations

from collections.abc import Callable

from worker.contracts.events import Event
from worker.tasks.example.heavy_calc import heavy_calc

Handler = Callable[[Event], dict]


def build_handlers() -> dict[str, Handler]:
    return {
        "task.requested": heavy_calc,
    }
