ALTER TABLE fixed_bills
    ADD COLUMN paid_by UUID REFERENCES users(id) ON DELETE SET NULL;

-- Backfill: use assigned_to if set, otherwise use household admin
UPDATE fixed_bills fb
SET paid_by = COALESCE(
    fb.assigned_to,
    (SELECT hm.user_id FROM household_members hm
     WHERE hm.household_id = fb.household_id AND hm.role = 'admin'
     LIMIT 1)
);

ALTER TABLE fixed_bills
    ALTER COLUMN paid_by SET NOT NULL;
