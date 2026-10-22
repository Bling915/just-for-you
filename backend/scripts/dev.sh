#!/usr/bin/env bash
set -euo pipefail

backend_dir="$(cd "$(dirname "$0")/.." && pwd -P)"
env_file="${ENV_FILE:-$backend_dir/.env}"

if [[ ! -f "$env_file" ]]; then
  printf '请先创建后端配置文件：复制 backend/.env.example 为 backend/.env，并填写本地配置。\n' >&2
  exit 1
fi
if ! command -v go >/dev/null 2>&1; then
  printf '没有找到 Go，请先安装 Go 1.23 或更高版本。\n' >&2
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
exec go run ./cmd/api
