USE `cashlenx`;

-- The verification-token columns and indexes are part of the canonical
-- operation_confirm_codes table created by migration 006. Keep this numbered
-- migration as an explicit compatibility marker for databases that already
-- applied the earlier development sequence.
SELECT 'verification token schema already present from migration 006' AS migration_status;
