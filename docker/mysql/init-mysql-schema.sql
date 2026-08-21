-- MySQL initialization script for CashLenX - SCHEMA ONLY
-- This script creates tables with basic default categories
-- Demo/test data is available in init-mysql-demo.sql (import manually via CLI: cashlenx manage import)

USE cashlenx;

-- -------------------
-- Create table `users`
-- -------------------
DROP TABLE IF EXISTS users;
CREATE TABLE `users`
(
    `id`             VARCHAR(24)  NOT NULL,
    `username`       VARCHAR(100) NOT NULL,
    `password_hash`  VARCHAR(255) NOT NULL,
    `is_active`      BOOLEAN      NOT NULL DEFAULT TRUE,
    `role`           VARCHAR(20)  NOT NULL DEFAULT 'user' COMMENT 'user/admin',
    `nickname`       VARCHAR(100) NULL COMMENT 'User display name',
    `avatar_url`     VARCHAR(500) NULL COMMENT 'User avatar URL',
    `email_address`  VARCHAR(255) NULL COMMENT 'User email address',
    `gender`         VARCHAR(20)  NULL COMMENT 'User gender',
    `phone_number`   VARCHAR(32)  NULL COMMENT 'Optional phone number',
    `location`       VARCHAR(200) NULL COMMENT 'Optional location',
    `birth_date`     CHAR(10)     NULL COMMENT 'Optional YYYY-MM-DD birth date',
    `create_user_id` VARCHAR(24)  NOT NULL,
    `create_time`    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `update_user_id` VARCHAR(24)  NOT NULL,
    `update_time`    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    `delete_user_id` VARCHAR(24)           DEFAULT NULL,
    `delete_time`    TIMESTAMP             DEFAULT NULL,
    `is_delete`      BOOLEAN      NOT NULL DEFAULT FALSE,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = UTF8MB4 COMMENT ='User Table';

CREATE UNIQUE INDEX users_username_unique_index ON users (username);

-- -------------------
-- Create table `user_configurations`
-- -------------------
DROP TABLE IF EXISTS user_configurations;
CREATE TABLE `user_configurations`
(
    `id`                 VARCHAR(24) NOT NULL,
    `belongs_user_id`    VARCHAR(24) NOT NULL,
    `display_language`   VARCHAR(20) NOT NULL DEFAULT 'en',
    `currency_code`      CHAR(3)     NOT NULL DEFAULT 'USD',
    `active_theme_color` VARCHAR(50) NOT NULL DEFAULT '#2563eb',
    `create_user_id`     VARCHAR(24) NOT NULL,
    `create_time`        TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `update_user_id`     VARCHAR(24) NOT NULL,
    `update_time`        TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    `delete_user_id`     VARCHAR(24)          DEFAULT NULL,
    `delete_time`        TIMESTAMP            DEFAULT NULL,
    `is_delete`          BOOLEAN     NOT NULL DEFAULT FALSE,
    PRIMARY KEY (`id`),
    UNIQUE INDEX user_configurations_belongs_user_id_unique_index (`belongs_user_id`),
    CONSTRAINT user_configurations_user_id_fk FOREIGN KEY (`belongs_user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = UTF8MB4 COMMENT ='User Configuration Table';

-- -------------------
-- Create table `refresh_tokens`
-- -------------------
DROP TABLE IF EXISTS refresh_tokens;
CREATE TABLE `refresh_tokens`
(
    `id`             VARCHAR(24)  NOT NULL,
    `user_id`        VARCHAR(24)  NULL,
    `token`          VARCHAR(255) NOT NULL,
    `verification_token` VARCHAR(255) NULL,
    `expires_at`     TIMESTAMP    NOT NULL,
    `revoked_at`     TIMESTAMP    NULL,
    `revoked_by`     VARCHAR(24)  NULL,
    `device_id`      VARCHAR(100) NULL,
    `device_name`    VARCHAR(100) NULL,
    `ip_address`     VARCHAR(45)  NULL,
    `user_agent`     VARCHAR(255) NULL,
    `create_user_id` VARCHAR(24)  NULL,
    `create_time`    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `update_user_id` VARCHAR(24)  NULL,
    `update_time`    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    `delete_user_id` VARCHAR(24)           DEFAULT NULL,
    `delete_time`    TIMESTAMP             DEFAULT NULL,
    `is_delete`      BOOLEAN      NOT NULL DEFAULT FALSE,
    PRIMARY KEY (`id`),
    UNIQUE INDEX refresh_tokens_token_unique_index (`token`),
    INDEX refresh_tokens_user_id_index (`user_id`),
    INDEX refresh_tokens_expires_at_index (`expires_at`),
    CONSTRAINT refresh_tokens_user_id_fk FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = UTF8MB4 COMMENT ='Refresh Tokens';

-- -------------------
-- Create table `operation_confirm_codes`
-- -------------------
DROP TABLE IF EXISTS operation_confirm_codes;
CREATE TABLE `operation_confirm_codes`
(
    `id`             VARCHAR(24)  NOT NULL,
    `user_id`        VARCHAR(24)  NOT NULL,
    `token`          VARCHAR(255) NOT NULL,
    `operation_type` VARCHAR(50)  NOT NULL DEFAULT 'password_reset',
    `payload`        TEXT         NULL,
    `expires_at`     TIMESTAMP    NOT NULL,
    `used_at`        TIMESTAMP    NULL,
    `ip_address`     VARCHAR(45)  NULL,
    `create_user_id` VARCHAR(24)  NOT NULL,
    `create_time`    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `update_user_id` VARCHAR(24)  NOT NULL,
    `update_time`    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    `delete_user_id` VARCHAR(24)           DEFAULT NULL,
    `delete_time`    TIMESTAMP             DEFAULT NULL,
    `is_delete`      BOOLEAN      NOT NULL DEFAULT FALSE,
    PRIMARY KEY (`id`),
    INDEX operation_confirm_codes_token_index (`token`),
    UNIQUE INDEX operation_confirm_codes_verification_token_unique_index (`verification_token`),
    INDEX operation_confirm_codes_user_id_index (`user_id`),
    INDEX operation_confirm_codes_operation_type_index (`operation_type`),
    INDEX operation_confirm_codes_expires_at_index (`expires_at`)
) ENGINE = InnoDB DEFAULT CHARSET = UTF8MB4 COMMENT ='Operation Confirmation Codes';

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
-- Create table `budgets`
-- -------------------
DROP TABLE IF EXISTS budgets;
CREATE TABLE `budgets`
(
    `id`              VARCHAR(24)    NOT NULL,
    `belongs_user_id` VARCHAR(24)    NOT NULL,
    `category_id`     VARCHAR(24)    NOT NULL,
    `period`          CHAR(7)        NOT NULL,
    `limit_amount`    DECIMAL(18, 2) NOT NULL,
    `create_user_id`  VARCHAR(24)    NOT NULL,
    `create_time`     TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `update_user_id`  VARCHAR(24)    NOT NULL,
    `update_time`     TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    `delete_user_id`  VARCHAR(24)             DEFAULT NULL,
    `delete_time`     TIMESTAMP               DEFAULT NULL,
    `is_delete`       BOOLEAN        NOT NULL DEFAULT FALSE,
    `active_scope_key` VARCHAR(80) GENERATED ALWAYS AS (
        IF(`is_delete` = FALSE, CONCAT(`belongs_user_id`, '|', `period`, '|', `category_id`), NULL)
    ) STORED,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `budgets_active_scope_unique_index` (`active_scope_key`),
    INDEX `budgets_user_period_active_index` (`belongs_user_id`, `period`, `is_delete`),
    INDEX `budgets_user_category_index` (`belongs_user_id`, `category_id`)
) ENGINE = InnoDB DEFAULT CHARSET = UTF8MB4 COMMENT ='Monthly Budget Table';

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
    `amount`          DECIMAL(18, 2) NOT NULL,
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

-- Note: Default categories are automatically created for each user when they register
-- See config/default_categories.json for the list of default categories

-- Print initialization summary
SELECT
    'Schema initialized successfully!' AS message,
    (SELECT COUNT(*) FROM users) AS users_count,
    (SELECT COUNT(*) FROM categories) AS default_categories,
    (SELECT COUNT(*) FROM cash_flows) AS initial_transactions,
    (SELECT COUNT(*) FROM budgets) AS budgets_count,
    (SELECT COUNT(*) FROM refresh_tokens) AS refresh_tokens_count,
    (SELECT COUNT(*) FROM operation_confirm_codes) AS operation_confirm_codes_count,
    NOW() AS initialized_at;
