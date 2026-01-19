USE `cashlenx`;

-- -------------------
-- Create table `refresh_tokens`
-- -------------------
DROP TABLE IF EXISTS refresh_tokens;
CREATE TABLE `refresh_tokens`
(
    `id`              VARCHAR(24)  NOT NULL,
    `belongs_user_id` VARCHAR(24)  NOT NULL,
    `token`           VARCHAR(255) NOT NULL,
    `expires_at`      TIMESTAMP    NOT NULL,
    `created_at`      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `revoked_at`      TIMESTAMP    NULL,
    `revoked_by`      VARCHAR(24)  NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX refresh_tokens_token_unique_index ON refresh_tokens (token),
    INDEX refresh_tokens_belongs_user_id_index ON refresh_tokens (belongs_user_id),
    INDEX refresh_tokens_expires_at_index ON refresh_tokens (expires_at),
    CONSTRAINT refresh_tokens_belongs_user_id_fk FOREIGN KEY (belongs_user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB
  DEFAULT CHARSET = UTF8MB4
    COMMENT ='Refresh Tokens';