from __future__ import annotations

from collections.abc import Callable

from worker.contracts.events import Event

Handler = Callable[[Event], dict]


class Dispatcher:
    def __init__(self, handlers: dict[str, Handler]) -> None:
        self._handlers = handlers

    def dispatch(self, event: Event) -> dict:
        handler = self._handlers.get(event.type)
        if handler is None:
            raise ValueError(f"unknown event.type={event.type}")
        return handler(event)
