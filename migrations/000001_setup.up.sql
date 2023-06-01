CREATE TABLE IF NOT EXISTS rules
(
    id                UUID        DEFAULT gen_random_uuid()
        CONSTRAINT rules_pk
            PRIMARY KEY,
    name              TEXT                      NOT NULL,
    created_at        TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at        TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    user_id           TEXT                      NOT NULL,
    definition        TEXT                      NOT NULL,
    compiled          BYTEA                     NOT NULL,
    is_active         BOOL                      NOT NULL,
    included_entities TEXT[],
    excluded_entities TEXT[]
);

CREATE INDEX IF NOT EXISTS rules_user_id_is_active_index
    ON rules (user_id, is_active);
CREATE INDEX IF NOT EXISTS rules_included_idx ON rules USING gin(included_entities);
CREATE INDEX IF NOT EXISTS rules_excluded_idx ON rules USING gin(excluded_entities);
