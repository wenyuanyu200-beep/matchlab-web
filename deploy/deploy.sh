#!/usr/bin/env sh
set -eu

if [ "$(id -u)" -ne 0 ]; then
    echo "Please run with sudo: sudo sh deploy/deploy.sh" >&2
    exit 1
fi

PROJECT_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
INSTALL_ROOT=${INSTALL_ROOT:-/opt/matchlab}
SERVICE_USER=${SERVICE_USER:-matchlab}

if ! command -v go >/dev/null 2>&1; then
    echo "Go is required but was not found in PATH." >&2
    exit 1
fi

echo "Building MatchLab API..."
mkdir -p "$PROJECT_ROOT/backend/bin"
cd "$PROJECT_ROOT/backend"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/matchlab-api ./cmd/server

if ! id "$SERVICE_USER" >/dev/null 2>&1; then
    useradd --system --home-dir "$INSTALL_ROOT" --shell /usr/sbin/nologin "$SERVICE_USER"
fi

install -d -o "$SERVICE_USER" -g "$SERVICE_USER" "$INSTALL_ROOT/backend/bin"
install -m 0755 "$PROJECT_ROOT/backend/bin/matchlab-api" "$INSTALL_ROOT/backend/bin/matchlab-api"

if [ ! -f "$INSTALL_ROOT/backend/.env" ]; then
    install -m 0600 -o "$SERVICE_USER" -g "$SERVICE_USER" \
        "$PROJECT_ROOT/backend/.env.example" "$INSTALL_ROOT/backend/.env"
    echo "Created $INSTALL_ROOT/backend/.env; edit its database password before production use."
fi

install -m 0644 "$PROJECT_ROOT/deploy/matchlab-api.service" /etc/systemd/system/matchlab-api.service
install -m 0644 "$PROJECT_ROOT/deploy/nginx-matchlab.conf" /etc/nginx/sites-available/matchlab
ln -sfn /etc/nginx/sites-available/matchlab /etc/nginx/sites-enabled/matchlab

systemctl daemon-reload
nginx -t
systemctl enable --now matchlab-api
systemctl reload nginx

echo "Deployment complete. Check: curl http://127.0.0.1:8080/api/health"
