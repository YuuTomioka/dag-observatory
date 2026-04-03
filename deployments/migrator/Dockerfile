FROM postgres:16-alpine

WORKDIR /migrator

COPY scripts/tsdb/migrate.sh /migrator/migrate.sh
COPY data/tsdb/schema/migrate /migrator/data/tsdb/schema/migrate
COPY data/tsdb/schema/seed /migrator/data/tsdb/schema/seed

ENTRYPOINT ["/migrator/migrate.sh"]
