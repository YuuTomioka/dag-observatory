from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path

from minio import Minio


@dataclass(frozen=True)
class MinIOConfig:
    endpoint: str
    access_key: str
    secret_key: str
    secure: bool
    bucket: str


class MinIOClient:
    def __init__(self, cfg: MinIOConfig) -> None:
        if not cfg.endpoint:
            raise ValueError("MINIO_ENDPOINT is required")
        if not cfg.access_key or not cfg.secret_key:
            raise ValueError("MINIO_ACCESS_KEY and MINIO_SECRET_KEY are required")
        if not cfg.bucket:
            raise ValueError("MINIO_BUCKET is required")
        self._bucket = cfg.bucket
        self._client = Minio(
            cfg.endpoint,
            access_key=cfg.access_key,
            secret_key=cfg.secret_key,
            secure=cfg.secure,
        )

    def ensure_bucket(self) -> None:
        if not self._client.bucket_exists(self._bucket):
            self._client.make_bucket(self._bucket)

    def upload_file(self, object_key: str, file_path: str, content_type: str | None = None) -> dict:
        self.ensure_bucket()
        result = self._client.fput_object(
            self._bucket,
            object_key,
            file_path,
            content_type=content_type,
        )
        size_bytes = Path(file_path).stat().st_size
        return {
            "bucket": self._bucket,
            "object_key": object_key,
            "etag": result.etag,
            "size_bytes": size_bytes,
        }
