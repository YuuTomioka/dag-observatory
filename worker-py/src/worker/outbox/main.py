from __future__ import annotations

import logging

from worker.config import load_config
from worker.db.tsdb import TSDB, TSDBConfig
from worker.kafka.producer import KafkaProducer
from worker.outbox.publisher import OutboxPublisher, PublisherConfig


def main() -> None:
    logging.basicConfig(level=logging.INFO, format="[worker-py-outbox] %(message)s")
    cfg = load_config()
    tsdb = TSDB(TSDBConfig(url=cfg.tsdb_url))
    producer = KafkaProducer(cfg)
    publisher = OutboxPublisher(
        tsdb,
        producer,
        PublisherConfig(
            poll_interval_sec=cfg.outbox_poll_interval_sec,
            batch_size=cfg.outbox_batch_size,
        ),
    )
    try:
        publisher.run_forever()
    finally:
        producer.flush()
        producer.close()


if __name__ == "__main__":
    main()
