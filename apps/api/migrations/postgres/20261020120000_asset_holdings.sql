-- +goose Up
-- +goose StatementBegin

-- ---------------------------------------------------------------------------
-- Daily-priced asset holdings [033]. An asset item either carries a worth the
-- user sets by hand, or it is linked to a market-data listing and its worth
-- follows that listing's daily price. The link is a column on the item rather
-- than a side table: every item has at most one listing, and a NULL reads as
-- "manual", which is exactly how every pre-existing row behaves.
--
-- listing_id is deliberately not a foreign key. Listings live in the marketdata
-- schema and are global reference data curated by administrators, while items
-- are account data; a cross-schema constraint would make deleting a listing fail
-- on someone else's holding. The assets feature resolves the listing through
-- marketdata.Queries and treats a missing one as "no price yet".
-- ---------------------------------------------------------------------------
ALTER TABLE assets.items
    ADD COLUMN listing_id UUID;

CREATE INDEX idx_asset_items_listing ON assets.items (listing_id)
    WHERE listing_id IS NOT NULL;

-- One recorded purchase of a linked item: when, how much, and the price paid per
-- unit. Quantity is a float because crypto is held in fractions far below one
-- unit; unit_price uses the same fixed scale as every other money column.
-- There are no sells in this feature, so a holding only ever grows.
CREATE TABLE assets.purchases (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id   UUID             NOT NULL REFERENCES assets.accounts(account_id) ON DELETE CASCADE,
    item_id      UUID             NOT NULL REFERENCES assets.items(id) ON DELETE CASCADE,
    purchased_on DATE             NOT NULL,
    quantity     DOUBLE PRECISION NOT NULL CHECK (quantity > 0),
    unit_price   BIGINT           NOT NULL CHECK (unit_price >= 0),
    created_at   TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- The rebuild walks one item's purchases in date order; the account index serves
-- the per-account rebuild that loads them all at once.
CREATE INDEX idx_asset_purchases_item_date ON assets.purchases (item_id, purchased_on);
CREATE INDEX idx_asset_purchases_account ON assets.purchases (account_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS assets.purchases;
DROP INDEX IF EXISTS assets.idx_asset_items_listing;
ALTER TABLE assets.items DROP COLUMN IF EXISTS listing_id;

-- +goose StatementEnd
