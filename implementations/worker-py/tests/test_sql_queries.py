import unittest

from worker.db import sql


class SQLQueryNormalizationTests(unittest.TestCase):
    def test_normalize_postgres_placeholders_for_psycopg(self) -> None:
        query = "select * from t where a = $1 and b = $2 limit $3"
        got = sql._normalize_placeholders(query)
        self.assertEqual(got, "select * from t where a = %s and b = %s limit %s")


if __name__ == "__main__":
    unittest.main()
