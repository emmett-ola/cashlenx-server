USE `cashlenx`;

-- -------------------
-- Create table `users`
-- -------------------
DROP TABLE IF EXISTS users;
CREATE TABLE `users`
(
    `id`           VARCHAR(24)  NOT NULL,
    `username`     VARCHAR(100) NOT NULL,
    `password_hash` VARCHAR(255) NOT NULL,
    `created_at`   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `updated_at`   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    `is_active`    BOOLEAN      NOT NULL DEFAULT TRUE,
    `role`         VARCHAR(20)  NOT NULL DEFAULT 'user' COMMENT 'user/admin',
    `is_external`  BOOLEAN      NOT NULL DEFAULT FALSE COMMENT 'true if user is from external auth system',
    `external_id`  VARCHAR(100) NULL COMMENT 'ID from external auth system',
    `password_set` BOOLEAN      NOT NULL DEFAULT FALSE COMMENT 'true if user has set their password',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = UTF8MB4
    COMMENT ='User Table';

CREATE UNIQUE INDEX users_username_unique_index ON users (username);
