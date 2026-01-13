-- +goose Up
-- +goose StatementBegin

CREATE TABLE multi_storages (
    storage_id            UUID PRIMARY KEY,
    primary_id            UUID NOT NULL,
    secondary_id          UUID NOT NULL
);

ALTER TABLE multi_storages
    ADD CONSTRAINT fk_multi_storages_storage
    FOREIGN KEY (storage_id)
    REFERENCES storages (id)
    ON DELETE CASCADE DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE multi_storages
    ADD CONSTRAINT fk_multi_storages_primary
    FOREIGN KEY (primary_id)
    REFERENCES storages (id)
    ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED;

ALTER TABLE multi_storages
    ADD CONSTRAINT fk_multi_storages_secondary
    FOREIGN KEY (secondary_id)
    REFERENCES storages (id)
    ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS multi_storages;

-- +goose StatementEnd
