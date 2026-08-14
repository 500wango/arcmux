#!/usr/bin/env bash
# ArcMux one-click Docker deploy (PostgreSQL + Redis + app image built from source)
#
# Usage:
#   ./scripts/deploy-docker.sh              # generate env if needed, build, start
#   ./scripts/deploy-docker.sh --rebuild    # force image rebuild
#   ./scripts/deploy-docker.sh --status     # show compose status + health
#   ./scripts/deploy-docker.sh --logs       # follow app logs
#   ./scripts/deploy-docker.sh --down       # stop stack (keeps volumes)
#   ./scripts/deploy-docker.sh --env-file /path/to.env
#
# First visit: http://<server-ip>:<HOST_PORT>/  complete the setup wizard.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"
ENV_FILE="${ENV_FILE:-.env.prod}"
ENV_EXAMPLE="deploy/docker.env.example"
SERVICE_APP="arcmux"

ACTION="up"
REBUILD=0
NO_CACHE=0

usage() {
  sed -n '2,13p' "$0" | sed 's/^# \?//'
  exit 0
}

log() { printf '==> %s\n' "$*"; }
warn() { printf 'WARNING: %s\n' "$*" >&2; }
die() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }

rand_hex() {
  local bytes="${1:-32}"
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex "$bytes"
  else
    head -c "$bytes" /dev/urandom | od -An -tx1 | tr -d ' \n'
  fi
}

compose() {
  docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" "$@"
}

require_docker() {
  command -v docker >/dev/null 2>&1 || die "docker not found. Install Docker first: https://docs.docker.com/engine/install/"
  docker info >/dev/null 2>&1 || die "cannot talk to Docker daemon (permission or service not running). Try: sudo usermod -aG docker \$USER && newgrp docker"
  docker compose version >/dev/null 2>&1 || die "docker compose plugin not found. Install Docker Compose v2."
}

ensure_env_file() {
  if [[ -f "$ENV_FILE" ]]; then
    log "Using existing env file: $ENV_FILE"
    return
  fi

  [[ -f "$ENV_EXAMPLE" ]] || die "missing $ENV_EXAMPLE"

  local pg_pass redis_pass session_secret crypto_secret
  pg_pass="$(rand_hex 16)"
  redis_pass="$(rand_hex 16)"
  session_secret="$(rand_hex 32)"
  crypto_secret="$(rand_hex 32)"

  # shellcheck disable=SC2016
  sed \
    -e "s|^POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=${pg_pass}|" \
    -e "s|^REDIS_PASSWORD=.*|REDIS_PASSWORD=${redis_pass}|" \
    -e "s|^SESSION_SECRET=.*|SESSION_SECRET=${session_secret}|" \
    -e "s|^CRYPTO_SECRET=.*|CRYPTO_SECRET=${crypto_secret}|" \
    "$ENV_EXAMPLE" >"$ENV_FILE"

  chmod 600 "$ENV_FILE"
  log "Generated $ENV_FILE with random secrets (mode 600)"
  warn "Keep $ENV_FILE private. Losing it means losing DB/Redis/app secrets."
}

validate_env_file() {
  local required=(POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB REDIS_PASSWORD SESSION_SECRET CRYPTO_SECRET)
  local key missing=0
  for key in "${required[@]}"; do
    if ! grep -Eq "^${key}=.+" "$ENV_FILE"; then
      warn "missing or empty: $key in $ENV_FILE"
      missing=1
    fi
  done
  if grep -Eq 'CHANGE_ME_' "$ENV_FILE"; then
    die "$ENV_FILE still contains CHANGE_ME_ placeholders. Edit secrets or delete the file and re-run to auto-generate."
  fi
  [[ "$missing" -eq 0 ]] || die "env file incomplete: $ENV_FILE"
}

read_env_var() {
  local key="$1" line value
  line="$(grep -E "^${key}=" "$ENV_FILE" | tail -n1 || true)"
  value="${line#*=}"
  value="${value%$'\r'}"
  # strip optional surrounding quotes
  if [[ "$value" =~ ^\".*\"$ ]]; then
    value="${value:1:${#value}-2}"
  elif [[ "$value" =~ ^\'.*\'$ ]]; then
    value="${value:1:${#value}-2}"
  fi
  printf '%s' "$value"
}

ensure_dirs() {
  local data_dir log_dir
  data_dir="$(read_env_var DATA_DIR)"
  log_dir="$(read_env_var LOG_DIR)"
  data_dir="${data_dir:-./data}"
  log_dir="${log_dir:-./logs}"
  mkdir -p "$data_dir" "$log_dir"
}

wait_healthy() {
  local host_port tries=0 max_tries=60
  host_port="$(read_env_var HOST_PORT)"
  host_port="${host_port:-3000}"

  log "Waiting for API health on :${host_port} ..."
  while ((tries < max_tries)); do
    if curl -fsS "http://127.0.0.1:${host_port}/api/status" 2>/dev/null | grep -q '"success"[[:space:]]*:[[:space:]]*true'; then
      log "API is healthy"
      return 0
    fi
    # fall back to container health if curl fails (e.g. host bind not local)
    if compose ps --status running --format json 2>/dev/null | grep -q "\"Health\":\"healthy\""; then
      :
    fi
    tries=$((tries + 1))
    sleep 2
  done

  warn "Timed out waiting for /api/status. Recent logs:"
  compose logs --tail=80 "$SERVICE_APP" || true
  return 1
}

print_summary() {
  local host_port host_bind
  host_port="$(read_env_var HOST_PORT)"
  host_bind="$(read_env_var HOST_BIND)"
  host_port="${host_port:-3000}"
  host_bind="${host_bind:-0.0.0.0}"

  local ip
  ip="$(hostname -I 2>/dev/null | awk '{print $1}')"
  ip="${ip:-<server-ip>}"

  cat <<EOF

------------------------------------------------------------
 ArcMux deploy finished
------------------------------------------------------------
 Compose : $COMPOSE_FILE
 Env     : $ENV_FILE
 Bind    : ${host_bind}:${host_port}

 Open:
   http://127.0.0.1:${host_port}/
   http://${ip}:${host_port}/

 First run: complete the setup wizard and create the admin user.

 Useful commands:
   $0 --status
   $0 --logs
   $0 --rebuild
   $0 --down

 HTTPS tip (after reverse proxy):
   set SESSION_COOKIE_SECURE=true
   set SESSION_COOKIE_TRUSTED_URL=https://your-dashboard-domain[,https://api-domain]
   IMPORTANT: list the browser-facing dashboard Origin(s) too; without them
   the dashboard logs you out ~15 minutes after login (403 AUTH_ORIGIN_FORBIDDEN).
   then re-run: $0 --rebuild
------------------------------------------------------------
EOF
}

do_up() {
  require_docker
  [[ -f "$COMPOSE_FILE" ]] || die "missing $COMPOSE_FILE"
  ensure_env_file
  validate_env_file
  ensure_dirs

  if [[ "$REBUILD" -eq 1 ]]; then
    log "Building images..."
    if [[ "$NO_CACHE" -eq 1 ]]; then
      compose build --no-cache
    else
      compose build
    fi
  fi

  log "Starting stack..."
  if [[ "$REBUILD" -eq 1 ]]; then
    compose up -d --build
  else
    # First deploy should build; later runs reuse image unless --rebuild
    compose up -d --build
  fi

  compose ps
  wait_healthy || true
  print_summary
}

do_status() {
  require_docker
  [[ -f "$ENV_FILE" ]] || die "missing $ENV_FILE (run deploy first)"
  compose ps
  echo
  local host_port
  host_port="$(read_env_var HOST_PORT)"
  host_port="${host_port:-3000}"
  if curl -fsS "http://127.0.0.1:${host_port}/api/status" >/tmp/arcmux-status.json 2>/dev/null; then
    log "API /api/status OK"
    head -c 400 /tmp/arcmux-status.json
    echo
  else
    warn "API /api/status not reachable on 127.0.0.1:${host_port}"
  fi
}

do_logs() {
  require_docker
  [[ -f "$ENV_FILE" ]] || die "missing $ENV_FILE"
  compose logs -f --tail=200 "$SERVICE_APP"
}

do_down() {
  require_docker
  [[ -f "$ENV_FILE" ]] || die "missing $ENV_FILE"
  log "Stopping stack (volumes kept)..."
  compose down
  log "Stopped. Data volumes retained (pg_data / redis_data / data dir)."
}

# ---- args ----
while [[ $# -gt 0 ]]; do
  case "$1" in
    -h | --help) usage ;;
    --rebuild) REBUILD=1; shift ;;
    --no-cache) NO_CACHE=1; REBUILD=1; shift ;;
    --status) ACTION="status"; shift ;;
    --logs) ACTION="logs"; shift ;;
    --down) ACTION="down"; shift ;;
    --env-file)
      [[ $# -ge 2 ]] || die "--env-file needs a path"
      ENV_FILE="$2"
      shift 2
      ;;
    --compose-file)
      [[ $# -ge 2 ]] || die "--compose-file needs a path"
      COMPOSE_FILE="$2"
      shift 2
      ;;
    *)
      die "unknown argument: $1 (try --help)"
      ;;
  esac
done

case "$ACTION" in
  up) do_up ;;
  status) do_status ;;
  logs) do_logs ;;
  down) do_down ;;
  *) die "unknown action: $ACTION" ;;
esac
