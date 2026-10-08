#!/usr/bin/env bash
set -Eeuo pipefail
mkdir -p results/logs
app_pid=""
stop_app() {
  if [[ -n "$app_pid" ]]; then
    kill "$app_pid" 2>/dev/null || true
    wait "$app_pid" 2>/dev/null || true
    app_pid=""
  fi
  docker rm -f bench-nginx bench-fpm >/dev/null 2>&1 || true
  sleep 2
}
trap stop_app EXIT
wait_http() {
  for _ in $(seq 1 60); do
    if curl --silent --fail --max-time 2 http://127.0.0.1:8080/health >/dev/null; then return 0; fi
    sleep 0.5
  done
  echo "Server never became healthy" >&2
  return 1
}
for language in go rust node php; do
  echo "=== Starting $language ==="
  stop_app
  case "$language" in
    go)
      bin/go-server >"results/logs/go.log" 2>&1 & app_pid=$!
      ;;
    rust)
      rust/target/release/benchmark-rust >"results/logs/rust.log" 2>&1 & app_pid=$!
      ;;
    node)
      node node/server.js >"results/logs/node.log" 2>&1 & app_pid=$!
      ;;
    php)
      docker run -d --rm --name bench-fpm --network host \
        -e DB_HOST -e DB_USER -e DB_PASSWORD local/php-mariadb-benchmark
      docker run -d --rm --name bench-nginx --network host \
        -v "$PWD/php/nginx.conf:/etc/nginx/conf.d/default.conf:ro" nginx:stable
      ;;
  esac
  wait_http
  # Distinct IDs including the highest row; incorrect routing / missing joins aborts.
  for id in 1 500000 1000000; do
    curl --fail --silent --show-error --max-time 10 "http://127.0.0.1:8080/lookup?id=$id" >"results/logs/$language-smoke-$id.json"
    python3 - "$id" "results/logs/$language-smoke-$id.json" <<'PY'
import json,sys
id=int(sys.argv[1])
obj=json.load(open(sys.argv[2]))
assert obj['id']==id and obj['username']==f'user_{id}'
assert obj['balance_cents']==id*37%100000 and obj['bio'] and obj['amount_cents']>=100
PY
  done
  status=$(curl -s -o /dev/null -w '%{http_code}' 'http://127.0.0.1:8080/lookup?id=0')
  [[ "$status" == "400" ]] || { echo "Expected invalid id 400, got $status"; exit 1; }
  for c in 1 16 64 128; do
    echo "=== $language concurrency=$c warming up ==="
    bin/load -backend "$language" -concurrency "$c" -duration 3s > /dev/null
    echo "=== $language concurrency=$c measuring ==="
    bin/load -backend "$language" -concurrency "$c" -duration 10s -output "results/$language-$c.json"
    sleep 1
  done
done
