from __future__ import annotations

from dataclasses import dataclass
from typing import Any

import psycopg
from psycopg.rows import dict_row

from worker.db.sql import get_query


@dataclass(frozen=True)
class TSDBConfig:
    url: str


class TSDB:
    def __init__(self, cfg: TSDBConfig) -> None:
        if not cfg.url:
            raise ValueError("TSDB_URL is required")
        self._url = cfg.url

    def _connect(self) -> psycopg.Connection:
        return psycopg.connect(self._url, row_factory=dict_row)

    def insert_processed_event(self, event_id: str, task_id: str | None, attempt_no: int | None) -> bool:
        query = get_query("InsertProcessedEvent")
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(query, (event_id, task_id, attempt_no))
            row = cur.fetchone()
            return row is not None

    def upsert_task_attempt_started(
        self,
        task_id: str,
        attempt_no: int,
        workflow_run_id: str | None,
        status: str,
        started_at: str | None,
        worker_id: str | None,
    ) -> None:
        query = get_query("UpsertTaskAttemptStarted")
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(
                query,
                (task_id, attempt_no, workflow_run_id, status, started_at, worker_id),
            )

    def update_task_attempt_completed(
        self,
        task_id: str,
        attempt_no: int,
        status: str,
        finished_at: str | None,
        output_ref: str | None,
    ) -> None:
        query = get_query("UpdateTaskAttemptCompleted")
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(query, (task_id, attempt_no, status, finished_at, output_ref))

    def update_task_attempt_failed(
        self,
        task_id: str,
        attempt_no: int,
        status: str,
        finished_at: str | None,
        error: str | None,
    ) -> None:
        query = get_query("UpdateTaskAttemptFailed")
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(query, (task_id, attempt_no, status, finished_at, error))

    def insert_outbox_event(self, event_type: str, partition_key: str, payload: dict[str, Any]) -> None:
        query = get_query("InsertOutboxEvent")
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(query, (event_type, partition_key, payload))

    def list_pending_outbox_events(self, limit: int) -> list[dict[str, Any]]:
        query = get_query("ListPendingOutboxEvents")
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(query, (limit,))
            return list(cur.fetchall())

    def mark_outbox_event_published(self, event_id: int, published_at: str) -> None:
        query = get_query("MarkOutboxEventPublished")
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(query, (event_id, published_at))

    def mark_outbox_event_failed(self, event_id: int, error: str, next_publish_at: str) -> None:
        query = get_query("MarkOutboxEventFailed")
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(query, (event_id, error, next_publish_at))

    def insert_artifact(self, payload: dict[str, Any]) -> None:
        query = get_query("InsertArtifact")
        params = (
            payload["artifact_id"],
            payload["workflow_run_id"],
            payload["task_id"],
            payload["attempt_no"],
            payload["bucket"],
            payload["object_key"],
            payload["kind"],
            payload.get("filename"),
            payload.get("content_type"),
            payload["size_bytes"],
            payload.get("checksum_sha256"),
            payload.get("etag"),
            payload["status"],
            payload.get("tags", {}),
        )
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(query, params)

    def insert_upload_session(self, payload: dict[str, Any]) -> None:
        query = get_query("InsertUploadSession")
        params = (
            payload["upload_id"],
            payload["workflow_run_id"],
            payload["task_id"],
            payload["attempt_no"],
            payload["bucket"],
            payload["object_key"],
            payload["kind"],
            payload.get("filename"),
            payload.get("content_type"),
            payload["status"],
        )
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(query, params)

    def mark_upload_session_uploaded(self, upload_id: str, status: str, uploaded_at: str) -> None:
        query = get_query("MarkUploadSessionUploaded")
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(query, (upload_id, status, uploaded_at))

    def mark_upload_session_failed(self, upload_id: str, status: str, error: str) -> None:
        query = get_query("MarkUploadSessionFailed")
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(query, (upload_id, status, error))

    def list_expired_upload_sessions(self, expired_before: str, limit: int) -> list[dict[str, Any]]:
        query = get_query("ListExpiredUploadSessions")
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(query, (expired_before, limit))
            return list(cur.fetchall())

    def mark_upload_session_expired(self, upload_id: str) -> None:
        query = get_query("MarkUploadSessionExpired")
        with self._connect() as conn, conn.cursor() as cur:
            cur.execute(query, (upload_id,))
