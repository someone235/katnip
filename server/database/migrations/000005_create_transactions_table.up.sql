CREATE TABLE transactions
(
    id                 BIGSERIAL,
    accepting_block_id BIGINT NULL,
    transaction_hash   CHAR(64)                                      NOT NULL,
    transaction_id     CHAR(64)                                      NOT NULL,
    lock_time          BYTEA                                         NOT NULL,
    subnetwork_id      BIGINT                                        NOT NULL,
    gas                BIGINT CHECK (gas >= 0)                       NOT NULL,
    payload            BYTEA                                         NOT NULL,
    mass               BIGINT CHECK (mass >= 0)                      NOT NULL,
    version            INT CHECK (version >= 0 AND version <= 65535) NOT NULL,
    PRIMARY KEY (id)
);

CREATE
INDEX idx_transactions_transaction_id ON transactions (transaction_id);

CREATE
INDEX idx_transactions_transaction_hash ON transactions (transaction_hash);
