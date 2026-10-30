#!/bin/bash
# Home SOCKS for vitek worker: Mac:11081 ← SSH reverse → vdserv:11080
# Worker uses socks5://172.19.0.1:11080 (docker bridge → host).
# Uses gost (not microsocks): Avito has multiple A records; microsocks sticks to a dead one.
set -uo pipefail

MS_PORT="${VITEK_HOME_MS_PORT:-11081}"
TUN_PORT="${VITEK_HOME_TUN_PORT:-11080}"
SSH_HOST="${VITEK_HOME_SSH_HOST:-vdserv}"
LOG="${VITEK_HOME_PROXY_LOG:-$HOME/Library/Logs/vitek-home-proxy.log}"
GOST_BIN="${GOST_BIN:-$(command -v gost || true)}"
AUTOSSH_BIN="${AUTOSSH_BIN:-$(command -v autossh || true)}"

mkdir -p "$(dirname "$LOG")"
exec >>"$LOG" 2>&1

log() { printf '%s %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*"; }

if [[ -z "$GOST_BIN" || ! -x "$GOST_BIN" ]]; then
  log "ERROR: gost not found (brew install gost)"
  exit 1
fi

if command -v caffeinate >/dev/null 2>&1; then
  caffeinate -i -w $$ &
fi

cleanup() {
  if [[ -n "${SOCKPID:-}" ]]; then
    kill "$SOCKPID" 2>/dev/null || true
    wait "$SOCKPID" 2>/dev/null || true
  fi
}
trap cleanup EXIT INT TERM

free_port() {
  if command -v lsof >/dev/null 2>&1; then
    local pids
    pids="$(lsof -nP -iTCP:"${MS_PORT}" -sTCP:LISTEN -t 2>/dev/null || true)"
    if [[ -n "$pids" ]]; then
      # shellcheck disable=SC2086
      kill $pids 2>/dev/null || true
      sleep 0.3
    fi
  fi
}

start_socks() {
  if [[ -n "${SOCKPID:-}" ]] && kill -0 "$SOCKPID" 2>/dev/null; then
    return 0
  fi
  free_port
  "$GOST_BIN" -L "socks5://127.0.0.1:${MS_PORT}" &
  SOCKPID=$!
  sleep 0.5
  if ! kill -0 "$SOCKPID" 2>/dev/null; then
    log "ERROR: gost failed to start on :${MS_PORT}"
    return 1
  fi
  if ! curl -sf --max-time 8 --socks5-hostname "127.0.0.1:${MS_PORT}" https://example.com >/dev/null; then
    log "WARN: gost up but local SOCKS probe failed"
  fi
  log "gost ready pid=$SOCKPID port=$MS_PORT"
}

run_tunnel() {
  local ssh_opts=(
    -N
    -o ServerAliveInterval=15
    -o ServerAliveCountMax=3
    -o ExitOnForwardFailure=yes
    -o TCPKeepAlive=yes
    -R "0.0.0.0:${TUN_PORT}:127.0.0.1:${MS_PORT}"
    "$SSH_HOST"
  )
  if [[ -n "$AUTOSSH_BIN" && -x "$AUTOSSH_BIN" ]]; then
    AUTOSSH_GATETIME=0 AUTOSSH_PORT=0 "$AUTOSSH_BIN" -M 0 "${ssh_opts[@]}"
  else
    ssh "${ssh_opts[@]}"
  fi
}

log "start host=$SSH_HOST socks=$MS_PORT tun=$TUN_PORT gost=$GOST_BIN autossh=${AUTOSSH_BIN:-none}"

backoff=2
while true; do
  start_socks || {
    sleep "$backoff"
    backoff=$((backoff < 60 ? backoff * 2 : 60))
    continue
  }
  log "opening reverse tunnel :${TUN_PORT} -> 127.0.0.1:${MS_PORT}"
  if run_tunnel; then
    log "tunnel exited 0"
  else
    log "tunnel exited $? — reconnect"
  fi
  backoff=2
  sleep 2
done
