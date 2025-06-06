CREATE TABLE `t_greeter`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
    `created_at`  datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`  datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `created_by`  int unsigned NOT NULL DEFAULT 0 COMMENT '创建人',
    `updated_by`  int unsigned NOT NULL DEFAULT 0 COMMENT '更新人',
    `age`  int unsigned NOT NULL DEFAULT 0 COMMENT '年龄',
    `name` varchar(16) NOT NULL COMMENT '姓名',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_bin COMMENT = '问候者表';