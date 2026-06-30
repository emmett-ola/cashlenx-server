ALTER TABLE categories
    ADD COLUMN active_scope_key VARCHAR(300)
    GENERATED ALWAYS AS (
        IF(is_delete = FALSE, CONCAT(belongs_user_id, '|', type, '|', COALESCE(parent_id, ''), '|', name), NULL)
    ) STORED;

CREATE UNIQUE INDEX categories_active_scope_unique_index ON categories (active_scope_key);
CREATE INDEX categories_user_scope_index ON categories (belongs_user_id, type, parent_id, name, is_delete);

DROP INDEX categories_belongs_user_id_name_unique_index ON categories;
