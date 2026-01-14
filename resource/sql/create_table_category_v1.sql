USE
    `cashlenx`;

-- -------------------
-- Create table `categories`
-- -------------------
DROP TABLE IF EXISTS categories;
CREATE TABLE `categories`
(
    `id`          VARCHAR(24)  NOT NULL,
    `parent_id`   VARCHAR(24)           DEFAULT NULL,
    `name`        VARCHAR(200) NOT NULL,
    `type`        VARCHAR(10)   NOT NULL,
    `remark`      VARCHAR(200)          DEFAULT NULL,
    `create_time` TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP(),
    `modify_time` TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP() ON UPDATE CURRENT_TIMESTAMP(),
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = UTF8MB4
    COMMENT ='Category Table';

CREATE UNIQUE INDEX category_name_unique_index ON categories (name);
