DROP TABLE IF EXISTS urls;

CREATE TABLE
  urls (
    id UUID,
    long_url TEXT,
    short_code TEXT,
    created_at TIMESTAMP,
    PRIMARY KEY (id)
  );

CREATE INDEX IF NOT EXISTS urls_short_code_idx ON urls (short_code);

CREATE INDEX IF NOT EXISTS urls_long_url_idx ON urls (long_url);