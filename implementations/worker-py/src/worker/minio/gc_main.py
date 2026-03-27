from __future__ import annotations

import logging

from worker.config import load_config
from worker.db.tsdb import TSDB, TSDBConfig
from worker.minio.client import MinIOClient, MinIOConfig
from worker.minio.gc import GCConfig, UploadSessionGC


def main() -> None:
    logging.basicConfig(level=logging.INFO, format="[worker-py-gc] %(message)s")
    cfg = load_config()
    tsdb = TSDB(TSDBConfig(url=cfg.tsdb_url))
    minio = MinIOClient(
        MinIOConfig(
            endpoint=cfg.minio_endpoint,
            access_key=cfg.minio_access_key,
            secret_key=cfg.minio_secret_key,
            secure=cfg.minio_secure,
            bucket=cfg.minio_bucket,
        )
    )
    gc = UploadSessionGC(
        tsdb,
        minio,
        GCConfig(
            poll_interval_sec=cfg.gc_poll_interval_sec,
            expire_minutes=cfg.gc_expire_minutes,
            batch_size=cfg.gc_batch_size,
        ),
    )
    gc.run_forever()


if __name__ == "__main__":
    main()
