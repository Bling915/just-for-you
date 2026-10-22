#!/usr/bin/env bash
set -euo pipefail

backend_dir="$(cd "$(dirname "$0")/.." && pwd -P)"
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT
mkdir -p "$tmp_dir/bin"

fail() { printf 'FAIL: %s\n' "$1" >&2; exit 1; }
assert_contains() { case "$1" in *"$2"*) ;; *) fail "缺少文本：$2" ;; esac; }

# dev.sh 必须给缺失配置一个可操作的中文提示。
if ENV_FILE="$tmp_dir/missing.env" "$backend_dir/scripts/dev.sh" >"$tmp_dir/missing.log" 2>&1; then
  fail '缺少 .env 时不应启动成功'
fi
assert_contains "$(cat "$tmp_dir/missing.log")" '请先创建后端配置文件'

cat >"$tmp_dir/dev.env" <<'EOF'
HTTP_ADDR=:18080
MYSQL_DSN=user:secret-dsn@tcp(127.0.0.1:3306)/db
JWT_SECRET=this-is-a-test-secret-longer-than-32-bytes
WECHAT_APP_ID=test-app
WECHAT_APP_SECRET=test-secret
COS_BUCKET_URL=https://example.invalid
COS_SECRET_ID=test-id
COS_SECRET_KEY=test-key
ENABLE_SOLO_TEST="true"
EOF
cat >"$tmp_dir/bin/go" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$PWD" >"$CALL_LOG.cwd"
printf '%s\n' "$*" >"$CALL_LOG.args"
printf '%s\n' "$ENABLE_SOLO_TEST" >"$CALL_LOG.bool"
EOF
chmod +x "$tmp_dir/bin/go"
dev_output="$(cd / && CALL_LOG="$tmp_dir/dev" ENV_FILE="$tmp_dir/dev.env" PATH="$tmp_dir/bin:$PATH" "$backend_dir/scripts/dev.sh")"
[[ "$(cat "$tmp_dir/dev.cwd")" == "$backend_dir" ]] || fail 'dev.sh 未切换到 backend 根目录'
[[ "$(cat "$tmp_dir/dev.args")" == 'run ./cmd/api' ]] || fail 'dev.sh 启动参数错误'
[[ "$(cat "$tmp_dir/dev.bool")" == 'true' ]] || fail 'dev.sh 没有去掉 dotenv 值外层引号'
[[ "$dev_output" != *'secret-dsn'* && "$dev_output" != *'test-key'* ]] || fail 'dev.sh 泄露了配置值'

# setup-local-db.sh 必须探活、建库，并依次执行 1/2/3 号迁移。
cat >"$tmp_dir/bin/mysqladmin" <<'EOF'
#!/usr/bin/env bash
printf 'ping\n' >>"$CALL_LOG"
EOF
cat >"$tmp_dir/bin/mysql" <<'EOF'
#!/usr/bin/env bash
printf 'mysql %s\n' "$*" >>"$CALL_LOG"
cat >/dev/null
EOF
cat >"$tmp_dir/bin/go" <<'EOF'
#!/usr/bin/env bash
printf 'go %s | cwd=%s | dsn=%s\n' "$*" "$PWD" "$MYSQL_DSN" >>"$CALL_LOG"
if [[ "${FAIL_MIGRATION_VERSION:-}" == "${*: -1}" ]]; then
  exit 9
fi
EOF
chmod +x "$tmp_dir/bin/mysqladmin" "$tmp_dir/bin/mysql" "$tmp_dir/bin/go"
: >"$tmp_dir/setup.log"
setup_output="$(cd / && CALL_LOG="$tmp_dir/setup.log" PATH="$tmp_dir/bin:$PATH" "$backend_dir/scripts/setup-local-db.sh")"
calls="$(cat "$tmp_dir/setup.log")"
assert_contains "$calls" 'ping'
assert_contains "$calls" 'go run ./cmd/migrate -direction up -version 1'
assert_contains "$calls" 'go run ./cmd/migrate -direction up -version 2'
assert_contains "$calls" 'go run ./cmd/migrate -direction up -version 3'
assert_contains "$calls" "cwd=$backend_dir"
[[ "$setup_output" != *'local-development'* ]] || fail 'setup-local-db.sh 输出了数据库密码'

: >"$tmp_dir/custom-setup.log"
CALL_LOG="$tmp_dir/custom-setup.log" MYSQL_HOST=192.0.2.10 MYSQL_PORT=3307 PATH="$tmp_dir/bin:$PATH" \
  "$backend_dir/scripts/setup-local-db.sh" >/dev/null
custom_calls="$(cat "$tmp_dir/custom-setup.log")"
assert_contains "$custom_calls" 'dsn=just_for_you:local-development@tcp(192.0.2.10:3307)/just_for_you'

: >"$tmp_dir/failing-setup.log"
if CALL_LOG="$tmp_dir/failing-setup.log" FAIL_MIGRATION_VERSION=2 PATH="$tmp_dir/bin:$PATH" \
  "$backend_dir/scripts/setup-local-db.sh" >/dev/null 2>&1; then
  fail '迁移失败时 setup-local-db.sh 不应成功'
fi
failing_calls="$(cat "$tmp_dir/failing-setup.log")"
assert_contains "$failing_calls" 'version 2'
[[ "$failing_calls" != *'version 3'* ]] || fail '迁移 2 失败后仍执行了迁移 3'

printf 'scripts_test: PASS\n'
