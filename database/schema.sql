BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'user'
        CHECK (role IN ('user', 'admin')),
    status VARCHAR(32) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'disabled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_email_not_blank CHECK (BTRIM(email) <> '')
);

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
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title VARCHAR(120) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category VARCHAR(64) NOT NULL,
    city VARCHAR(80),
    location TEXT,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ,
    capacity INTEGER NOT NULL DEFAULT 2 CHECK (capacity > 0),
    status VARCHAR(32) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'published', 'closed', 'cancelled', 'completed')),
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT activities_title_not_blank CHECK (BTRIM(title) <> ''),
    CONSTRAINT activities_time_order CHECK (ends_at IS NULL OR ends_at > starts_at)
);

CREATE TABLE IF NOT EXISTS applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT,
    status VARCHAR(32) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'rejected', 'cancelled')),
    decided_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (activity_id, user_id)
);

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
    ON activities (status, starts_at);
CREATE INDEX IF NOT EXISTS activities_owner_idx
    ON activities (owner_id, created_at DESC);
CREATE INDEX IF NOT EXISTS applications_user_status_idx
    ON applications (user_id, status);
CREATE INDEX IF NOT EXISTS applications_activity_status_idx
    ON applications (activity_id, status);
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
