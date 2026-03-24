# Playtest Plan — Juna Ironhand / Sable Brightwater

**Session:** 2026-03-24
**NPCs:** Juna Ironhand (Farmer, Ticks 1-7) → Sable Brightwater (Blacksmith, Tick 8)
**Last updated:** Final — Tick 8

---

## Confirmed Bugs

### BUG-J1: `gather_clay` shows no location names from non-clay locations (LOW)
- **Evidence:** From Juna's Home and The Gritroot Farm, `gather_clay` shows "+5 min travel" with NO location names listed (unlike drink/farm/fish which list all locations with names).
- **Observed:** Ticks 1-4 consistently. At The Boneridge Well (clay: 8), it correctly shows "(here)".
- **Expected:** Should list the destination name like other actions.

### BUG-J2: `repair_building` appears with no locations (LOW)
- **Evidence:** Action appeared at Tick 4 but lists no locations. NPC doesn't own a building.
- **Expected:** Shouldn't appear if NPC has nothing to repair.

### BUG-J3: NPC session token expired unexpectedly (HIGH)
- **Evidence:** After ~35 minutes of active play (7 ticks), Juna's token returned "unknown token". No warning.
- **Impact:** Critical for agent-based gameplay — any long-running session can lose progress.

### BUG-J4: Duplicate world event text in observe (LOW)
- **Evidence:** "Strange omens in the sky" appeared TWICE with identical text in Sable's observe output at Tick 8.
- **Expected:** Events should be deduplicated.

## Design/Balance Issues

### DESIGN-J1: Blacksmith has no profession-specific actions (MEDIUM)
- **Evidence:** Sable (blacksmith) has iron ore x4 and iron ingot x2 in inventory but NO smelt/forge actions available.
- **Impact:** Profession identity is broken — a blacksmith can't do blacksmithing.
- **Suggestion:** Add `smelt` (ore→ingot at forge) and `forge` (ingot→tools/weapons) actions at forge locations.

### DESIGN-J2: Fatigue increases rapidly during physical work (OBSERVATION)
- Farming: +12 fatigue. Gather clay: +11 fatigue. After 6 ticks: 23%→49%.

### DESIGN-J3: Social drain constant even when alone (OBSERVATION)
- Social dropped 68%→54% over 6 ticks while alone (~2.3/tick).

### DESIGN-J4: `gather_clay` UX — no destination shown (LOW)
- Unlike other actions, gather_clay doesn't tell you WHERE it will take you. Shows "+5 min travel" with no location name.

## Working Mechanics (Confirmed Good)

- **eat (solo):** +8 hunger from wheat, consumes 1 from inventory. **NEW since last playtest!**
- **farm:** +4 wheat, +2 gold (wage), +12 fatigue, -5 stress. Good loop.
- **gather_clay:** Finds nearest clay, travels automatically. +1 clay.
- **talk:** +20 social, +3 relationship, +3 happiness. No immediate cooldown observed.
- **drink:** Available at wells, shows "(here)" correctly.
- **Location awareness:** "(here)" displays correctly. go_home appears when away from home.
- **teach:** New action discovered — 2 gold lesson fee.
- **eat_together:** Available with nearby NPCs.
- **steal:** Available as RISKY action.
- **start_business:** Available at manors/establishments.
- **World events:** Wolves, omens working (but duplicated).
- **Weather transitions:** Clear → cloudy observed.
- **Memory system:** Correctly logs actions with stat changes.
- **Relationship tracking:** "slight positive (talk)" after chatting.

## Regression Check (vs Ulric Greenleaf 2026-03-22)

| Previous Bug | Status |
|---|---|
| BUG-1: gather_clay ignores local resources | **IMPROVED** — Shows "(here)" at clay locations now. But no location names from other spots. |
| BUG-2: No solo eat action | **FIXED!** — `eat` action works correctly. |
| BUG-3: explore wrong travel time | UNABLE TO TEST — session lost before testing |
| BUG-4: sleep ignores home ownership | UNABLE TO TEST |
| BUG-5: Fatigue rises during sleep | UNABLE TO TEST |
| DESIGN-1: Memory text "social -20" confusing | **FIXED!** — Now shows "social +20" correctly. |
| DESIGN-2: talk cooldown invisible | **POSSIBLY FIXED** — talk still available immediately after talking (no cooldown observed in 1 tick) |
| DESIGN-3: Sleep duration too long | UNABLE TO TEST |
| DESIGN-6: Business has no benefit | UNABLE TO TEST |

## Priority Ranking

| ID | Issue | Priority | Impact |
|----|-------|----------|--------|
| BUG-J3 | Session token expiry | **HIGH** | Breaks long-running agent sessions |
| BUG-J4 | Duplicate world events | **LOW** | Display clutter |
| BUG-J1 | gather_clay no location names | **LOW** | UX confusion |
| BUG-J2 | repair_building shows when not applicable | **LOW** | Minor UX |
| DESIGN-J1 | Blacksmith no profession actions | **MEDIUM** | Profession identity broken |
