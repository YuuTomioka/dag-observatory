-- Symbol master seeds
INSERT INTO symbol (code, base, quote, price_scale, tick_size_raw, pip_size_raw)
VALUES
  ('EURUSD', 'EUR', 'USD', 5, 1, 10),
  ('USDJPY', 'USD', 'JPY', 3, 1, 10),
  ('GBPUSD', 'GBP', 'USD', 5, 1, 10),
  ('AUDUSD', 'AUD', 'USD', 5, 1, 10),
  ('USDCAD', 'USD', 'CAD', 5, 1, 10),
  ('USDCHF', 'USD', 'CHF', 5, 1, 10),
  ('NZDUSD', 'NZD', 'USD', 5, 1, 10),
  ('EURJPY', 'EUR', 'JPY', 3, 1, 10),
  ('GBPJPY', 'GBP', 'JPY', 3, 1, 10),
  ('EURGBP', 'EUR', 'GBP', 5, 1, 10)
ON CONFLICT (code) DO NOTHING;