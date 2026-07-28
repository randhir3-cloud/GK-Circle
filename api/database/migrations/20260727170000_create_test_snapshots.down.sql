-- +migrate Down
-- EXAM-P4-T04: drop immutable test composition snapshot tables.

DROP TABLE IF EXISTS test_snapshot_items;
DROP TABLE IF EXISTS test_snapshots;
