-- MySQL initialization script for CashLenX - SCHEMA ONLY
-- This script creates tables with basic default categories
-- Demo/test data is available in init-mysql-demo.sql (import manually via CLI: cashlenx manage import)

USE cashlenx;

-- -------------------
-- Create table `categories`
-- -------------------
DROP TABLE IF EXISTS categories;
CREATE TABLE `categories`
(
    `id`              VARCHAR(24)  NOT NULL,
    `belongs_user_id` VARCHAR(24)  NOT NULL,
    `parent_id`       VARCHAR(24)           DEFAULT NULL,
    `name`            VARCHAR(200) NOT NULL,
    `type`            VARCHAR(10)   NOT NULL,
    `remark`          VARCHAR(200)          DEFAULT NULL,
    `create_user_id`  VARCHAR(24)  NOT NULL,
    `create_time`     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `update_user_id`  VARCHAR(24)  NOT NULL,
    `update_time`     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    `delete_user_id`  VARCHAR(24)           DEFAULT NULL,
    `delete_time`     TIMESTAMP             DEFAULT NULL,
    `is_delete`       BOOLEAN      NOT NULL DEFAULT FALSE,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = UTF8MB4 COMMENT ='Category Table';

CREATE INDEX categories_belongs_user_id_index ON categories (belongs_user_id);
CREATE INDEX categories_parent_id_index ON categories (parent_id);
CREATE UNIQUE INDEX categories_belongs_user_id_name_unique_index ON categories (belongs_user_id, name);

-- -------------------
-- Create table `cash_flows`
-- -------------------
DROP TABLE IF EXISTS cash_flows;
CREATE TABLE `cash_flows`
(
    `id`              VARCHAR(24)  NOT NULL,
    `belongs_user_id` VARCHAR(24)  NOT NULL,
    `category_id`     VARCHAR(24)  NOT NULL,
    `belongs_date`    TIMESTAMP    NOT NULL,
    `flow_type`       VARCHAR(10)  NOT NULL COMMENT 'INCOME/EXPENSE',
    `amount`          DECIMAL      NOT NULL,
    `description`     VARCHAR(200) NOT NULL,
    `remark`          VARCHAR(200)          DEFAULT NULL COMMENT 'KEEP EMPTY',
    `create_user_id`  VARCHAR(24)  NOT NULL,
    `create_time`     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `update_user_id`  VARCHAR(24)  NOT NULL,
    `update_time`     TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    `delete_user_id`  VARCHAR(24)           DEFAULT NULL,
    `delete_time`     TIMESTAMP             DEFAULT NULL,
    `is_delete`       BOOLEAN      NOT NULL DEFAULT FALSE,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = UTF8MB4 COMMENT ='Cash Flow Table';

CREATE INDEX cash_flows_belongs_user_id_index ON cash_flows (belongs_user_id);
CREATE INDEX cash_flows_category_id_index ON cash_flows (category_id);
CREATE INDEX cash_flows_belongs_date_index ON cash_flows (belongs_date);
CREATE INDEX cash_flows_flow_type_index ON cash_flows (flow_type);

-- Insert basic default categories (auto-loaded on init)
-- These categories are available for all users by default
-- Using '000000000000000000000000' as default system user ID
INSERT INTO categories (id, belongs_user_id, name, type, remark, create_user_id, update_user_id) VALUES
('000000000000000000000001', '000000000000000000000000', 'Salary', 'INCOME', 'Income from employment', '000000000000000000000000', '000000000000000000000000'),
('000000000000000000000002', '000000000000000000000000', 'Freelance', 'INCOME', 'Income from freelance work', '000000000000000000000000', '000000000000000000000000'),
('000000000000000000000003', '000000000000000000000000', 'Investment', 'INCOME', 'Income from investments and dividends', '000000000000000000000000', '000000000000000000000000'),
('000000000000000000000004', '000000000000000000000000', 'Other Income', 'INCOME', 'Other income sources', '000000000000000000000000', '000000000000000000000000'),
('000000000000000000000005', '000000000000000000000000', 'Food & Dining', 'EXPENSE', 'Restaurants, groceries, food delivery', '000000000000000000000000', '000000000000000000000000'),
('000000000000000000000006', '000000000000000000000000', 'Transportation', 'EXPENSE', 'Gas, public transport, car maintenance', '000000000000000000000000', '000000000000000000000000'),
('000000000000000000000007', '000000000000000000000000', 'Shopping', 'EXPENSE', 'Retail purchases, online shopping', '000000000000000000000000', '000000000000000000000000'),
('000000000000000000000008', '000000000000000000000000', 'Entertainment', 'EXPENSE', 'Movies, games, hobbies', '000000000000000000000000', '000000000000000000000000'),
('000000000000000000000009', '000000000000000000000000', 'Healthcare', 'EXPENSE', 'Medical expenses, pharmacy, fitness', '000000000000000000000000', '000000000000000000000000'),
('000000000000000000000010', '000000000000000000000000', 'Utilities', 'EXPENSE', 'Electricity, water, internet, phone', '000000000000000000000000', '000000000000000000000000');

-- Print initialization summary
SELECT 
    'Schema initialized successfully!' AS message,
    (SELECT COUNT(*) FROM categories) AS default_categories,
    (SELECT COUNT(*) FROM cash_flows) AS initial_transactions,
    NOW() AS initialized_at;
