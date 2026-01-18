USE `cashlenx`;

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
    `is_external`    BOOLEAN      NOT NULL DEFAULT FALSE COMMENT 'true if user is from external auth system',
    `external_id`    VARCHAR(100) NULL COMMENT 'ID from external auth system',
    `password_set`   BOOLEAN      NOT NULL DEFAULT FALSE COMMENT 'true if user has set their password',
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
