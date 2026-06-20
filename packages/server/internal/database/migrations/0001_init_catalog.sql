-- Labzyn catalog schema — metadata-driven model (philosophy from Twenty, in Postgres).
-- core:     tenants, users, memberships, record storage (MVP = JSONB).
-- metadata: object / field / view catalog describing the dynamic schema.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE SCHEMA IF NOT EXISTS core;
CREATE SCHEMA IF NOT EXISTS metadata;

-- ── Tenants ────────────────────────────────────────────────────────────────
CREATE TABLE core.workspace (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    display_name text        NOT NULL,
    subdomain    text        NOT NULL UNIQUE,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);

-- ── Users & memberships ──────────────────────────────────────────────────────
CREATE TABLE core."user" (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         text        NOT NULL UNIQUE,
    first_name    text        NOT NULL DEFAULT '',
    last_name     text        NOT NULL DEFAULT '',
    password_hash text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    deleted_at    timestamptz
);

CREATE TABLE core.user_workspace (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      uuid        NOT NULL REFERENCES core."user"(id)      ON DELETE CASCADE,
    workspace_id uuid        NOT NULL REFERENCES core.workspace(id)   ON DELETE CASCADE,
    role         text        NOT NULL DEFAULT 'member',
    created_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, workspace_id)
);

-- ── Object metadata (logical "tables") ───────────────────────────────────────
CREATE TABLE metadata.object (
    id             uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    workspace_id   uuid        NOT NULL REFERENCES core.workspace(id) ON DELETE CASCADE,
    name_singular  text        NOT NULL,
    name_plural    text        NOT NULL,
    label_singular text        NOT NULL,
    label_plural   text        NOT NULL,
    labels         jsonb       NOT NULL DEFAULT '{}',  -- i18n: {"fr":{...},"en":{...}}
    icon           text        NOT NULL DEFAULT 'IconList',
    description    text        NOT NULL DEFAULT '',
    is_system      boolean     NOT NULL DEFAULT false,
    is_active      boolean     NOT NULL DEFAULT true,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, name_singular),
    UNIQUE (workspace_id, name_plural)
);

-- ── Field metadata ───────────────────────────────────────────────────────────
CREATE TABLE metadata.field (
    id            uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    object_id     uuid        NOT NULL REFERENCES metadata.object(id) ON DELETE CASCADE,
    name          text        NOT NULL,
    type          text        NOT NULL,                -- FieldType (see internal/metadata)
    label         text        NOT NULL,
    labels        jsonb       NOT NULL DEFAULT '{}',
    icon          text        NOT NULL DEFAULT '',
    description   text        NOT NULL DEFAULT '',
    is_nullable   boolean     NOT NULL DEFAULT true,
    is_unique     boolean     NOT NULL DEFAULT false,
    is_system     boolean     NOT NULL DEFAULT false,
    is_active     boolean     NOT NULL DEFAULT true,
    default_value jsonb,
    options       jsonb,                                -- SELECT options, RELATION settings…
    position      integer     NOT NULL DEFAULT 0,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (object_id, name)
);

-- ── Views (table / board) ────────────────────────────────────────────────────
CREATE TABLE metadata.view (
    id         uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    object_id  uuid        NOT NULL REFERENCES metadata.object(id) ON DELETE CASCADE,
    name       text        NOT NULL,
    kind       text        NOT NULL DEFAULT 'table',
    filters    jsonb       NOT NULL DEFAULT '[]',
    sorts      jsonb       NOT NULL DEFAULT '[]',
    position   integer     NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- ── Record storage (MVP: JSONB; v2: per-workspace real tables) ────────────────
CREATE TABLE core.record (
    id           uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    object_id    uuid        NOT NULL REFERENCES metadata.object(id) ON DELETE CASCADE,
    workspace_id uuid        NOT NULL REFERENCES core.workspace(id)  ON DELETE CASCADE,
    data         jsonb       NOT NULL DEFAULT '{}',
    position     numeric     NOT NULL DEFAULT 0,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);

CREATE INDEX record_object_idx    ON core.record (object_id) WHERE deleted_at IS NULL;
CREATE INDEX record_workspace_idx ON core.record (workspace_id);
CREATE INDEX record_data_gin      ON core.record USING gin (data jsonb_path_ops);
