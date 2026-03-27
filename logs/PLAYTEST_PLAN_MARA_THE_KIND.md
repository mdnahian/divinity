# Playtest Plan — Mara the Kind / Elara the Compassionate

**Session:** 2026-03-27
**NPCs:** Mara the Kind (Healer, Ticks 1-8, DIED) -> Elara the Compassionate (Healer, Tick 9+)
**Last updated:** Final — Phase 3 Analysis

---

## Confirmed Bugs

### BUG-M1: All gather action Conditions check first location's resources (HIGH)
- **Evidence (code):** gather.go — forage (line 79), hunt (line 125), farm (line 168), fish (line 205), chop_wood (line 246), mine_stone (line 282), mine_ore (line 318), gather_thatch (line 354) all use `[type]s[0].Resources` to check resource availability.
- **Impact:** If the first location of that type has no resources, the action disappears globally for all NPCs even though other locations have plenty.
- **Fix:** Check current location resources first, or check any location of that type.

### BUG-M2: All gather action Execute functions drain first location's resources (HIGH)
- **Evidence (code):** gather.go Execute functions for forage (line 83-94), hunt (line 129-134), farm (line 172-185), fish (line 207-223), chop_wood (line 249-259), mine_stone (line 285-295), mine_ore (line 321-331), gather_thatch (line 357-371) all reference `[type]s[0].Resources`.
- **Impact:** Resources are always consumed from the first location, not where the NPC actually is. This silently drains the wrong location.
- **Fix:** Use `w.LocationByID(n.LocationID).Resources` instead.

### BUG-M3: Trade only works at first market (MEDIUM)
- **Evidence (code):** economy.go line 24: `n.LocationID != markets[0].ID`
- **Confirmed by gameplay:** Trade worked at Golden Meadow Market (first market). NPCs at other markets (East Side, Aethelgard, etc.) cannot trade.
- **Fix:** Check `n.LocationID` against current location type instead.

### BUG-M4: IsWorkerAtType only checks first location (MEDIUM)
- **Evidence (code):** world.go line 667: `w.IsWorkerAt(n, locs[0].ID)`
- **Impact:** Workers at the 2nd+ inn/forge/mill are not recognized. Affects cook, brew_ale, serve_customer, smelt_ore, forge_weapon, forge_tool.
- **Fix:** Iterate all locations of that type.

### BUG-M5: Inn table fee only charged at first inn (MEDIUM)
- **Evidence (code):** registry.go line 149: `inn := inns[0]`
- **Impact:** NPCs at other inns perform actions for free, breaking inn economy.
- **Fix:** Check NPC's current location instead.

### BUG-M6: mill_grain only works at first mill (LOW)
- **Evidence (code):** craft.go line 218: `mills[0]` used for both worker check and owner check.
- **Fix:** Check current location or iterate all mills.

### BUG-M7: copy_text only checks first library ownership (LOW)
- **Evidence (code):** craft.go line 329: `libs[0].OwnerID`
- **Fix:** Check current location.

### BUG-M8: tan_hide always travels to first dock (LOW)
- **Evidence (code):** craft.go line 165: `docks[0].ID`
- **Fix:** Use destNearestOfType("dock") instead.

## Design/Balance Issues

### DESIGN-M1: Forest enemy presence not communicated in action descriptions (HIGH)
- **Evidence:** Mara died at Whispering Forest Edge after foraging. No warning shown about wolf/bear presence. 5+ NPCs already died there.
- **Suggestion:** Add enemy warnings to forage/hunt action descriptions, or add a Conditions check that warns about enemies.

### DESIGN-M2: Flee action too slow to save NPCs (HIGH)
- **Evidence:** Mara's flee_area had 32 ticks of travel remaining when she died. Enemies attack every tick.
- **Suggestion:** Either reduce flee travel time, add damage reduction during flee, or make flee immediate with a fatigue/stress penalty.

### DESIGN-M3: Healer profession lacks early-game path to medicine (MEDIUM)
- **Evidence:** heal_patient requires medicine; brew_potion requires 2 herbs; herbs only come from foraging (15% chance per forage). A new healer cannot practice their profession.
- **Suggestion:** Start healers with 2 herbs, or add a dedicated "gather herbs" action at forests/gardens.

### DESIGN-M4: Scavenge overpowered at death locations (LOW)
- **Evidence:** At Whispering Forest Edge, 350+ gold and many items available from dead NPCs. One scavenge picked up 17 gold, 8 fish, 1 herbs, 2 nets, 3 hooks.
- **Observation:** Not necessarily a bug — creates an interesting risk/reward dynamic at dangerous locations.

### DESIGN-M5: Social need decays constantly even when alone (OBSERVATION)
- **Evidence:** Social dropped from 62 to 43 over ~16 ticks while alone. Rate is about 1.2 per tick.

## Improvement Ideas

1. **Gather actions should use current location resources** — the single biggest systemic bug
2. **Trade should work at any market the NPC is at** — simple fix, high impact
3. **IsWorkerAtType should check all locations** — simple fix
4. **Inn table fee should check current location** — simple fix
5. **Add enemy warnings to gather/hunt actions** — safety improvement
6. **Reduce flee travel time** — balance fix

## New Feature Ideas

1. **Dedicated herb-gathering action** for healers at forests/gardens
2. **Enemy danger level** shown in location descriptions and travel options
3. **Group flee** — NPCs at same location can flee together

## Regression Check

| Previous Bug | Status |
|---|---|
| BUG-1: gather actions use wrong location | **STILL PRESENT** — confirmed in code |
| BUG-2: No solo eat action | **FIXED!** — eat works, bread gives +25 hunger |
| BUG-3: explore wrong travel time | **IMPROVED** — explore shows travel time matching actual distance |
| BUG-4: sleep ignores home ownership | **UNABLE TO FULLY TEST** — sleep does go to HomeID. Need to test inn vs home path |
| BUG-5: Fatigue rises during sleep | **FIXED!** — code shows fatigue -0.5/tick during sleep |
| BUG-J3: Session token expiry | **NOT OBSERVED** — 20+ ticks without token loss |
| BUG-J4: Duplicate world events | **NOT OBSERVED** this session |
| Trade only at first market | **STILL PRESENT** in code |
| IsWorkerAtType first location only | **STILL PRESENT** in code |
| Inn table fee first inn only | **STILL PRESENT** in code |
| mill_grain/copy_text first location | **STILL PRESENT** in code |

## Working Mechanics (Confirmed Good)
- eat: +25 hunger from bread, consumes 1 item
- drink: Fully restores thirst at well, consumes 1 water
- talk: +20 social, +3 relationship, +3 happiness
- trade: Works at first market, +3-5 gold, +5 relationship
- farm: +4 wheat, +12 fatigue
- gather_clay: Works at location with clay resources
- explore: Random nearby location, small chance of finding items
- scavenge: Picks up all ground items at location
- forage: +1-3 berries, 15% chance herbs, +8 fatigue
- flee_area: Works but very slow travel time
- sleep: Available at fatigue >50%, travels to home/inn
- Weather system: clear -> rain -> cloudy transitions
- Memory system: Tracks actions, defining moments
- Relationship tracking: Increments correctly from social actions
- Enemy attacks: Consistent damage per tick at locations with enemies
- HP regen: +1/tick natural healing
- NPC death: Correctly tracked, drops items on ground

## Priority Ranking

| ID | Issue | Priority | Impact |
|----|-------|----------|--------|
| BUG-M1/M2 | Gather actions wrong location resources | **HIGH** | Breaks all resource gathering |
| BUG-M3 | Trade only at first market | **MEDIUM** | Breaks economy at most markets |
| BUG-M4 | IsWorkerAtType first location only | **MEDIUM** | Breaks profession actions at 2nd+ locations |
| BUG-M5 | Inn table fee first inn only | **MEDIUM** | Breaks inn economy |
| DESIGN-M1 | No enemy warning on gather/hunt | **HIGH** | NPCs die unknowingly |
| DESIGN-M2 | Flee too slow | **HIGH** | Flee doesn't save NPCs |
| BUG-M6/M7/M8 | mill/copy/tan first location | **LOW** | Minor location bugs |
