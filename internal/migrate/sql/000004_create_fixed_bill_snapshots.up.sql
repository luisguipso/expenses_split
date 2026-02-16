CREATE TABLE fixed_bill_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fixed_bill_id UUID NOT NULL REFERENCES fixed_bills(id) ON DELETE CASCADE,
    household_id UUID NOT NULL REFERENCES households(id) ON DELETE CASCADE,
    year SMALLINT NOT NULL,
    month SMALLINT NOT NULL,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    description VARCHAR(255) NOT NULL,
    amount_cents BIGINT NOT NULL,
    due_day SMALLINT NOT NULL,
    is_shared BOOLEAN NOT NULL,
    paid_by UUID NOT NULL,
    assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
    frozen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (fixed_bill_id, year, month)
);

CREATE INDEX idx_fixed_bill_snapshots_household_month ON fixed_bill_snapshots(household_id, year, month);
