#!/bin/bash
set -euo pipefail

: ${DATA_PATH:=/tmp/vsdata}
: ${OPTIONS:=}

mkdir -p "$DATA_PATH"
chown -R 1000:1000 "$DATA_PATH" 2>/dev/null || true

cd /opt/vintagestory

if [ ! -f "VintagestoryServer.dll" ] && [ ! -x "VintagestoryServer" ]; then
  echo "Fatal: Vintagestory server files not found in /opt/vintagestory"
  echo "Make sure you placed the archive in the build context and built the image."
  exit 1
fi

echo "Starting Vintage Story server with data path: ${DATA_PATH}"


if [ -f "${DATA_PATH}/serverconfig.json" ]; then
  if command -v jq >/dev/null 2>&1; then
    tmpcfg=$(mktemp)
    jq '.AdvertiseServer = true | .WhitelistMode = 1' "${DATA_PATH}/serverconfig.json" > "$tmpcfg" && mv "$tmpcfg" "${DATA_PATH}/serverconfig.json" || rm -f "$tmpcfg"
  fi
else
  if [ -f "/opt/vintagestory/serverconfig.json" ]; then
    cp "/opt/vintagestory/serverconfig.json" "${DATA_PATH}/serverconfig.json"
    if command -v jq >/dev/null 2>&1; then
      tmpcfg=$(mktemp)
      jq '.AdvertiseServer = true | .WhitelistMode = 1' "${DATA_PATH}/serverconfig.json" > "$tmpcfg" && mv "$tmpcfg" "${DATA_PATH}/serverconfig.json" || rm -f "$tmpcfg"
    fi
  else
    echo "Warning: serverconfig.json template missing; letting server create defaults."
  fi
fi

if [ -x "./VintagestoryServer" ]; then
  exec ./VintagestoryServer --dataPath "${DATA_PATH}" ${OPTIONS}
else
  exec dotnet VintagestoryServer.dll --dataPath "${DATA_PATH}" ${OPTIONS}
fi