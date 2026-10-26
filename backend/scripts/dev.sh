#!/usr/bin/env bash
set -euo pipefail

backend_dir="$(cd "$(dirname "$0")/.." && pwd -P)"
env_file="${ENV_FILE:-$backend_dir/.env}"

if [[ ! -f "$env_file" ]]; then
  printf '请先创建后端配置文件：复制 backend/.env.example 为 backend/.env，并填写本地配置。\n' >&2
  exit 1
fi

# 定位 Go：IDE 启动的进程常常继承不到 Homebrew 的 PATH，
# 所以 command -v 失败时再到常见安装目录里找。
go_bin=""
if command -v go >/dev/null 2>&1; then
  go_bin="$(command -v go)"
else
  for candidate in /opt/homebrew/bin/go /usr/local/bin/go /usr/local/go/bin/go "$HOME/go/bin/go" "$HOME/sdk"/go*/bin/go; do
    if [[ -x "$candidate" ]]; then
      go_bin="$candidate"
      break
    fi
  done
fi
if [[ -z "$go_bin" ]]; then
  printf '没有找到 Go，请先安装 Go 1.23 或更高版本，或在运行配置中补充 PATH。\n' >&2
  exit 1
fi

while IFS= read -r line || [[ -n "$line" ]]; do
  line="${line%$'\r'}"
  [[ -z "$line" || "$line" == \#* ]] && continue
  if [[ "$line" != *=* ]]; then
    printf '配置文件中存在无效行，请使用 KEY=value 格式。\n' >&2
    exit 1
  fi
  key="${line%%=*}"
  value="${line#*=}"
  if [[ ! "$key" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]]; then
    printf '配置文件中存在无效变量名：%s。\n' "$key" >&2
    exit 1
  fi
  if [[ ${#value} -ge 2 && ( ( "$value" == \"*\" ) || ( "$value" == \'*\' ) ) ]]; then
    value="${value:1:${#value}-2}"
  fi
  export "$key=$value"
done <"$env_file"

cd "$backend_dir"

# go run 会把真正的服务编译到临时目录再作为子进程运行，
# IDE 点「停止」时只杀掉 go run，编译出的服务会残留并占住端口，
# 下次启动就报 "address already in use"。这里统一登记并在退出时清理。
child_pid=""
cleanup() {
  if [[ -n "$child_pid" ]] && kill -0 "$child_pid" 2>/dev/null; then
    kill "$child_pid" 2>/dev/null || true
    wait "$child_pid" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

"$go_bin" run ./cmd/api &
child_pid=$!
wait "$child_pid"
