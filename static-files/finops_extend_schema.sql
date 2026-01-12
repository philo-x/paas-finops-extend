SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 数据库创建
-- ----------------------------
CREATE DATABASE IF NOT EXISTS `finops_extend` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE `finops_extend`;

-- ----------------------------
-- 管理员用户表
-- ----------------------------
DROP TABLE IF EXISTS `admin_user`;

CREATE TABLE `admin_user` (
  `admin_user_id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '管理员id',
  `login_user_name` varchar(50) NOT NULL COMMENT '管理员登陆名称',
  `login_password` varchar(50) NOT NULL COMMENT '管理员登陆密码',
  `nick_name` varchar(50) NOT NULL COMMENT '管理员显示昵称',
  `locked` tinyint(4) DEFAULT '0' COMMENT '是否锁定 0未锁定 1已锁定无法登陆',
  PRIMARY KEY (`admin_user_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 ROW_FORMAT=DYNAMIC COMMENT='管理员用户表';

-- 默认管理员账户 admin/123456
INSERT INTO `admin_user` (`admin_user_id`, `login_user_name`, `login_password`, `nick_name`, `locked`)
VALUES (1, 'admin', 'e10adc3949ba59abbe56e057f20f883e', '管理员', 0);

-- ----------------------------
-- 管理员Token表
-- ----------------------------
DROP TABLE IF EXISTS `admin_user_token`;

CREATE TABLE `admin_user_token` (
  `admin_user_id` bigint(20) NOT NULL COMMENT '用户主键id',
  `token` varchar(32) NOT NULL COMMENT 'token值(32位字符串)',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '修改时间',
  `expire_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'token过期时间',
  PRIMARY KEY (`admin_user_id`),
  UNIQUE KEY `uq_token` (`token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='管理员Token表';

-- ----------------------------
-- 告警信息表
-- ----------------------------
DROP TABLE IF EXISTS `prometheus_alert`;

CREATE TABLE `prometheus_alert` (
  `alert_id` int(11) NOT NULL AUTO_INCREMENT COMMENT '告警ID',
  `status` varchar(50) NOT NULL DEFAULT '' COMMENT '告警状态(firing/resolved)',
  `starts_at` datetime NOT NULL COMMENT '告警开始时间',
  `ends_at` datetime DEFAULT NULL COMMENT '告警结束时间',
  `annotations` json DEFAULT NULL COMMENT '告警注解(JSON格式)',
  `labels` json DEFAULT NULL COMMENT '告警标签(JSON格式)',
  `is_deleted` tinyint(4) NOT NULL DEFAULT '0' COMMENT '删除标识字段(0-未删除 1-已删除)',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`alert_id`) USING BTREE,
  KEY `idx_status` (`status`) USING BTREE,
  KEY `idx_starts_at` (`starts_at`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 ROW_FORMAT=DYNAMIC COMMENT='告警信息表';

-- ----------------------------
-- 文件上传表（用于文件上传功能）
-- ----------------------------
DROP TABLE IF EXISTS `exa_file_upload_and_downloads`;

CREATE TABLE `exa_file_upload_and_downloads` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` varchar(191) DEFAULT NULL COMMENT '文件名',
  `url` varchar(191) DEFAULT NULL COMMENT '文件地址',
  `tag` varchar(191) DEFAULT NULL COMMENT '文件标签',
  `key` varchar(191) DEFAULT NULL COMMENT '编号',
  PRIMARY KEY (`id`),
  KEY `idx_exa_file_upload_and_downloads_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件上传表';

SET FOREIGN_KEY_CHECKS = 1;
