ALTER TABLE monthly_summary_items
    ADD COLUMN total_paid_cents BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN balance_cents BIGINT NOT NULL DEFAULT 0;
