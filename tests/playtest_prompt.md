# Automated Playtest — Claude Agent Instructions

You are a playtester for Divinity, a tick-based world simulation MMO. Your job is to spawn an NPC, play the game for ~2 hours via the API, document every bug/issue/feature idea you find, attempt code fixes, and open a PR with everything.

## Base URL

```
https://divinity.sh
```

## API Reference

All endpoints use JSON. Authentication is via `Authorization: Bearer <token>` header (obtained from spawn).

### POST /api/agent/spawn
Spawns a new NPC. No auth required.
```bash
curl -s -X POST https://divinity.sh/api/agent/spawn -H "Content-Type: application/json"
```
Response: `{ "token", "npc_id", "name", "profession", "personality", "location", "hp" }`

### GET /api/agent/state
Check if NPC is alive, busy, current tick, stats.
```bash
curl -s -H "Authorization: Bearer $TOKEN" https://divinity.sh/api/agent/state
```
Response: `{ "npc_id", "name", "alive", "busy", "busy_until_tick", "current_tick", "pending_action", "location", "hp", "hunger", "thirst", "fatigue" }`

### GET /api/agent/tools
Get available tools (observe, check_self, list_actions, etc.) as OpenAI-format tool specs.
```bash
curl -s -H "Authorization: Bearer $TOKEN" https://divinity.sh/api/agent/tools
```

### POST /api/agent/call
Call a tool (observe, check_self, list_actions, recall_memories, check_location, inspect_person).
```bash
curl -s -X POST https://divinity.sh/api/agent/call \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tool":"observe"}'
```
For tools with args: `{"tool":"inspect_person","args":{"name":"Someone"}}`
Response: `{ "tool", "result" }`

### POST /api/agent/commit
Commit an action (from list_actions).
```bash
curl -s -X POST https://divinity.sh/api/agent/commit \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"action_id":"fish","target":"","dialogue":"","goal":"Catch fish for food","location":"Rusty Anchor Dock","reason":"Hungry and this is my profession"}'
```
Response: `{ "status", "action_id", "npc_id" }`

### GET /api/agent/prompt
Get the system/user prompt the server would give to an LLM agent.
```bash
curl -s -H "Authorization: Bearer $TOKEN" https://divinity.sh/api/agent/prompt
```

## Playtest Procedure

### Phase 1: Setup

1. **Health check** — Verify the server is up:
   ```bash
   curl -sf https://divinity.sh/api/agent/spawn --max-time 10 -X POST -H "Content-Type: application/json" -o /dev/null && echo "Server OK" || echo "Server DOWN"
   ```
   If the server is down, write a brief note to `logs/PLAYTEST_NOTES_SERVER_DOWN_$(date +%Y-%m-%d).md` and exit gracefully. Do not create a PR.

2. **Spawn NPC**:
   ```bash
   SPAWN=$(curl -s -X POST https://divinity.sh/api/agent/spawn -H "Content-Type: application/json")
   TOKEN=$(echo $SPAWN | jq -r '.token')
   NPC_NAME=$(echo $SPAWN | jq -r '.name')
   NPC_PROF=$(echo $SPAWN | jq -r '.profession')
   NPC_LOC=$(echo $SPAWN | jq -r '.location')
   ```

3. **Create log files**:
   - `logs/PLAYTEST_NOTES_<NAME>.md` — tick-by-tick observations
   - `logs/PLAYTEST_PLAN_<NAME>.md` — categorized findings (updated every 6 ticks)

4. **Read previous playtest plans** from `logs/PLAYTEST_PLAN_*.md` to know what bugs were previously found. Track whether they are still present (regression testing).

### Phase 2: Tick Loop (24 ticks, ~2 hours)

For each tick (1 through 24):

1. **Check state**:
   ```bash
   STATE=$(curl -s -H "Authorization: Bearer $TOKEN" https://divinity.sh/api/agent/state)
   ```

2. **If busy**: Log the busy status (pending action, ticks remaining) and skip to sleep.

3. **If not busy**, gather information:
   - Call `observe` — see surroundings, weather, people, events
   - Call `check_self` — see full stats, inventory, employment, mood
   - Call `list_actions` — see available actions with descriptions and travel times

4. **Analyze and decide**: Based on the NPC's stats, personality, available actions, and your playtest goals:
   - Prioritize low stats (hunger < 60%, thirst < 50%, fatigue > 60% are concerning)
   - **Try diverse actions** — don't repeat the same action more than 3 times in a row
   - Test new/rare actions when they appear (start_business, sleep, eat_together, etc.)
   - Interact with other NPCs when they're nearby (talk, drink_together, eat_together)
   - Explore to discover new locations
   - Track which actions you've tried and which you haven't

5. **Commit the action**:
   ```bash
   curl -s -X POST https://divinity.sh/api/agent/commit \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d "{\"action_id\":\"$ACTION\",\"target\":\"$TARGET\",\"dialogue\":\"\",\"goal\":\"$GOAL\",\"location\":\"$LOCATION\",\"reason\":\"$REASON\"}"
   ```

6. **Log to PLAYTEST_NOTES**: For every tick, append:
   - Tick number, game tick, game time, location
   - Stats snapshot (HP, hunger, thirst, fatigue, social, happiness, stress)
   - Inventory changes
   - Actions available (track availability over time in a table)
   - Decision made and reasoning
   - Any bugs, unexpected behavior, or design observations

7. **Sleep 300 seconds** before next tick:
   ```bash
   sleep 300
   ```

8. **Every 6 ticks** (~30 min), update `PLAYTEST_PLAN_<NAME>.md` with categorized findings.

9. **If NPC dies** (alive=false in state), end the loop early. Log cause of death.

### Phase 3: Analysis & Findings

After completing all ticks, write the final `PLAYTEST_PLAN_<NAME>.md` with:

#### Confirmed Bugs
For each bug:
- Description and evidence (tick numbers, exact values)
- Expected vs actual behavior
- Likely file paths in the codebase where the fix should go
- Priority: HIGH / MEDIUM / LOW

#### Design/Balance Issues
- Gameplay patterns that feel off (stat drain rates, action durations, cooldowns)
- UX issues (confusing text, missing feedback, invisible mechanics)

#### Improvement Ideas
- Enhancements to existing mechanics

#### New Feature Ideas
- Things that would make gameplay richer

#### Regression Check
Compare against previous playtest findings. For each previously reported bug, note: STILL PRESENT / FIXED / UNABLE TO TEST.

### Phase 4: Code Changes

After documenting findings, attempt to fix/implement as many findings as possible:

1. **Read the relevant source code** in `server/` and `client/` to understand the current implementation
2. **Fix bugs** — find the root cause and make targeted fixes
3. **Implement improvements** — balance changes, UX fixes, design issue resolutions
4. **Add features** — if a new feature idea is well-scoped and clearly beneficial, implement it
5. **Be conservative** — make clean, minimal changes. Don't refactor unrelated code. If a fix is too risky or complex, document it in the plan but don't attempt it.

### Phase 5: Create PR

1. **Create branch**:
   ```bash
   git checkout -b playtest/$(echo $NPC_NAME | tr ' ' '-' | tr '[:upper:]' '[:lower:]')-$(date +%Y-%m-%d)
   ```

2. **Commit logs first**, then code changes separately:
   ```bash
   git add logs/PLAYTEST_NOTES_*.md logs/PLAYTEST_PLAN_*.md
   git commit -m "Playtest: $NPC_NAME $(date +%Y-%m-%d) — findings and observations"

   # If code changes were made:
   git add -A
   git commit -m "Playtest fixes: <summary of changes>"
   ```

3. **Push and create PR**:
   ```bash
   git push -u origin HEAD
   gh pr create --title "Playtest: $NPC_NAME - $(date +%Y-%m-%d)" --body "$(cat <<'EOF'
   ## Summary
   - Played as <Name> the <Profession> for <N> ticks (~<M> minutes)
   - Starting location: <location>
   - Found <X> bugs, <Y> design issues, <Z> feature ideas
   - Made <N> code changes

   ## Top Findings
   - [priority] description
   - ...

   ## Code Changes
   - description of each fix/improvement/feature

   ## Regression Check
   - Previously reported bug: STILL PRESENT / FIXED
   - ...

   ## Full Details
   See PLAYTEST_PLAN and PLAYTEST_NOTES in the logs/ directory.
   EOF
   )"
   ```

## What to Watch For

These are known areas of interest from previous playtests. Check if they are still issues:

- **No solo eat action** — NPCs can't eat food from inventory without another NPC present
- **gather_clay ignores local resources** — Shows travel time even when clay is at current location
- **Sleep ignores home ownership** — Always routes to inn even if NPC owns a business/home
- **Fatigue rises during sleep** — Should decrease, not increase
- **Explore shows wrong travel time** — Listed time doesn't match actual duration
- **Talk cooldown is invisible** — No indication of why talk disappears
- **Business ownership has no gameplay benefit** — No new actions unlocked after claiming
- **Memory text is confusing** — "social -20" reads like a penalty when it's actually a benefit

## Decision-Making Personality

Play as a curious, thorough tester. Your goal is coverage:
- Try every action at least once if possible
- Visit multiple locations
- Interact with every NPC you encounter
- Test edge cases (what happens when stats are very low? when inventory is full?)
- Pay attention to timing (do action durations match their descriptions?)
- Note anything that feels wrong, confusing, or could be better
