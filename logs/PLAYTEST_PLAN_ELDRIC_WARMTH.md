# Playtest Plan — Eldric Warmth

**Session:** 2026-03-25
**NPC:** Eldric Warmth (Innkeeper at Golden Meadow Market)
**Personality:** wise, neurotic
**Last updated:** Final — Tick 13 (game tick ~2011)

---

## Confirmed Bugs

### BUG-E1: Gather actions use `[0]` index instead of NPC's actual location (HIGH)
- **Evidence:** Code review of `server/action/gather.go`. All gather-type actions (forage, hunt, farm, fish, chop_wood, mine_stone, mine_ore, gather_thatch) check resources at `LocationsByType()[0]` and modify resources at `[0]` in both `Conditions` and `Execute` functions. But `Destination` sends the NPC to `destNearestOfType()` which may be a DIFFERENT location.
- **Example:** `forage` checks `forests[0].Resources["berries"]` (line 79-80, 85-93) but NPC travels to nearest forest via `destNearestOfType("forest")`. If there are 3 forests and the NPC goes to forest[2], resources are still deducted from forest[0].
- **Expected:** Should use `w.LocationByID(n.LocationID)` in Execute, like `gather_clay` does correctly.
- **Affected actions:** forage (berries/herbs), hunt (game), farm (wheat), fish (fish), chop_wood (wood), mine_stone (stone), mine_ore (iron_ore), gather_thatch (thatch)
- **Fix location:** `server/action/gather.go` — lines 79-93, 125-142, 168-185, 205-223, 246-259, 282-295, 318-331, 354-367

### BUG-E2: `trade` action only works at first market in world (MEDIUM)
- **Evidence:** `server/action/economy.go` line 24: `if n.LocationID != markets[0].ID { return false }`. This checks if the NPC is at `markets[0]` specifically. NPCs at any other market cannot trade.
- **Expected:** Should check if NPC is at ANY market: iterate `markets` and check if `n.LocationID` matches any.
- **Contrast:** `buy_food` (line 160) and `buy_supplies` (line 208) correctly iterate all markets.
- **Fix location:** `server/action/economy.go` lines 20-26

### BUG-E3: `IsWorkerAtType` only checks first location of type (MEDIUM)
- **Evidence:** `server/world/world.go` line 667: `return w.IsWorkerAt(n, locs[0].ID)`. Only checks if NPC is a worker at the first inn/forge/mill.
- **Impact:** Actions requiring `IsWorkerAtType` (serve_customer, cook, brew_ale, smelt_ore, forge_weapon, forge_tool, mill_grain) only work at the first location of that type, not at any location of that type.
- **Expected:** Should iterate all locations of that type.
- **Fix location:** `server/world/world.go` lines 662-668

### BUG-E4: Force-interrupt loses resume state (MEDIUM)
- **Evidence:** In `SubmitExternalAction` (server/engine/turns.go), force-interrupting an action sets `n.ResumeActionID` (line 247) and `n.ResumeTicksLeft` (line 248), but lines 297-298 unconditionally clear them: `n.ResumeActionID = ""` and `n.ResumeTicksLeft = 0`.
- **Tested:** Force-interrupted work_together to eat. API response correctly showed `interrupted_action: "work_together"` and `resume_ticks_left: 5`, but server state showed `resume_action: ""`.
- **Expected:** Resume state should persist so NPC can resume interrupted action later.
- **Fix location:** `server/engine/turns.go` lines 297-298 — should be conditional on `!force` or should only clear if the force block didn't run.

## Design/Balance Issues

### DESIGN-E1: Mass NPC misery (MEDIUM)
- **Evidence:** At Golden Meadow Market, 6 of 7 NPCs were "miserable" throughout the session. Countess Zarah deteriorated from "anxious" to "miserable" during the session.
- **Root cause:** NPC stats drain continuously (hunger -0.069/tick, thirst -0.116/tick, social +0.5/tick) but autonomous NPCs (tier 3, rule-based) may not be effectively managing their needs.
- **Suggestion:** Ensure tier-3 NPC AI prioritizes critical needs more aggressively, or reduce drain rates for unattended NPCs.

### DESIGN-E2: Slow server tick rate limits playtesting (OBSERVATION)
- **Evidence:** Config shows `TickIntervalMs: 60000` (60 seconds) but actual observed rate was ~300 seconds per tick (5 min). A 30 game-minute action (6 ticks) took ~30 real minutes.
- **Impact:** The 24-tick playtest procedure requires ~2 hours but at this rate would take 6+ hours.

### DESIGN-E3: Innkeeper has no profession-specific actions at market (LOW)
- **Evidence:** Eldric Warmth (innkeeper) at Golden Meadow Market had no innkeeper-specific actions available. `serve_customer`, `cook`, `brew_ale` all require `IsWorkerAtType("inn")` which needs the NPC to be at an inn AND be a worker/owner there.
- **Impact:** An innkeeper spawned at a market has no profession-relevant actions until they travel to an inn and either get hired or start a business.
- **Suggestion:** Consider adding a generic "practice profession" action or making profession actions available at any location.

## Improvement Ideas

1. **Fix gather actions to use NPC's actual location** — Most impactful bug fix. Would ensure resource depletion happens at the correct location.
2. **Fix trade to work at any market** — Simple one-line fix with high impact.
3. **Fix IsWorkerAtType to check all locations** — Simple fix enabling profession actions at any location of that type.
4. **Preserve resume state on force-interrupt** — Add a conditional to not clear resume state when force was used.

## New Feature Ideas

1. **Innkeeper profession actions at market** — Allow innkeepers to "set up a temporary stall" or "offer advice on local inns" at markets.
2. **NPC need indicator in observe** — Show hunger/thirst/fatigue levels for observed NPCs so players can better target comfort/gift actions.

## Working Mechanics (Confirmed Good)

- **scavenge:** Picks up all ground items correctly. Gold x13 picked up successfully.
- **offer_counsel:** Wise NPCs can counsel stressed NPCs for 1 gold. Correct stat changes (-15 stress, +5 happiness to target, +1 gold to counselor).
- **eat (solo):** Available with food in inventory (bread x2). Confirmed from Juna's previous session - still working.
- **work_together:** Available when same-profession NPC is nearby. Correctly checks profession match.
- **Force interrupt:** API correctly returns interrupted action info and resume ticks. (Server-side resume is broken though.)
- **gather_clay:** Now shows full location list with travel times. Fixed from previous playtest.
- **Event deduplication:** observe tool now deduplicates events by name (seen map). Fixed from previous playtest.
- **Token persistence:** Token remained valid throughout entire session (~90+ real minutes). BUG-J3 not reproduced.
- **Stat decay:** Matches config values precisely. Hunger -0.069/tick, Thirst -0.116/tick, Fatigue +0.1/tick.
- **Location awareness:** "(here)" displays correctly for actions at current location.
- **Memory system:** Correctly logs actions with dialogue, stat changes, and timestamps.
- **Relationship tracking:** Correctly adjusts after social interactions (offer_counsel: +4).

## Regression Check (vs Juna Ironhand 2026-03-24)

| Previous Bug | Status |
|---|---|
| BUG-J1: gather_clay no location names | **FIXED!** — Now shows full location list with travel times |
| BUG-J2: repair_building appears with no locations | UNABLE TO TEST — repair_building did not appear (no owned buildings) |
| BUG-J3: Session token expired unexpectedly | **NOT REPRODUCED** — Token valid for 90+ min. Likely was a server restart. |
| BUG-J4: Duplicate world event text | **FIXED!** — observe tool now deduplicates events by name |
| BUG-3 (Ulric): explore wrong travel time | UNABLE TO TEST — did not use explore |
| BUG-4 (Ulric): sleep ignores home ownership | **FIXED** (code review) — sleep Destination now checks HomeID first, then falls back to inn/rough sleep |
| BUG-5 (Ulric): Fatigue rises during sleep | **FIXED** (code review) — DecayNeeds now checks `sleeping` state and reduces fatigue by 0.5/tick instead of increasing |
| DESIGN-J1: Blacksmith no profession actions | N/A — innkeeper this session |
| DESIGN-2 (Ulric): talk cooldown invisible | UNABLE TO TEST — did not test consecutive talks |
| DESIGN-6 (Ulric): Business has no benefit | UNABLE TO TEST — did not start business |

## Priority Ranking

| ID | Issue | Priority | Impact |
|----|-------|----------|--------|
| BUG-E1 | Gather actions use [0] index | **HIGH** | All gather resources deplete wrong location |
| BUG-E4 | Force-interrupt loses resume | **MEDIUM** | Resume feature is non-functional |
| BUG-E2 | Trade only at first market | **MEDIUM** | Trade broken at most markets |
| BUG-E3 | IsWorkerAtType only checks [0] | **MEDIUM** | Profession actions limited to first location |
| DESIGN-E1 | Mass NPC misery | **MEDIUM** | Most NPCs miserable |
| DESIGN-E3 | Innkeeper no actions at market | **LOW** | Profession identity weak at spawn |
