BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    target_id UUID NOT NULL,
    target_type TEXT NOT NULL DEFAULT 'activity',
    questionnaire_id UUID REFERENCES questionnaires(id) ON DELETE SET NULL,
    algorithm TEXT NOT NULL DEFAULT 'rules',
    score NUMERIC(5, 2) NOT NULL DEFAULT 0 CHECK (score >= 0 AND score <= 100),
    detail_scores JSONB NOT NULL DEFAULT '{}'::JSONB,
    reason TEXT NOT NULL DEFAULT '',
    explanation JSONB NOT NULL DEFAULT '{}'::JSONB,
    algorithm_version VARCHAR(32) NOT NULL DEFAULT 'activity-rules-v1',
    status VARCHAR(32) NOT NULL DEFAULT 'recommended',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE matches
    ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS activity_id UUID REFERENCES activities(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS target_id UUID,
    ADD COLUMN IF NOT EXISTS target_type TEXT,
    ADD COLUMN IF NOT EXISTS questionnaire_id UUID REFERENCES questionnaires(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS algorithm TEXT,
    ADD COLUMN IF NOT EXISTS score NUMERIC(5, 2),
    ADD COLUMN IF NOT EXISTS detail_scores JSONB,
    ADD COLUMN IF NOT EXISTS reason TEXT,
    ADD COLUMN IF NOT EXISTS explanation JSONB,
    ADD COLUMN IF NOT EXISTS algorithm_version VARCHAR(32),
    ADD COLUMN IF NOT EXISTS status VARCHAR(32),
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;

UPDATE matches
SET target_id = activity_id
WHERE target_id IS NULL
  AND activity_id IS NOT NULL;

UPDATE matches
SET target_type = 'activity'
WHERE target_type IS NULL OR BTRIM(target_type) = '';

UPDATE matches
SET algorithm = 'rules'
WHERE algorithm IS NULL OR BTRIM(algorithm) = '';

UPDATE matches
SET score = 0
WHERE score IS NULL;

UPDATE matches
SET explanation = '{}'::JSONB
WHERE explanation IS NULL;

UPDATE matches
SET detail_scores = COALESCE(explanation->'detail_scores', '{}'::JSONB)
WHERE detail_scores IS NULL;

UPDATE matches
SET reason = COALESCE(explanation->>'reason', '')
WHERE reason IS NULL;

UPDATE matches
SET algorithm_version = 'activity-rules-v1'
WHERE algorithm_version IS NULL OR BTRIM(algorithm_version) = '';

UPDATE matches
SET status = 'recommended'
WHERE status IS NULL OR BTRIM(status) = '';

UPDATE matches
SET created_at = NOW()
WHERE created_at IS NULL;

UPDATE matches
SET updated_at = COALESCE(updated_at, created_at, NOW())
WHERE updated_at IS NULL;

WITH ranked_matches AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY user_id, activity_id, algorithm_version
               ORDER BY updated_at DESC NULLS LAST, created_at DESC NULLS LAST, id DESC
           ) AS duplicate_rank
    FROM matches
    WHERE user_id IS NOT NULL
      AND activity_id IS NOT NULL
      AND algorithm_version IS NOT NULL
)
UPDATE matches AS m
SET algorithm_version = 'legacy-' || LEFT(MD5(m.id::TEXT), 24)
FROM ranked_matches AS ranked
WHERE ranked.id = m.id
  AND ranked.duplicate_rank > 1;

ALTER TABLE matches
    ALTER COLUMN target_type SET DEFAULT 'activity',
    ALTER COLUMN algorithm SET DEFAULT 'rules',
    ALTER COLUMN score SET DEFAULT 0,
    ALTER COLUMN detail_scores SET DEFAULT '{}'::JSONB,
    ALTER COLUMN reason SET DEFAULT '',
    ALTER COLUMN explanation SET DEFAULT '{}'::JSONB,
    ALTER COLUMN algorithm_version SET DEFAULT 'activity-rules-v1',
    ALTER COLUMN status SET DEFAULT 'recommended',
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE matches
    ALTER COLUMN target_type SET NOT NULL,
    ALTER COLUMN algorithm SET NOT NULL,
    ALTER COLUMN score SET NOT NULL,
    ALTER COLUMN detail_scores SET NOT NULL,
    ALTER COLUMN reason SET NOT NULL,
    ALTER COLUMN explanation SET NOT NULL,
    ALTER COLUMN algorithm_version SET NOT NULL,
    ALTER COLUMN status SET NOT NULL,
    ALTER COLUMN created_at SET NOT NULL,
    ALTER COLUMN updated_at SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM matches WHERE user_id IS NULL) THEN
        ALTER TABLE matches ALTER COLUMN user_id SET NOT NULL;
    ELSE
        RAISE NOTICE 'matches.user_id contains NULL legacy rows; leaving column nullable until data is repaired';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM matches WHERE activity_id IS NULL) THEN
        ALTER TABLE matches ALTER COLUMN activity_id SET NOT NULL;
    ELSE
        RAISE NOTICE 'matches.activity_id contains NULL legacy rows; leaving column nullable until data is repaired';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM matches WHERE target_id IS NULL) THEN
        ALTER TABLE matches ALTER COLUMN target_id SET NOT NULL;
    ELSE
        RAISE NOTICE 'matches.target_id contains NULL legacy rows; leaving column nullable until data is repaired';
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS matches_user_score_idx
    ON matches (user_id, score DESC);

CREATE INDEX IF NOT EXISTS matches_user_updated_idx
    ON matches (user_id, updated_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS matches_user_activity_algorithm_uq
    ON matches (user_id, activity_id, algorithm_version);

CREATE INDEX IF NOT EXISTS matches_activity_idx
    ON matches (activity_id);

COMMIT;
