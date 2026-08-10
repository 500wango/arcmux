#!/usr/bin/env bash
# ==============================================================================
# ArcMux VPS One-Click Deployment Script
# Turnkey installer: Auto-installs Docker, configures secrets, builds & runs ArcMux
#
# Quick Install:
#   bash <(curl -fsSL https://raw.githubusercontent.com/500wango/arcmux/main/scripts/install.sh)
#
# Local Run:
#   ./scripts/install.sh
# ==============================================================================

set -euo pipefail

# ----------------- Colors & Formatting -----------------
GREEN="\033[32m"
BLUE="\033[34m"
CYAN="\033[36m"
YELLOW="\033[33m"
RED="\033[31m"
BOLD="\033[1m"
RESET="\033[0m"

log()     { printf "${GREEN}${BOLD}==>${RESET} %s\n" "$*"; }
info()    { printf "${CYAN}-->${RESET} %s\n" "$*"; }
warn()    { printf "${YELLOW}${BOLD}WARNING:${RESET} %s\n" "$*" >&2; }
die()     { printf "${RED}${BOLD}ERROR:${RESET} %s\n" "$*" >&2; exit 1; }

# ----------------- Default Configurations -----------------
REPO_URL="https://github.com/500wango/arcmux.git"
DEFAULT_INSTALL_DIR="/opt/arcmux"
DEFAULT_PORT="3000"
DEFAULT_BIND="0.0.0.0"

INSTALL_DIR=""
HOST_PORT=""
HOST_BIND=""
NON_INTERACTIVE=0
FORCE_REBUILD=0

# ----------------- Helper Functions -----------------
rand_hex() {
  local bytes="${1:-32}"
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex "$bytes"
  else
    head -c "$bytes" /dev/urandom | od -An -tx1 | tr -d ' \n'
  fi
}

check_root() {
  if [[ "$EUID" -ne 0 ]]; then
    if command -v sudo >/dev/null 2>&1; then
      warn "Running with sudo privileges..."
      exec sudo -E bash "$0" "$@"
    else
      die "This script must be run as root or with sudo."
    fi
  fi
}

get_public_ip() {
  local ip=""
  ip=$(curl -s4m 3 https://api.ipify.org 2>/dev/null || curl -s4m 3 https://icanhazip.com 2>/dev/null || hostname -I 2>/dev/null | awk '{print $1}')
  echo "${ip:-127.0.0.1}"
}

# ----------------- Dependency Installation -----------------
install_pkg_dependencies() {
  log "Checking system dependencies..."
  local to_install=()

  for cmd in curl git openssl tar; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
      to_install+=("$cmd")
    fi
  done

  if [[ ${#to_install[@]} -gt 0 ]]; then
    info "Installing required packages: ${to_install[*]}"
    if command -v apt-get >/dev/null 2>&1; then
      apt-get update -qq && apt-get install -y -qq "${to_install[@]}" ca-certificates
    elif command -v dnf >/dev/null 2>&1; then
      dnf install -y -q "${to_install[@]}" ca-certificates
    elif command -v yum >/dev/null 2>&1; then
      yum install -y -q "${to_install[@]}" ca-certificates
    elif command -v apk >/dev/null 2>&1; then
      apk add --no-cache "${to_install[@]}" ca-certificates
    elif command -v pacman >/dev/null 2>&1; then
      pacman -Sy --noconfirm "${to_install[@]}" ca-certificates
    else
      die "Unsupported package manager. Please install: ${to_install[*]}"
    fi
  fi
}

install_docker_if_needed() {
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    info "Docker and Docker Compose are already installed."
    return
  fi

  log "Docker not found or incomplete. Installing Docker and Docker Compose plugin..."
  curl -fsSL https://get.docker.com | sh

  # Start and enable docker service
  if command -v systemctl >/dev/null 2>&1; then
    systemctl daemon-reload || true
    systemctl start docker || true
    systemctl enable docker || true
  elif command -v service >/dev/null 2>&1; then
    service docker start || true
  fi

  if ! command -v docker >/dev/null 2>&1; then
    die "Docker installation failed. Please install Docker manually: https://docs.docker.com/engine/install/"
  fi

  if ! docker compose version >/dev/null 2>&1; then
    log "Installing docker compose plugin..."
    if command -v apt-get >/dev/null 2>&1; then
      apt-get update -qq && apt-get install -y -qq docker-compose-plugin
    elif command -v dnf >/dev/null 2>&1; then
      dnf install -y -q docker-compose-plugin
    elif command -v yum >/dev/null 2>&1; then
      yum install -y -q docker-compose-plugin
    fi
  fi
}

# ----------------- Repository Setup -----------------
setup_repository() {
  # If currently running inside a cloned arcmux repository (contains go.mod and Dockerfile)
  if [[ -f "./go.mod" && -f "./Dockerfile" && -f "./docker-compose.prod.yml" ]]; then
    INSTALL_DIR="$(pwd)"
    log "Using current repository directory: $INSTALL_DIR"
    return
  fi

  INSTALL_DIR="${INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"
  log "Setting up repository at $INSTALL_DIR..."

  if [[ -d "$INSTALL_DIR/.git" ]]; then
    info "Existing repository detected at $INSTALL_DIR, pulling latest changes..."
    cd "$INSTALL_DIR"
    git fetch --all --prune
    git reset --hard origin/main || git pull origin main || true
  else
    mkdir -p "$(dirname "$INSTALL_DIR")"
    info "Cloning ArcMux from $REPO_URL to $INSTALL_DIR..."
    git clone "$REPO_URL" "$INSTALL_DIR"
    cd "$INSTALL_DIR"
  fi
}

# ----------------- Environment & Secrets -----------------
setup_env_file() {
  local env_file="$INSTALL_DIR/.env.prod"
  local env_example="$INSTALL_DIR/deploy/docker.env.example"

  if [[ -f "$env_file" ]]; then
    info "Existing environment file found: $env_file"
    # Update port if explicitly customized
    if [[ -n "$HOST_PORT" ]]; then
      sed -i "s|^HOST_PORT=.*|HOST_PORT=${HOST_PORT}|" "$env_file"
    fi
    if [[ -n "$HOST_BIND" ]]; then
      sed -i "s|^HOST_BIND=.*|HOST_BIND=${HOST_BIND}|" "$env_file"
    fi
    return
  fi

  log "Generating secure production environment file (.env.prod)..."
  [[ -f "$env_example" ]] || die "Cannot find $env_example"

  local pg_pass redis_pass session_secret crypto_secret oauth_secret
  pg_pass="$(rand_hex 16)"
  redis_pass="$(rand_hex 16)"
  session_secret="$(rand_hex 32)"
  crypto_secret="$(rand_hex 32)"
  oauth_secret="$(rand_hex 32)"

  local port="${HOST_PORT:-$DEFAULT_PORT}"
  local bind="${HOST_BIND:-$DEFAULT_BIND}"

  sed \
    -e "s|^HOST_PORT=.*|HOST_PORT=${port}|" \
    -e "s|^HOST_BIND=.*|HOST_BIND=${bind}|" \
    -e "s|^POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=${pg_pass}|" \
    -e "s|^REDIS_PASSWORD=.*|REDIS_PASSWORD=${redis_pass}|" \
    -e "s|^SESSION_SECRET=.*|SESSION_SECRET=${session_secret}|" \
    -e "s|^CRYPTO_SECRET=.*|CRYPTO_SECRET=${crypto_secret}|" \
    -e "s|^UPSTREAM_CREDENTIAL_ENCRYPTION_KEY=.*|UPSTREAM_CREDENTIAL_ENCRYPTION_KEY=${oauth_secret}|" \
    "$env_example" > "$env_file"

  chmod 600 "$env_file"
  info "Created $env_file with randomized secrets."
}

# ----------------- Global CLI Management Script -----------------
install_cli_helper() {
  local cli_path="/usr/local/bin/arcmux"
  log "Installing global management CLI tool ($cli_path)..."

  cat > "$cli_path" <<EOF
#!/usr/bin/env bash
# ArcMux CLI Manager
set -euo pipefail

APP_DIR="$INSTALL_DIR"
COMPOSE_FILE="\$APP_DIR/docker-compose.prod.yml"
ENV_FILE="\$APP_DIR/.env.prod"

cd "\$APP_DIR"

compose() {
  docker compose -f "\$COMPOSE_FILE" --env-file "\$ENV_FILE" "\$@"
}

case "\${1:-help}" in
  start)
    echo "==> Starting ArcMux stack..."
    compose up -d
    ;;
  stop)
    echo "==> Stopping ArcMux stack..."
    compose stop
    ;;
  restart)
    echo "==> Restarting ArcMux stack..."
    compose restart
    ;;
  down)
    echo "==> Shutting down ArcMux stack..."
    compose down
    ;;
  status)
    echo "==> ArcMux Container Status:"
    compose ps
    ;;
  logs)
    shift
    compose logs -f --tail=100 "\$@"
    ;;
  update)
    echo "==> Pulling latest source and rebuilding..."
    git pull origin main
    compose build --no-cache
    compose up -d --force-recreate
    echo "==> ArcMux update completed."
    ;;
  rebuild)
    echo "==> Rebuilding ArcMux container..."
    compose up -d --build
    ;;
  env)
    \${EDITOR:-nano} "\$ENV_FILE"
    echo "==> Restarting containers with updated configuration..."
    compose up -d
    ;;
  info)
    source "\$ENV_FILE"
    IP=\$(curl -s4m 3 https://api.ipify.org 2>/dev/null || hostname -I | awk '{print \$1}')
    echo "--------------------------------------------------"
    echo " ArcMux Service Information"
    echo "--------------------------------------------------"
    echo " Install Dir : \$APP_DIR"
    echo " Access URL  : http://\${IP}:\${HOST_PORT:-3000}/"
    echo " Local URL   : http://127.0.0.1:\${HOST_PORT:-3000}/"
    echo " Status      : \$(compose ps --status running --format '{{.Service}}' | tr '\n' ' ')"
    echo "--------------------------------------------------"
    ;;
  help|--help|-h)
    cat <<HELP
ArcMux Command Line Utility

Usage: arcmux <command>

Commands:
  status     Show running containers and health status
  logs       Follow real-time service logs (e.g. arcmux logs arcmux)
  start      Start the container stack
  stop       Stop all services
  restart    Restart all services
  rebuild    Rebuild and restart the container stack
  update     Pull latest git release and upgrade stack
  env        Edit .env.prod configuration file
  info       Display service URLs and port info
  down       Tear down container stack (keeps database volumes)
HELP
    ;;
  *)
    echo "Unknown command: \$1. Use 'arcmux help' for options."
    exit 1
    ;;
esac
EOF

  chmod +x "$cli_path"
  info "CLI helper installed! You can now manage your service using 'arcmux <command>' from anywhere."
}

# ----------------- Start Services & Healthcheck -----------------
start_services() {
  log "Starting ArcMux stack (PostgreSQL 15 + Redis 7 + ArcMux Gateway)..."
  cd "$INSTALL_DIR"

  mkdir -p data logs

  if [[ "$FORCE_REBUILD" -eq 1 ]]; then
    docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build --force-recreate
  else
    docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --build
  fi

  log "Waiting for services to become healthy..."
  local port
  port="$(grep -E '^HOST_PORT=' .env.prod | cut -d'=' -f2 || echo 3000)"
  port="${port:-3000}"

  local max_retries=30
  local count=0
  local healthy=0

  while [[ $count -lt $max_retries ]]; do
    if curl -fsS "http://127.0.0.1:${port}/api/status" 2>/dev/null | grep -q '"success"[[:space:]]*:[[:space:]]*true'; then
      healthy=1
      break
    fi
    sleep 2
    count=$((count + 1))
    printf "."
  done
  echo ""

  if [[ "$healthy" -eq 1 ]]; then
    log "${GREEN}ArcMux is UP and HEALTHY!${RESET}"
  else
    warn "Service started, but health check is still initializing. Check logs with 'arcmux logs'."
  fi
}

# ----------------- Print Summary Banner -----------------
print_banner() {
  local public_ip port
  public_ip="$(get_public_ip)"
  port="$(grep -E '^HOST_PORT=' "$INSTALL_DIR/.env.prod" | cut -d'=' -f2 || echo 3000)"
  port="${port:-3000}"

  cat <<EOF

${GREEN}${BOLD}===================================================================${RESET}
${GREEN}${BOLD}              🎉 ArcMux Deployment Succeeded!                      ${RESET}
${GREEN}${BOLD}===================================================================${RESET}

${BOLD}Access your ArcMux Web Console:${RESET}
  ➜ ${CYAN}${BOLD}http://${public_ip}:${port}/${RESET}
  ➜ ${CYAN}http://127.0.0.1:${port}/${RESET} (Local)

${BOLD}First Time Setup:${RESET}
  1. Open the URL above in your browser.
  2. Follow the setup wizard to initialize the database and create the root admin account.

${BOLD}Management Commands:${RESET}
  • View status  : ${YELLOW}arcmux status${RESET}
  • View logs    : ${YELLOW}arcmux logs${RESET} (or ${YELLOW}arcmux logs arcmux${RESET})
  • Restart app  : ${YELLOW}arcmux restart${RESET}
  • Upgrade app  : ${YELLOW}arcmux update${RESET}
  • Edit config  : ${YELLOW}arcmux env${RESET}
  • View info    : ${YELLOW}arcmux info${RESET}

${BOLD}Installation Details:${RESET}
  • Directory    : ${INSTALL_DIR}
  • Config File  : ${INSTALL_DIR}/.env.prod
  • Data Storage : ${INSTALL_DIR}/data
  • Logs Directory: ${INSTALL_DIR}/logs

${GREEN}===================================================================${RESET}
EOF
}

# ----------------- Interactive Prompts -----------------
interactive_setup() {
  if [[ "$NON_INTERACTIVE" -eq 1 || ! -t 0 ]]; then
    return
  fi

  printf "\n${BOLD}${CYAN}=== ArcMux Installation Configuration ===${RESET}\n"

  # Installation directory prompt
  read -r -p "Install directory [default: ${DEFAULT_INSTALL_DIR}]: " input_dir
  if [[ -n "$input_dir" ]]; then
    INSTALL_DIR="$input_dir"
  fi

  # Host port prompt
  read -r -p "Web port [default: ${DEFAULT_PORT}]: " input_port
  if [[ -n "$input_port" ]]; then
    HOST_PORT="$input_port"
  fi

  echo ""
}

# ----------------- CLI Argument Parsing -----------------
parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -p|--port)
        HOST_PORT="$2"
        shift 2
        ;;
      -b|--bind)
        HOST_BIND="$2"
        shift 2
        ;;
      -d|--dir)
        INSTALL_DIR="$2"
        shift 2
        ;;
      -y|--yes|--non-interactive)
        NON_INTERACTIVE=1
        shift
        ;;
      --rebuild)
        FORCE_REBUILD=1
        shift
        ;;
      -h|--help)
        cat <<USAGE
ArcMux VPS One-Click Installer

Usage: $0 [OPTIONS]

Options:
  -p, --port <port>       Web console port (default: 3000)
  -b, --bind <ip>         Host bind address (default: 0.0.0.0)
  -d, --dir <path>        Installation directory (default: /opt/arcmux)
  -y, --yes               Non-interactive mode (auto-accept defaults)
      --rebuild           Force image rebuild during setup
  -h, --help              Show this help message

Examples:
  bash install.sh
  bash install.sh -p 8080 -y
USAGE
        exit 0
        ;;
      *)
        die "Unknown option: $1 (Use --help for usage)"
        ;;
    esac
  done
}

# ----------------- Main Execution -----------------
main() {
  parse_args "$@"
  check_root "$@"
  interactive_setup
  install_pkg_dependencies
  install_docker_if_needed
  setup_repository
  setup_env_file
  install_cli_helper
  start_services
  print_banner
}

main "$@"
