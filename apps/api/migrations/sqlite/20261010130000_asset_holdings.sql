-- +goose Up
-- +goose StatementBegin

-- Mirror of migrations/postgres/20261009120000_asset_holdings.sql. See there for why
-- the listing link is a nullable column on the item and carries no foreign key.
ALTER TABLE asset_items ADD COLUMN listing_id TEXT;

CREATE INDEX idx_asset_items_listing ON asset_items (listing_id) WHERE listing_id IS NOT NULL;

CREATE TABLE asset_purchases (
    id           TEXT PRIMARY KEY,
    account_id   TEXT     NOT NULL REFERENCES assets_accounts(account_id) ON DELETE CASCADE,
    item_id      TEXT     NOT NULL REFERENCES asset_items(id) ON DELETE CASCADE,
    purchased_on DATE     NOT NULL,
    quantity     REAL     NOT NULL CHECK (quantity > 0),
    unit_price   INTEGER  NOT NULL CHECK (unit_price >= 0),
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_asset_purchases_item_date ON asset_purchases (item_id, purchased_on);
CREATE INDEX idx_asset_purchases_account ON asset_purchases (account_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS asset_purchases;
DROP INDEX IF EXISTS idx_asset_items_listing;
ALTER TABLE asset_items DROP COLUMN listing_id;

-- +goose StatementEnd
