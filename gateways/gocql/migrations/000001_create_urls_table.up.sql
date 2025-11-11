DROP TABLE IF EXISTS urls;

CREATE TABLE
  urls (
    id UUID,
    raw_url TEXT,
    short_url TEXT,
    created_at TIMESTAMP,
    PRIMARY KEY (id)
  );

CREATE INDEX IF NOT EXISTS urls_short_url_idx ON urls (short_url);

CREATE INDEX IF NOT EXISTS urls_raw_url_idx ON urls (raw_url);