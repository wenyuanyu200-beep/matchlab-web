BEGIN;

-- Rollback is intentionally non-destructive:
-- keep added columns and migrated data so a forward deployment can be retried.
-- This only relaxes constraints that an old binary may not satisfy.

ALTER TABLE matches
    ALTER COLUMN target_id DROP NOT NULL,
    ALTER COLUMN target_type DROP NOT NULL,
    ALTER COLUMN algorithm DROP NOT NULL,
    ALTER COLUMN score DROP NOT NULL,
    ALTER COLUMN detail_scores DROP NOT NULL,
    ALTER COLUMN reason DROP NOT NULL,
    ALTER COLUMN explanation DROP NOT NULL,
    ALTER COLUMN algorithm_version DROP NOT NULL,
    ALTER COLUMN status DROP NOT NULL,
    ALTER COLUMN updated_at DROP NOT NULL;

ALTER TABLE matches
    ALTER COLUMN target_id DROP DEFAULT,
    ALTER COLUMN target_type DROP DEFAULT,
    ALTER COLUMN algorithm DROP DEFAULT,
    ALTER COLUMN detail_scores DROP DEFAULT,
    ALTER COLUMN reason DROP DEFAULT,
    ALTER COLUMN explanation DROP DEFAULT,
    ALTER COLUMN algorithm_version DROP DEFAULT,
    ALTER COLUMN status DROP DEFAULT,
    ALTER COLUMN updated_at DROP DEFAULT;

COMMIT;
