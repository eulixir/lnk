DROP TABLE IF EXISTS urls;

CREATE TABLE
  urls (
    short_code TEXT,
    long_url TEXT,
    created_at TIMESTAMP,
    PRIMARY KEY (short_code)
  );