USE `cashlenx`;

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
