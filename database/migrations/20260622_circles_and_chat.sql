BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS circles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    name VARCHAR(120) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category VARCHAR(64) NOT NULL DEFAULT 'general',
    status VARCHAR(32) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    member_count INTEGER NOT NULL DEFAULT 1 CHECK (member_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT circles_name_not_blank CHECK (BTRIM(name) <> '')
);

CREATE TABLE IF NOT EXISTS circle_members (
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL DEFAULT 'member' CHECK (role IN ('owner', 'member')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (circle_id, user_id)
);

CREATE TABLE IF NOT EXISTS circle_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    name VARCHAR(80) NOT NULL DEFAULT 'general',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT circle_channels_name_not_blank CHECK (BTRIM(name) <> ''),
    UNIQUE (circle_id, name)
);

CREATE TABLE IF NOT EXISTS circle_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    channel_id UUID NOT NULL REFERENCES circle_channels(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content VARCHAR(1000) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT circle_messages_content_length CHECK (CHAR_LENGTH(BTRIM(content)) BETWEEN 1 AND 1000)
);

CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind VARCHAR(32) NOT NULL DEFAULT 'direct' CHECK (kind = 'direct'),
    direct_key VARCHAR(73) NOT NULL,
    last_message_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT conversations_direct_key_not_blank CHECK (BTRIM(direct_key) <> '')
);

CREATE TABLE IF NOT EXISTS conversation_members (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_read_at TIMESTAMPTZ,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (conversation_id, user_id)
);

CREATE TABLE IF NOT EXISTS direct_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content VARCHAR(1000) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT direct_messages_content_length CHECK (CHAR_LENGTH(BTRIM(content)) BETWEEN 1 AND 1000)
);

CREATE INDEX IF NOT EXISTS circles_status_created_idx ON circles (status, created_at DESC);
CREATE INDEX IF NOT EXISTS circles_creator_created_idx ON circles (creator_id, created_at DESC);
CREATE INDEX IF NOT EXISTS circle_members_user_joined_idx ON circle_members (user_id, joined_at DESC);
CREATE INDEX IF NOT EXISTS circle_channels_circle_idx ON circle_channels (circle_id, created_at);
CREATE INDEX IF NOT EXISTS circle_messages_channel_created_idx ON circle_messages (channel_id, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS conversations_direct_key_uq ON conversations (direct_key);
CREATE INDEX IF NOT EXISTS conversation_members_user_idx ON conversation_members (user_id, joined_at DESC);
CREATE INDEX IF NOT EXISTS direct_messages_conversation_created_idx ON direct_messages (conversation_id, created_at DESC);

COMMIT;
