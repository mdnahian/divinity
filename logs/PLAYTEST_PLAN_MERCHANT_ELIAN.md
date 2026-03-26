# Playtest Plan — Merchant Elian

**Session:** 2026-03-26
**NPC:** Merchant Elian (Merchant at Golden Meadow Market)
**Personality:** empathetic, curious, brave, creative, unscrupulous, introverted
**Last updated:** Final — 17 ticks played (~1.5 hours real time)

---

## Confirmed Bugs

### BUG-M1 (HIGH): Gather actions check/deplete resources at first location only
- **Evidence:** Code review of `server/action/gather.go`. All 8 gather actions (forage, hunt, farm, fish, chop_wood, mine_stone, mine_ore, gather_thatch) use `LocationsByType(type)[0].Resources` in both Conditions and Execute functions.
- **Pattern:** `Destination` correctly routes NPC to nearest location via `destNearestOfType`, but `Execute` reads/writes resources from the FIRST location globally.
- **Example:** NPC at forest #3 forages -> berries are checked/depleted at forest #1.
- **Expected:** Execute should use `w.LocationByID(n.LocationID)` to operate on the NPC's current location.
- **Same as BUG-E1 from Eldric playtest. Still present.**
- **FIX APPLIED:** Changed all 8 gather actions to use `loc := w.LocationByID(n.LocationID)` in Execute and iterate all locations in Conditions.
- **File:** `server/action/gather.go`

### BUG-M2 (MEDIUM): Trade only works at first market
- **Evidence:** `server/action/economy.go` line 24: `n.LocationID != markets[0].ID` — condition only passes for NPCs at the first market.
- **Verified during gameplay:** Trade worked at Golden Meadow Market (which happened to be the first market). Would fail at East Side Market, Aethelgard Market, etc.
- **Same as BUG-E2. Still present.**
- **FIX APPLIED:** Changed to check `loc.Type != "market"` using the NPC's current location.
- **File:** `server/action/economy.go`

### BUG-M3 (MEDIUM): IsWorkerAtType only checks first location of a type
- **Evidence:** `server/world/world.go` line 667: `w.IsWorkerAt(n, locs[0].ID)` — only checks if NPC is a worker at the first location of that type.
- **Impact:** Affects cook, brew_ale, smelt_ore, forge_weapon, forge_tool, serve_customer actions. Workers at secondary forges/inns lose their profession-specific actions.
- **Same as BUG-E3. Still present.**
- **FIX APPLIED:** Changed to iterate all locations: `for _, loc := range locs { if w.IsWorkerAt(n, loc.ID) { return true } }`.
- **File:** `server/world/world.go`

### BUG-M4 (MEDIUM): Inn table fee only charged at first inn
- **Evidence:** `server/action/registry.go` lines 147-149: `inn := inns[0]` then checks `n.LocationID == inn.ID`. Only charges NPCs at the first inn.
- **Impact:** Economic imbalance — first inn overtaxed, other inns free.
- **FIX APPLIED:** Changed to check if NPC's current location is an inn: `loc := w.LocationByID(n.LocationID); if loc != nil && loc.Type == "inn"`.
- **File:** `server/action/registry.go`

### BUG-M5 (MEDIUM): Force-interrupt clears resume state on new action
- **Evidence:** `server/engine/turns.go` line 297: `n.ResumeActionID = ""` clears resume state whenever any new non-resume action is submitted.
- **Impact:** If NPC is interrupted by enemies and then does something else (eat, flee), their interrupted action is lost. Sleep interrupted by combat can never be resumed.
- **Same as BUG-E4. Still present. NOT FIXED** — too risky to change without more testing.
- **File:** `server/engine/turns.go`

### BUG-M6 (LOW): mill_grain only works at first mill
- **Evidence:** `server/action/craft.go` line 218: only checks `mills[0]`.
- **FIX APPLIED:** Changed to iterate all mills.
- **File:** `server/action/craft.go`

### BUG-M7 (LOW): copy_text only works at first library
- **Evidence:** `server/action/craft.go` lines 328-331: only checks `libs[0]`.
- **FIX APPLIED:** Changed to iterate all libraries.
- **File:** `server/action/craft.go`

## Design/Balance Issues

### DESIGN-M1: go_home only available at night (MEDIUM)
- **Evidence:** Tick 16 — NPC at South Gate Forge with fatigue 55%, go_home not available because it's 10:40 AM. Code: `w.IsNight() && n.LocationID != n.HomeID`.
- **Impact:** NPCs can't return home to rest during the day, forcing them to use inns or rough-sleep.
- **FIX APPLIED:** Also allow go_home when fatigue >= 60%.
- **File:** `server/action/movement.go`

### DESIGN-M2: Large amounts of gold on ground (LOW)
- **Evidence:** 48 gold + 2 bread on the ground at Golden Meadow Market at spawn.
- **Likely cause:** NPCs dying or economic overflow dropping items.
- **Impact:** Scavenge gives free massive gold early, breaks early economy.
- **Suggestion:** Add gold decay on ground items, or cap scavenge gold pickup.

### DESIGN-M3: Social drains fast when alone (OBSERVATION)
- **Evidence:** Social dropped from 51% to 31% over ~16 ticks while mostly alone in wilderness.
- **Impact:** NPCs who travel to gather resources become socially starved. Could force them to make suboptimal choices.
- **Suggestion:** Reduce social drain rate when busy with actions, or add a "solitude tolerance" trait.

### DESIGN-M4: Fatigue accumulates quickly from travel + actions (OBSERVATION)
- **Evidence:** Fatigue went from 28% to 65% over 16 ticks (2.3%/tick average). Multiple actions + travel fatigue compounds.
- **Impact:** NPCs need to sleep every ~30 ticks or so, spending a significant portion of time sleeping.

## Improvement Ideas

1. **Show travel fatigue in action listings** — Add "(+X fatigue from travel)" to help decision-making
2. **Add destination name to explore** — "Wander and explore" should mention the potential area
3. **Comfort should show mood change** — "Comforted X (stress went from 80 to 65)" instead of just "-15 stress"

## New Feature Ideas

1. **Rest at home action** — Shorter than sleep, just sit and reduce fatigue by ~15. Available when at home, fatigue > 30%.
2. **Merchant-specific selling bonus** — Merchants should get better prices or sell more per action.
3. **Scavenge gold limit** — Cap gold pickup at e.g. 10 per scavenge to prevent economy breaks.

## Regression Check

| Previous Bug | Status |
|---|---|
| BUG-E1: Gather actions use wrong location | **STILL PRESENT — FIX APPLIED** |
| BUG-E2: Trade only works at first market | **STILL PRESENT — FIX APPLIED** |
| BUG-E3: IsWorkerAtType only checks first location | **STILL PRESENT — FIX APPLIED** |
| BUG-E4: Force-interrupt clears resume state | **STILL PRESENT — NOT FIXED (too risky)** |
| BUG-J1: gather_clay no location names | **FIXED** — Shows all dock/well locations with travel times |
| BUG-J2: repair_building appears with no locations | UNABLE TO TEST |
| BUG-J3: Session token expiry | **NOT TRIGGERED** — Session active for ~90 min (17 ticks) |
| BUG-J4: Duplicate world events | UNABLE TO TEST — only one event observed |
| BUG-2 (No solo eat): | **FIXED** — Eat works, bread +25 hunger |
| Fatigue rises during sleep | **FIXED in code** — DecayNeeds reduces fatigue during sleep. Travel fatigue still applies while en route. |
| Sleep ignores home ownership | **FIXED in code** — sleep.Destination checks HomeID first |
| Memory "social -20" confusing | **FIXED** — Memory shows "social +20" correctly |
| Talk cooldown invisible | **LIKELY FIXED** — talk still available after talking (no invisible cooldown) |
| Business ownership no benefit | UNABLE TO TEST |

## Working Mechanics (Confirmed)

- **comfort:** -15 stress, -3 trauma, +10 relationship. Requires Empathy >= 45.
- **scavenge:** Picks up ALL ground items. Correctly removes from ground.
- **trade:** Sells 1 item at market price. +5 relationship. +0.3 merchant skill.
- **talk:** +20 social, +3 relationship, +3 happiness. Custom dialogue stored in memory and shared with target.
- **eat:** Bread +25 hunger, wheat +8 hunger (from Juna playtest). Consumes 1 item.
- **drink:** Fully restores thirst to 100%. Uses 1 well water.
- **gather_clay:** +1 clay, +10 fatigue. Works at locations with clay resources. Shows "(here)" correctly.
- **explore:** Random nearby location within 60 units. 20% chance to find item. +5 fatigue.
- **forage:** +1-3 berries, +8 fatigue. 15% chance for herbs. Works at forests.
- **sleep:** Routes to home (if owned), then inn, then rough-sleep. 240 game min. Reduces fatigue -40 on completion.
- **Location awareness:** "(here)" labels, travel times, location descriptions all correct.
- **Weather system:** Transitions observed: storm -> cloudy -> clear -> rain.
- **World events:** Atmospheric text about wildlife threats.
- **Memory system:** Correct action logging, "vivid" marker for important events, "Defining moments" section.
- **Relationship tracking:** Labels progress: neutral -> slight positive -> acquaintance.

## Priority Ranking

| ID | Issue | Priority | Fix Status |
|----|-------|----------|------------|
| BUG-M1 | Gather wrong location resources | **HIGH** | FIXED |
| BUG-M2 | Trade first market only | **MEDIUM** | FIXED |
| BUG-M3 | IsWorkerAtType first location | **MEDIUM** | FIXED |
| BUG-M4 | Inn fee first inn only | **MEDIUM** | FIXED |
| BUG-M5 | Resume state lost | **MEDIUM** | NOT FIXED |
| BUG-M6 | mill_grain first mill only | **LOW** | FIXED |
| BUG-M7 | copy_text first library only | **LOW** | FIXED |
| DESIGN-M1 | go_home only at night | **MEDIUM** | FIXED |
