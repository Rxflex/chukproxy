#!/bin/bash
set -e

# ---------- OS & ARCH DETECTION ----------
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

EXT=""
case "$OS" in
    linux|darwin) ;;
    msys*|mingw*|cygwin*|windows)
        OS="windows"
        EXT=".exe"
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

# ---------- SELECT BINARY ----------
FILENAME="goproxy_${OS}_${ARCH}${EXT}"

# ---------- DOWNLOAD ----------
BASE_URL="https://github.com/Rxflex/chukproxy/releases/download/v1.0.0/"
URL="${BASE_URL}${FILENAME}"

echo "Detected: OS=$OS, ARCH=$ARCH"
echo "Binary to download: $FILENAME"
echo "Downloading from $URL..."

curl -LO "$URL"

# ---------- INSTALL DIRECTORY ----------
read -p "Enter installation path [/opt/chukproxy]: " INSTALL_DIR
INSTALL_DIR=${INSTALL_DIR:-/opt/chukproxy}

read -p "Enter desired binary name [goproxy]: " BINARY_NAME
BINARY_NAME=${BINARY_NAME:-goproxy}

mkdir -p "$INSTALL_DIR"
chmod +x "$FILENAME"
mv "$FILENAME" "$INSTALL_DIR/${BINARY_NAME}${EXT}"

echo "Binary installed to $INSTALL_DIR/${BINARY_NAME}${EXT}"

# ---------- CONFIGURATION ----------
read -p "Create config.yaml? [y/N]: " CONFIRM_CONFIG
if [[ "$CONFIRM_CONFIG" =~ ^[Yy]$ ]]; then
    read -p "Database user: " DB_USER
    read -p "Database password: " DB_PASS
    read -p "Database host [127.0.0.1]: " DB_HOST
    read -p "Database port [3306]: " DB_PORT
    read -p "Database name: " DB_NAME

    DB_HOST=${DB_HOST:-127.0.0.1}
    DB_PORT=${DB_PORT:-3306}

    cat > "$INSTALL_DIR/config.yaml" <<EOF
database:
  user: "$DB_USER"
  password: "$DB_PASS"
  host: "$DB_HOST"
  port: $DB_PORT
  dbname: "$DB_NAME"
EOF

    echo "config.yaml created in $INSTALL_DIR"
fi

# ---------- SYSTEMD ----------
if [[ "$OS" == "linux" ]]; then
    read -p "Install as systemd service? [y/N]: " CONFIRM_SYSTEMD
    if [[ "$CONFIRM_SYSTEMD" =~ ^[Yy]$ ]]; then
        read -p "Systemd service name [chukproxy]: " SERVICE_NAME
        SERVICE_NAME=${SERVICE_NAME:-chukproxy}
        SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}.service"

        sudo tee "$SERVICE_PATH" > /dev/null <<EOF
[Unit]
Description=ChukProxy Service
After=network.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/${BINARY_NAME}${EXT}
WorkingDirectory=$INSTALL_DIR
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

        sudo systemctl daemon-reexec
        sudo systemctl daemon-reload
        sudo systemctl enable "$SERVICE_NAME"
        sudo systemctl start "$SERVICE_NAME"

        echo "Systemd service '$SERVICE_NAME' has been installed and started."
    fi
fi

echo "✅ Installation complete."
