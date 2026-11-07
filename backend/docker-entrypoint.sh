#!/bin/sh
# Démarre en root pour attribuer le dossier de données à PUID:PGID,
# puis lance le serveur sous cet utilisateur (jamais en root).
set -eu
PUID="${PUID:-1000}"
PGID="${PGID:-1000}"
DATA_DIR="$(dirname "${DATA_FILE:-/data/counters.json}")"

if [ "$(id -u)" = "0" ]; then
  mkdir -p "$DATA_DIR"
  chown -R "$PUID:$PGID" "$DATA_DIR"
  exec su-exec "$PUID:$PGID" /usr/local/bin/creme "$@"
fi
exec /usr/local/bin/creme "$@"
