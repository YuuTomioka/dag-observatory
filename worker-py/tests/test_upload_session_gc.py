import unittest
from datetime import datetime, timedelta, timezone

from worker.minio.gc import GCConfig, UploadSessionGC


class _FakeTSDB:
    def __init__(self, rows):
        self._rows = rows
        self.expired = []

    def list_expired_upload_sessions(self, expired_before: str, limit: int):
        # Keep it simple: trust caller sets cutoff; return configured rows.
        return self._rows[:limit]

    def mark_upload_session_expired(self, upload_id: str) -> None:
        self.expired.append(upload_id)


class _FakeMinIO:
    def __init__(self, fail: bool = False):
        self.fail = fail
        self.removed = []

    def remove_object(self, bucket: str, object_key: str) -> None:
        if self.fail:
            raise RuntimeError("remove failed")
        self.removed.append((bucket, object_key))


class UploadSessionGCTests(unittest.TestCase):
    def test_gc_removes_object_and_marks_expired(self):
        rows = [{"upload_id": "u1", "bucket": "b", "object_key": "k"}]
        tsdb = _FakeTSDB(rows)
        minio = _FakeMinIO(fail=False)
        gc = UploadSessionGC(tsdb, minio, GCConfig(poll_interval_sec=0.0, expire_minutes=30, batch_size=10))

        gc._run_once()  # noqa: SLF001 - unit test

        self.assertEqual(minio.removed, [("b", "k")])
        self.assertEqual(tsdb.expired, ["u1"])

    def test_gc_failure_is_best_effort(self):
        rows = [{"upload_id": "u2", "bucket": "b", "object_key": "k"}]
        tsdb = _FakeTSDB(rows)
        minio = _FakeMinIO(fail=True)
        gc = UploadSessionGC(tsdb, minio, GCConfig(poll_interval_sec=0.0, expire_minutes=30, batch_size=10))

        # Should not raise even if MinIO delete fails.
        gc._run_once()  # noqa: SLF001 - unit test

        self.assertEqual(tsdb.expired, [])


if __name__ == "__main__":
    unittest.main()

