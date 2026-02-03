# TSDB Migrator

## Docker Compose 例

```sh
export TSDB_URL="postgres://user:pass@host:5432/dbname?sslmode=disable"
docker compose -f deployments/compose/docker-compose.migrator.yml run --rm tsdb-migrator
```

## CI 例（GitHub Actions）

```yaml
name: tsdb-migrate
on:
  workflow_dispatch:
jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run migrator
        run: |
          docker build -f deployments/migrator/Dockerfile -t tsdb-migrator .
          docker run --rm -e TSDB_URL="${{ secrets.TSDB_URL }}" tsdb-migrator
```
