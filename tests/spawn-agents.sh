#!/usr/bin/env bash
set -euo pipefail

CORE_URL="${CORE_URL:-http://localhost:8080}"
AGENT_COUNT="${NPC_COUNT:-500}"
PID_DIR="/tmp/divinity"
AGENT_DIR="$PID_DIR/agents"
NPC_BIN="$PID_DIR/npc-bin"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "[spawn-agents] Waiting for core server at $CORE_URL ..."

while true; do
    status=$(curl -s "$CORE_URL/health" 2>/dev/null || echo '{}')
    if echo "$status" | grep -q '"status":"ok"'; then
        break
    fi
    sleep 2
done

echo "[spawn-agents] Core server is ready"
echo "[spawn-agents] Building NPC binary ..."
cd "$SCRIPT_DIR/../npcs" && go build -o "$NPC_BIN" .
echo "[spawn-agents] NPC binary built at $NPC_BIN"

mkdir -p "$AGENT_DIR"

echo "[spawn-agents] Launching $AGENT_COUNT agents ..."

for i in $(seq 0 $((AGENT_COUNT - 1))); do
    dir="$AGENT_DIR/agent_$i"
    mkdir -p "$dir"
    log="$AGENT_DIR/agent_$i.log"
    pid_file="$AGENT_DIR/agent_$i.pid"

    cd "$dir"
    "$NPC_BIN" > "$log" 2>&1 &
    echo $! > "$pid_file"
done

echo "[spawn-agents] All $AGENT_COUNT agents launched"
