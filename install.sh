#!/usr/bin/env bash
set -e

# ==========================================
# MEECORE installer
# ==========================================

REPO="https://github.com/nxwex/meecore.git"
APP_DIR="/opt/meecore"
APP_NAME="meecore"
GO_VERSION="1.26.1"
DB_NAME="meecore"
DB_USER="meecore"
DB_PASSWORD="meecore"

# ==========================================
# Helpers
# ==========================================

info() {
    echo
    echo "==> $1"
}

error() {
    echo
    echo "ERROR: $1"
    exit 1
}

# ==========================================
# Root check
# ==========================================

if [ "$EUID" -ne 0 ]; then
    error "Run this script as root:
sudo ./install.sh"
fi

# ==========================================
# OS check
# ==========================================

if [ ! -f /etc/os-release ]; then
    error "Cannot detect operating system."
fi

. /etc/os-release

case "$ID" in
    ubuntu|debian)
        ;;
    *)
        error "Unsupported OS: $ID. Only Ubuntu and Debian are supported."
        ;;
esac

ARCH="$(dpkg --print-architecture)"

case "$ARCH" in
    amd64)
        GO_ARCH="amd64"
        ;;
    arm64)
        GO_ARCH="arm64"
        ;;
    *)
        error "Unsupported architecture: $ARCH"
        ;;
esac

echo
echo "=========================================="
echo "          MEECORE INSTALLER"
echo "=========================================="
echo
echo "OS:           $PRETTY_NAME"
echo "Architecture: $ARCH"
echo "Repository:   $REPO"
echo "Install path: $APP_DIR"
echo

# ==========================================
# Update system
# ==========================================

info "Updating package lists"
apt-get update

info "Upgrading system"
apt-get upgrade -y

# ==========================================
# Basic packages
# ==========================================

info "Installing basic packages"
apt-get install -y \
    ca-certificates \
    curl \
    wget \
    gnupg \
    lsb-release \
    git \
    unzip \
    jq \
    nano \
    vim \
    htop \
    net-tools \
    build-essential \
    pkg-config \
    openssl

# ==========================================
# Docker
# ==========================================

info "Installing Docker"

if command -v docker >/dev/null 2>&1; then
    echo "Docker is already installed."
else
    if [ "$ID" = "ubuntu" ]; then
        DOCKER_DISTRO="ubuntu"
    else
        DOCKER_DISTRO="debian"
    fi

    install -m 0755 -d /etc/apt/keyrings
    rm -f /etc/apt/keyrings/docker.asc

    curl -fsSL \
        "https://download.docker.com/linux/${DOCKER_DISTRO}/gpg" \
        -o /etc/apt/keyrings/docker.asc

    chmod a+r /etc/apt/keyrings/docker.asc
    rm -f /etc/apt/sources.list.d/docker.list

    echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/${DOCKER_DISTRO} ${VERSION_CODENAME} stable" \
        > /etc/apt/sources.list.d/docker.list

    apt-get update

    apt-get install -y \
        docker-ce \
        docker-ce-cli \
        containerd.io \
        docker-buildx-plugin \
        docker-compose-plugin
fi

systemctl enable docker
systemctl start docker

echo
docker --version
docker compose version

# ==========================================
# PostgreSQL
# ==========================================

info "Installing PostgreSQL"

if command -v psql >/dev/null 2>&1; then
    echo "PostgreSQL is already installed."
else
    apt-get install -y \
        postgresql \
        postgresql-contrib
fi

systemctl enable postgresql
systemctl start postgresql

echo
psql --version

# ==========================================
# PostgreSQL database
# ==========================================

info "Configuring PostgreSQL"

# Create or update user with password
if sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='${DB_USER}'" | grep -q 1; then
    echo "PostgreSQL user '${DB_USER}' already exists. Updating password..."
    sudo -u postgres psql -c "ALTER USER ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';"
else
    echo "Creating PostgreSQL user '${DB_USER}'..."
    sudo -u postgres psql -c "CREATE USER ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';"
fi

# Create database if missing
if sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}'" | grep -q 1; then
    echo "PostgreSQL database '${DB_NAME}' already exists."
else
    echo "Creating PostgreSQL database '${DB_NAME}'..."
    sudo -u postgres createdb -O "${DB_USER}" "${DB_NAME}"
fi

# ==========================================
# Go
# ==========================================

info "Installing Go"

if command -v go >/dev/null 2>&1; then
    CURRENT_GO=$(go version | awk '{print $3}' | sed 's/go//')
    echo "Go is already installed: $CURRENT_GO"

    if [ "$CURRENT_GO" != "$GO_VERSION" ]; then
        echo "Updating Go to ${GO_VERSION}..."
        cd /tmp
        GO_ARCHIVE="go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
        wget -q "https://go.dev/dl/${GO_ARCHIVE}" -O "${GO_ARCHIVE}"
        rm -rf /usr/local/go
        tar -C /usr/local -xzf "${GO_ARCHIVE}"
        rm -f "${GO_ARCHIVE}"
    fi
else
    cd /tmp
    GO_ARCHIVE="go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
    wget -q "https://go.dev/dl/${GO_ARCHIVE}" -O "${GO_ARCHIVE}"
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "${GO_ARCHIVE}"
    rm -f "${GO_ARCHIVE}"
fi

cat > /etc/profile.d/go.sh <<'EOF'
export PATH="/usr/local/go/bin:$PATH"
EOF

export PATH="/usr/local/go/bin:$PATH"

echo
go version

# ==========================================
# Clone repository
# ==========================================

info "Preparing MEECORE"

if [ -d "${APP_DIR}/.git" ]; then
    echo "Repository already exists."
    cd "${APP_DIR}"
    git fetch --all
    git pull --ff-only || true
else
    mkdir -p "$(dirname "${APP_DIR}")"
    git clone "${REPO}" "${APP_DIR}"
    cd "${APP_DIR}"
fi

# ==========================================
# Go dependencies
# ==========================================

info "Downloading Go dependencies"
go mod download

# ==========================================
# Build
# ==========================================

info "Building MEECORE"
go build \
    -o "${APP_NAME}" \
    ./cmd/meecore/main.go

# ==========================================
# Environment
# ==========================================

info "Creating environment file"

cat > "${APP_DIR}/.env" <<EOF
DATABASE_DSN=postgres://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}?sslmode=disable
HTTP_ADDR=:8080
EOF

chmod 600 "${APP_DIR}/.env"

# ==========================================
# Permissions
# ==========================================

info "Setting permissions"
chown -R root:root "${APP_DIR}"
chmod +x "${APP_DIR}/${APP_NAME}"

# ==========================================
# Systemd service
# ==========================================

info "Creating systemd service"

cat > /etc/systemd/system/meecore.service <<EOF
[Unit]
Description=MEECORE Game Server Management Panel
After=docker.service postgresql.service
Requires=docker.service postgresql.service

[Service]
Type=simple
WorkingDirectory=${APP_DIR}
ExecStart=${APP_DIR}/${APP_NAME}
Restart=always
RestartSec=5
EnvironmentFile=${APP_DIR}/.env

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable meecore

# ==========================================
# Start MEECORE
# ==========================================

info "Starting MEECORE"
systemctl restart meecore
sleep 2

# ==========================================
# Status
# ==========================================

info "Installation status"

echo
echo "Docker:"
systemctl is-active docker || true

echo
echo "PostgreSQL:"
systemctl is-active postgresql || true

echo
echo "MEECORE:"
systemctl is-active meecore || true

echo
echo "=========================================="
echo "       MEECORE INSTALLATION COMPLETE"
echo "=========================================="
echo
echo "Application:"
echo "  ${APP_DIR}/${APP_NAME}"
echo
echo "Repository:"
echo "  ${REPO}"
echo
echo "Environment (.env):"
echo "  DATABASE_DSN=postgres://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}?sslmode=disable"
echo "  HTTP_ADDR=:8080"
echo
echo "Service commands:"
echo "  systemctl status meecore"
echo "  systemctl restart meecore"
echo "  journalctl -u meecore -f"
echo
echo "=========================================="