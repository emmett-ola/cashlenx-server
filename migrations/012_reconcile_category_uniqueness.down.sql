DROP INDEX categories_user_scope_index ON categories;
DROP INDEX categories_active_scope_unique_index ON categories;
ALTER TABLE categories DROP COLUMN active_scope_key;
CREATE UNIQUE INDEX categories_belongs_user_id_name_unique_index ON categories (belongs_user_id, name);
