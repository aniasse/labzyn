-- Per-object storage mode: 'jsonb' (default, core.record) or 'table'
-- (a real per-workspace table with typed columns — the high-performance path).
ALTER TABLE metadata.object
    ADD COLUMN storage text NOT NULL DEFAULT 'jsonb';
