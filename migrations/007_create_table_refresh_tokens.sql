USE `cashlenx`;

-- -------------------
-- Create table `refresh_tokens`
-- -------------------
DROP TABLE IF EXISTS refresh_tokens;
CREATE TABLE `refresh_tokens`
(
    `id`             VARCHAR(24)  NOT NULL,
    `user_id`        VARCHAR(24)  NOT NULL,
    `token`          VARCHAR(255) NOT NULL,
    `expires_at`     TIMESTAMP    NOT NULL,
    `revoked_at`     TIMESTAMP    NULL,
    `revoked_by`     VARCHAR(24)  NULL,
    `device_id`      VARCHAR(100) NULL,
    `device_name`    VARCHAR(100) NULL,
    `ip_address`     VARCHAR(45)  NULL,
    `user_agent`     VARCHAR(255) NULL,
    `create_user_id` VARCHAR(24)  NOT NULL,
    `create_time`    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `update_user_id` VARCHAR(24)  NOT NULL,
    `update_time`    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    `delete_user_id` VARCHAR(24)           DEFAULT NULL,
    `delete_time`    TIMESTAMP             DEFAULT NULL,
    `is_delete`      BOOLEAN      NOT NULL DEFAULT FALSE,
    PRIMARY KEY (`id`),
    UNIQUE INDEX refresh_tokens_token_unique_index ON refresh_tokens (token),
    INDEX refresh_tokens_user_id_index ON refresh_tokens (user_id),
    INDEX refresh_tokens_expires_at_index ON refresh_tokens (expires_at),
    CONSTRAINT refresh_tokens_user_id_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = UTF8MB4
    COMMENT ='Refresh Tokens';
