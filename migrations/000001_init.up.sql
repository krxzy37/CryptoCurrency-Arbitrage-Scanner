CREATE SCHEMA crypto_scanner;

CREATE TABLE crypto_scanner.exchanges (
    id          uuid                PRIMARY KEY,
    name        VARCHAR(50)         NOT NULL UNIQUE ,
    maker_fee   NUMERIC(5,4)        NOT NULL DEFAULT 0.0010,
    taker_fee   NUMERIC(5,4)        NOT NULL DEFAULT 0.0010,
    is_active   BOOLEAN             NOT NULL DEFAULT TRUE
);

CREATE TABLE trading_pairs (
    id              uuid            PRIMARY KEY,
    exchange_id     uuid            NOT NULL REFERENCES crypto_scanner.exchanges(id),
    base_asset      VARCHAR(10)     NOT NULL,
    quote_offset    VARCHAR(10)     NOT NULL,
    symbol          VARCHAR(20)     NOT NULL,
    min_qty         NUMERIC(18,8)   NOT NULL,

    CONSTRAINT uq_exchange_symbol UNIQUE (exchange_id, symbol)
);

CREATE TABLE trades_history (
    id              uuid                        PRIMARY KEY,
    executed_at     TIMESTAMP WITH TIME ZONE    NOT NULL        DEFAULT NOW(),
    route           JSONB                       NOT NULL,
    invested_amount NUMERIC(18, 8)              NOT NULL,
    expected_profit NUMERIC(18, 8)              NOT NULL,
    actual_profit   NUMERIC(18, 8),
    status          VARCHAR(20)                 NOT NULL
);
