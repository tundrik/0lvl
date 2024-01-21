DROP TABLE item;
DROP TABLE trade;

CREATE TABLE trade (
    pk        VARCHAR(32) PRIMARY KEY,
    rang      BIGINT,
    entity    JSONB
);

CREATE TABLE trade (
    pk        VARCHAR(32) PRIMARY KEY,
    entry     VARCHAR(16)
);

CREATE TABLE item (
    trade_pk   VARCHAR(32) REFERENCES trade,
    chrt_id    INTEGER
);

CREATE TABLE trade (
    pk        VARCHAR(32) PRIMARY KEY,
    rang      BIGINT,
    entity    BYTEA
);

WITH new_order AS (INSERT INTO trade (pk, entry) VALUES ('rwmr69dshdq9nojsv98ep3qeejd6mytt', 'WBIL') RETURNING pk) 
INSERT INTO item(trade_pk, chrt_id) 
SELECT 'rwmr69dshdq9nojsv98ep3qeejd6mytt' trade_pk, chrt_id
FROM    unnest(ARRAY[1345,2345,3345,4356,524]) chrt_id;

SELECT pk, entity FROM trade;

SELECT trade.pk, trade.entry, array_agg(item) item
FROM trade
LEFT OUTER JOIN item
ON trade.pk = item.trade_pk
GROUP BY trade.pk;

SELECT DISTINCT i.chrt_id, r.pk, r.entry FROM item i JOIN trade r ON r.pk = i.trade_pk;

CREATE TABLE trade (
    pk              BIGSERIAL PRIMARY KEY,
    order_uid       VARCHAR(32) UNIQUE,
    track_number    VARCHAR(32),
    order_entry     VARCHAR(16)
);

CREATE TABLE item (
    pk         BIGSERIAL PRIMARY KEY,
    trade_pk   BIGINT REFERENCES trade,
    chrt_id    INTEGER
);


SELECT pk, order_uid FROM trade;

SELECT trade.pk, trade.order_uid, array_agg(item) item
FROM trade
LEFT OUTER JOIN item
ON trade.pk = item.trade_pk
GROUP BY trade.pk;


SELECT trade.pk, trade.order_uid, json_agg(item) item
FROM trade
LEFT OUTER JOIN item
ON trade.pk = item.trade_pk
WHERE trade.order_uid = 'rbiyqipfvjphwy88j3wn4p4mo7frx4eb'
GROUP BY trade.pk;