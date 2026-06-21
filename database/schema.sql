BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(80) NOT NULL DEFAULT '',
    role VARCHAR(32) NOT NULL DEFAULT 'user'
        CHECK (role IN ('user', 'admin')),
    school VARCHAR(120) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'disabled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_email_not_blank CHECK (BTRIM(email) <> '')
);

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS nickname VARCHAR(80) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS school VARCHAR(120) NOT NULL DEFAULT '';

UPDATE users SET nickname = '' WHERE nickname IS NULL;
UPDATE users SET school = '' WHERE school IS NULL;

ALTER TABLE users
    ALTER COLUMN nickname SET DEFAULT '',
    ALTER COLUMN nickname SET NOT NULL,
    ALTER COLUMN school SET DEFAULT '',
    ALTER COLUMN school SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_uq
    ON users (LOWER(email));

CREATE TABLE IF NOT EXISTS profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    display_name VARCHAR(80) NOT NULL,
    avatar_url TEXT,
    bio TEXT,
    city VARCHAR(80),
    gender VARCHAR(32),
    birth_date DATE,
    interests JSONB NOT NULL DEFAULT '[]'::JSONB,
    preferences JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT profiles_display_name_not_blank CHECK (BTRIM(display_name) <> '')
);

CREATE TABLE IF NOT EXISTS questionnaires (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    answers JSONB NOT NULL DEFAULT '{}'::JSONB,
    scores JSONB NOT NULL DEFAULT '{}'::JSONB,
    status VARCHAR(32) NOT NULL DEFAULT 'completed'
        CHECK (status IN ('draft', 'completed')),
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, version)
);

CREATE TABLE IF NOT EXISTS activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title VARCHAR(120) NOT NULL,
    type VARCHAR(64) NOT NULL DEFAULT 'project',
    description TEXT NOT NULL DEFAULT '',
    required_count INTEGER NOT NULL DEFAULT 2 CHECK (required_count > 0),
    joined_count INTEGER NOT NULL DEFAULT 0 CHECK (joined_count >= 0),
    tags JSONB NOT NULL DEFAULT '[]'::JSONB,
    preferred_tags JSONB NOT NULL DEFAULT '[]'::JSONB,
    time_text VARCHAR(120) NOT NULL DEFAULT '',
    location_text VARCHAR(160) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'recruiting'
        CHECK (status IN ('recruiting', 'full', 'closed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT activities_title_not_blank CHECK (BTRIM(title) <> '')
);

CREATE TABLE IF NOT EXISTS applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    applicant_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT NOT NULL DEFAULT '',
    match_score INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (activity_id, applicant_id)
);

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'activities' AND column_name = 'owner_id')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'activities' AND column_name = 'creator_id') THEN
        ALTER TABLE activities RENAME COLUMN owner_id TO creator_id;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'activities' AND column_name = 'category')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'activities' AND column_name = 'type') THEN
        ALTER TABLE activities RENAME COLUMN category TO type;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'activities' AND column_name = 'capacity')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'activities' AND column_name = 'required_count') THEN
        ALTER TABLE activities RENAME COLUMN capacity TO required_count;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'activities' AND column_name = 'location')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'activities' AND column_name = 'location_text') THEN
        ALTER TABLE activities RENAME COLUMN location TO location_text;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'applications' AND column_name = 'user_id')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'applications' AND column_name = 'applicant_id') THEN
        ALTER TABLE applications RENAME COLUMN user_id TO applicant_id;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'applications' AND column_name = 'message')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'applications' AND column_name = 'reason') THEN
        ALTER TABLE applications RENAME COLUMN message TO reason;
    END IF;
END $$;

ALTER TABLE activities DROP CONSTRAINT IF EXISTS activities_status_check;
ALTER TABLE activities DROP CONSTRAINT IF EXISTS activities_capacity_check;
ALTER TABLE activities DROP CONSTRAINT IF EXISTS activities_time_order;
ALTER TABLE applications DROP CONSTRAINT IF EXISTS applications_status_check;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'activities' AND column_name = 'starts_at') THEN
        ALTER TABLE activities ALTER COLUMN starts_at DROP NOT NULL;
    END IF;
END $$;

ALTER TABLE activities
    ADD COLUMN IF NOT EXISTS creator_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN IF NOT EXISTS type VARCHAR(64) NOT NULL DEFAULT 'project',
    ADD COLUMN IF NOT EXISTS required_count INTEGER NOT NULL DEFAULT 2,
    ADD COLUMN IF NOT EXISTS joined_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS tags JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN IF NOT EXISTS preferred_tags JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN IF NOT EXISTS time_text VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS location_text VARCHAR(160) NOT NULL DEFAULT '';

ALTER TABLE applications
    ADD COLUMN IF NOT EXISTS applicant_id UUID REFERENCES users(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS reason TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS match_score INTEGER NOT NULL DEFAULT 0;

UPDATE activities SET status = 'recruiting' WHERE status IN ('draft', 'published');
UPDATE activities SET status = 'closed' WHERE status IN ('cancelled', 'completed');
UPDATE applications SET status = 'approved' WHERE status = 'accepted';
UPDATE activities SET type = 'project' WHERE type IS NULL OR BTRIM(type) = '';
UPDATE activities SET description = '' WHERE description IS NULL;
UPDATE activities SET required_count = 2 WHERE required_count IS NULL OR required_count <= 0;
UPDATE activities SET joined_count = 0 WHERE joined_count IS NULL OR joined_count < 0;
UPDATE activities SET tags = '[]'::JSONB WHERE tags IS NULL;
UPDATE activities SET preferred_tags = '[]'::JSONB WHERE preferred_tags IS NULL;
UPDATE activities SET time_text = '' WHERE time_text IS NULL;
UPDATE activities SET location_text = '' WHERE location_text IS NULL;
UPDATE applications SET reason = '' WHERE reason IS NULL;
UPDATE applications SET match_score = 0 WHERE match_score IS NULL;

ALTER TABLE activities
    ALTER COLUMN creator_id SET NOT NULL,
    ALTER COLUMN type SET DEFAULT 'project',
    ALTER COLUMN type SET NOT NULL,
    ALTER COLUMN description SET DEFAULT '',
    ALTER COLUMN description SET NOT NULL,
    ALTER COLUMN required_count SET DEFAULT 2,
    ALTER COLUMN required_count SET NOT NULL,
    ALTER COLUMN joined_count SET DEFAULT 0,
    ALTER COLUMN joined_count SET NOT NULL,
    ALTER COLUMN tags SET DEFAULT '[]'::JSONB,
    ALTER COLUMN tags SET NOT NULL,
    ALTER COLUMN preferred_tags SET DEFAULT '[]'::JSONB,
    ALTER COLUMN preferred_tags SET NOT NULL,
    ALTER COLUMN time_text SET DEFAULT '',
    ALTER COLUMN time_text SET NOT NULL,
    ALTER COLUMN location_text SET DEFAULT '',
    ALTER COLUMN location_text SET NOT NULL,
    ALTER COLUMN status SET DEFAULT 'recruiting';

ALTER TABLE applications
    ALTER COLUMN applicant_id SET NOT NULL,
    ALTER COLUMN reason SET DEFAULT '',
    ALTER COLUMN reason SET NOT NULL,
    ALTER COLUMN match_score SET DEFAULT 0,
    ALTER COLUMN match_score SET NOT NULL,
    ALTER COLUMN status SET DEFAULT 'pending';

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'activities_status_check') THEN
        ALTER TABLE activities ADD CONSTRAINT activities_status_check CHECK (status IN ('recruiting', 'full', 'closed'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'activities_required_count_check') THEN
        ALTER TABLE activities ADD CONSTRAINT activities_required_count_check CHECK (required_count > 0);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'activities_joined_count_check') THEN
        ALTER TABLE activities ADD CONSTRAINT activities_joined_count_check CHECK (joined_count >= 0);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'applications_status_check') THEN
        ALTER TABLE applications ADD CONSTRAINT applications_status_check CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled'));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    questionnaire_id UUID REFERENCES questionnaires(id) ON DELETE SET NULL,
    score NUMERIC(5, 2) NOT NULL CHECK (score >= 0 AND score <= 100),
    explanation JSONB NOT NULL DEFAULT '{}'::JSONB,
    algorithm_version VARCHAR(32) NOT NULL DEFAULT 'v1',
    status VARCHAR(32) NOT NULL DEFAULT 'recommended'
        CHECK (status IN ('recommended', 'viewed', 'accepted', 'dismissed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, activity_id, algorithm_version)
);

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    event_type VARCHAR(80) NOT NULL,
    entity_type VARCHAR(64),
    entity_id UUID,
    payload JSONB NOT NULL DEFAULT '{}'::JSONB,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT events_type_not_blank CHECK (BTRIM(event_type) <> '')
);

CREATE TABLE IF NOT EXISTS feedbacks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_id UUID REFERENCES activities(id) ON DELETE CASCADE,
    match_id UUID REFERENCES matches(id) ON DELETE CASCADE,
    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT feedbacks_target_required CHECK (activity_id IS NOT NULL OR match_id IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS questionnaires_user_created_idx
    ON questionnaires (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS activities_status_starts_idx
    ON activities (status, created_at DESC);
CREATE INDEX IF NOT EXISTS activities_owner_idx
    ON activities (creator_id, created_at DESC);
CREATE INDEX IF NOT EXISTS applications_user_status_idx
    ON applications (applicant_id, status);
CREATE INDEX IF NOT EXISTS applications_activity_status_idx
    ON applications (activity_id, status);
CREATE UNIQUE INDEX IF NOT EXISTS applications_activity_applicant_uq
    ON applications (activity_id, applicant_id);
CREATE INDEX IF NOT EXISTS matches_user_score_idx
    ON matches (user_id, score DESC);
CREATE INDEX IF NOT EXISTS matches_activity_idx
    ON matches (activity_id);
CREATE INDEX IF NOT EXISTS events_type_occurred_idx
    ON events (event_type, occurred_at DESC);
CREATE INDEX IF NOT EXISTS events_user_occurred_idx
    ON events (user_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS feedbacks_user_created_idx
    ON feedbacks (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS feedbacks_activity_idx
    ON feedbacks (activity_id) WHERE activity_id IS NOT NULL;

COMMIT;
