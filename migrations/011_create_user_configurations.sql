USE `cashlenx`;

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
