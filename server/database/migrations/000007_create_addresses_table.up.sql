CREATE TABLE addresses
(
    id      BIGSERIAL,
    address VARCHAR(71) NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT idx_addresses_address UNIQUE  (address)
)
