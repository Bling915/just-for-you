#!/usr/bin/env bash
set -euo pipefail

backend_dir="$(cd "$(dirname "$0")/.." && pwd -P)"
admin_user="${MYSQL_ADMIN_USER:-root}"
admin_password="${MYSQL_ADMIN_PASSWORD:-}"
mysql_host="${MYSQL_HOST:-127.0.0.1}"
mysql_port="${MYSQL_PORT:-3306}"

for command_name in mysql mysqladmin go; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    printf '缺少命令 %s，请先安装 MySQL 8.x 和 Go。\n' "$command_name" >&2
    exit 1
  fi
done

export MYSQL_PWD="$admin_password"
connection=(--protocol=tcp -h "$mysql_host" -P "$mysql_port" -u "$admin_user")
if ! mysqladmin "${connection[@]}" ping --silent >/dev/null 2>&1; then
  printf '无法连接本地 MySQL，请先启动 MySQL 服务，并按需设置 MYSQL_ADMIN_USER/MYSQL_ADMIN_PASSWORD。\n' >&2
  exit 1
fi

mysql "${connection[@]}" <<'SQL'
CREATE DATABASE IF NOT EXISTS just_for_you CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'just_for_you'@'127.0.0.1' IDENTIFIED BY 'local-development';
CREATE USER IF NOT EXISTS 'just_for_you'@'localhost' IDENTIFIED BY 'local-development';
GRANT ALL PRIVILEGES ON just_for_you.* TO 'just_for_you'@'127.0.0.1';
GRANT ALL PRIVILEGES ON just_for_you.* TO 'just_for_you'@'localhost';
FLUSH PRIVILEGES;
SQL
unset MYSQL_PWD

export MYSQL_DSN="just_for_you:local-development@tcp(${mysql_host}:${mysql_port})/just_for_you?parseTime=true&charset=utf8mb4&multiStatements=true"
cd "$backend_dir"
for version in 1 2 3; do
  go run ./cmd/migrate -direction up -version "$version"
done
printf '本地数据库和 3 个迁移已准备完成。\n'
