# Automated Playtest — Claude Agent Instructions

You are a playtester **and game developer** for Divinity, a tick-based world simulation MMO. Your job is to spawn an NPC, play the game for ~4 hours via the API, document every bug/issue/feature idea you find, fix bugs, AND implement brand new features/game mechanics, then open a PR with everything.

## Mission Priorities (in order)

1. **Fix bugs** — Any broken or incorrect behavior you encounter. This is the top priority; a buggy game is not fun.
2. **Implement new features and game mechanics** — The game should grow each playtest. Add new actions, new interactions, new systems, new content that makes the world richer. Don't just test what exists; build what's missing.
3. **Polish** — Balance tweaks, UX improvements, clearer text, better feedback.

Every nightly playtest should ship meaningful new content, not just bug fixes. Previous playtests focused almost entirely on bug fixes — now it's time to expand the game.

## Empty World Handling

**The simulation must work even if you are the last NPC on earth.** It is possible to spawn into a world with zero other NPCs. Every action, location, and mechanic should still be playable and interesting for a single NPC:

- If you encounter actions/mechanics that REQUIRE another NPC to function (e.g. eat_together, talk, trade), that is a design gap — add a solo alternative or solo-mode behavior.
- Your playtest must run its full duration even with zero other NPCs present. Test solo gameplay thoroughly if this happens.
- If the world has 0 NPCs beyond yourself, this is a HIGH priority signal that solo-viability needs work. Document every blocker and fix as many as possible.
- A "good world" has content for a lone player: solo crafting, solo exploration, solo survival, solo progression, environmental interactions, inventory/item use, construction, journaling, etc.

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

### Phase 0: Repository Setup

1. **Configure git to use GH_TOKEN for authentication**:
   ```bash
   git config --global url."https://x-access-token:${GH_TOKEN}@github.com/".insteadOf "https://github.com/"
   git config --global user.email "playtest@divinity.sh"
   git config --global user.name "Divinity Playtest"
   ```

2. **Ensure you are on the main branch and up to date**:
   ```bash
   git checkout main
   git pull origin main
   ```
   This must be done before anything else to ensure you are working from the latest codebase.

3. **Check previous playtest PRs** — Review the status of recent playtest PRs and any comments/reviews:
   ```bash
   # List recent playtest PRs with their status
   gh pr list --repo mdnahian/divinity --state all --json number,title,state,mergedAt --limit 10

   # For each recent PR, check for comments and reviews
   gh api repos/mdnahian/divinity/issues/<NUMBER>/comments
   gh api repos/mdnahian/divinity/pulls/<NUMBER>/comments
   gh api repos/mdnahian/divinity/pulls/<NUMBER>/reviews
   ```
   - Note which PRs have been **MERGED** (their fixes are now in main) vs **OPEN** or **CLOSED** (fixes not in main).
   - Read any reviewer comments — they may contain feedback on fix approaches, requests for changes, or notes about bugs that should influence your playtest.
   - Save this context to reference during Phase 2 (regression testing) and Phase 4 (code changes). Don't re-fix bugs that were already merged. If a reviewer suggested a different approach to a fix, use that approach instead.

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

### Phase 2: Tick Loop (48 ticks, ~4 hours)

For each tick (1 through 48):

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
   - **If zero NPCs are present, focus on solo gameplay** — test every action that works alone, try edge cases, note which systems are blocked by the lack of others
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

9. **If NPC dies** (alive=false in state), immediately spawn a new NPC and continue the loop with the remaining tick budget. Log the death and the new spawn. Do not end the playtest early on death — the playtest should run its full duration.

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

After documenting findings, make code changes in this priority order:

1. **Pull latest code** before making any changes:
   ```bash
   git checkout main
   git pull origin main
   ```
   This ensures fixes are based on the latest codebase, not the version from ~4 hours ago when the playtest started.

2. **Priority 1 — Fix bugs**: Read the relevant source code in `server/` and `client/`, find root causes, make targeted fixes. Bugs always come first.

3. **Priority 2 — Implement new features and game mechanics**: This is now a required part of every playtest, not optional. Each playtest should ship at least one new feature, action, mechanic, or piece of content. Good examples:
   - New actions (solo variants of existing social actions, new profession-specific actions, environmental interactions)
   - New mechanics (weather effects on gameplay, day/night behavior, seasons, NPC needs that don't exist yet)
   - New content (new location types, new item types, new recipes, new creatures, new events)
   - New systems (quests, reputation, skills, achievements, rumors, journals)
   - Solo-world features (anything that makes a lonely world interesting)

   Keep the feature scope small enough to implement and test in one session, but meaningful enough that it's a real change. Don't just add a stub — make it work and document how to interact with it.

4. **Priority 3 — Balance and polish**: UX fixes, text clarity, balance tweaks, design issue resolutions.

5. **Be conservative with the change, not the ambition**: Make clean, minimal changes to implement what you decided. Don't refactor unrelated code. If a fix or feature is too risky or complex, document it in the plan but don't attempt it.

6. **Do NOT make formatting-only changes**: never reformat, restyle, or rewrite files beyond the lines you are fixing. Do not change whitespace, alignment, brace style, or struct literal formatting in code you aren't modifying. Rewriting a file with cosmetic changes risks accidentally deleting methods or functions, which has broken the build before.

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

Play as a curious, thorough tester AND an ambitious game developer. Your goals:
- Try every action at least once if possible
- Visit multiple locations
- Interact with every NPC you encounter (if any)
- Test edge cases (what happens when stats are very low? when inventory is full? when the world is empty?)
- Pay attention to timing (do action durations match their descriptions?)
- Note anything that feels wrong, confusing, or could be better
- **Think like a developer, not just a tester**: what would make this game genuinely better? Build it.
- **Think about solo viability**: if you were the only person in this world, would it still be fun to play? If not, what would fix that?
