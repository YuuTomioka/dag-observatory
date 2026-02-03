FROM postgres:16-alpine

WORKDIR /migrator

COPY scripts/tsdb/migrate.sh /migrator/migrate.sh
COPY docs/tsdb/schema/migrate /migrator/docs/tsdb/schema/migrate

ENTRYPOINT ["/migrator/migrate.sh"]
