FROM postgres:16-alpine

WORKDIR /migrator

COPY scripts/tsdb/migrate.sh /migrator/migrate.sh
COPY data/tsdb/schema/migrate /migrator/data/tsdb/schema/migrate

ENTRYPOINT ["/migrator/migrate.sh"]
