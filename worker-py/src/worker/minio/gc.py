from __future__ import annotations

import logging
import time
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone

from worker.db.tsdb import TSDB
from worker.minio.client import MinIOClient


@dataclass(frozen=True)
class GCConfig:
    poll_interval_sec: float
    expire_minutes: int
    batch_size: int


class UploadSessionGC:
    def __init__(self, tsdb: TSDB, minio: MinIOClient, cfg: GCConfig) -> None:
        self._tsdb = tsdb
        self._minio = minio
        self._cfg = cfg

    def run_forever(self) -> None:
        logging.info("upload session GC started")
        while True:
            self._run_once()
            time.sleep(self._cfg.poll_interval_sec)

    def _run_once(self) -> None:
        cutoff = datetime.now(timezone.utc) - timedelta(minutes=self._cfg.expire_minutes)
        rows = self._tsdb.list_expired_upload_sessions(cutoff.isoformat(), self._cfg.batch_size)
        for row in rows:
            upload_id = str(row.get("upload_id", ""))
            bucket = row.get("bucket", "")
            object_key = row.get("object_key", "")
            try:
                if bucket and object_key:
                    self._minio.remove_object(bucket, object_key)
                if upload_id:
                    self._tsdb.mark_upload_session_expired(str(upload_id))
            except Exception as exc:  # noqa: BLE001 - best-effort cleanup
                logging.warning("gc failed upload_id=%s err=%s", upload_id, exc)
