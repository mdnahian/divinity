# Playtest Plan — Vala Greenleaf (FINAL)

**Session:** 2026-03-25
**NPC:** Vala Greenleaf (Carpenter, Age 50)
**Route:** Soggy Keel → Drowned Lantern Inn → Echo Well → Fog Market
**Duration:** 24 ticks (~2 hours), Game Ticks 1097–1236, Day 5 01:25–13:00
**Last updated:** Final

---

## Confirmed Bugs

### BUG-V1: Duplicate world events in observe (LOW)
- **Evidence:** Tick 1 — "Strange omens in the sky" appeared TWICE with different text in single observe output.
- **Expected:** Events should be deduplicated by title or ID.
- **Same as:** BUG-J4. STILL PRESENT.
- **Likely files:** `server/api/views.go` or `server/engine/engine.go` (event dedup)

### BUG-V2: Resource mismatch — gather_clay (LOW)
- **Evidence:** Tick 6 — Location clay 10→9 (-1), but NPC received +2 clay in inventory.
- **Expected:** Resource decrease should match NPC gain.
- **Likely files:** `server/action/gather.go`

### BUG-V3: NPC profession name inconsistency (LOW)
- **Evidence:** Ulric Deeproot shown as "healer" at Soggy Keel (ticks 1-6), "herbalist" at Drowned Lantern Inn (tick 12).
- **Expected:** Profession display should be consistent.
- **Likely files:** `server/npc/professions.go`, `server/api/views.go`

### ~~BUG-V4: Sleep action missing when AT an inn~~ — NOT A BUG
- **Evidence:** Ticks 12-14 — At Drowned Lantern Inn, sleep NOT available. Available at Echo Well and Soggy Keel.
- **Root cause:** Sleep requires fatigue > 50% (survival.go:92). At the inn, fatigue was 46% (just rested). At other locations fatigue was 53-54%. Threshold working correctly.

### BUG-V5: Action durations significantly longer than listed (HIGH)
- **Evidence:**
  - drink at "(here)": listed ~15 min, actual ~75 game-min (5x)
  - talk: listed ~15 min, actual ~50 game-min (3.3x)
  - buy_food + 40 min travel: listed ~55 min, actual ~180 game-min (3.3x)
  - craft_pottery: listed ~45 min, actual ~8 ticks (~40 game-min, roughly correct)
- **Impact:** Players can't plan. Agent scheduling unreliable. Most actions take 3-5x longer than advertised.
- **Likely files:** `server/engine/turns.go`, `server/action/registry.go` (duration → tick conversion)

### BUG-V6: buy_food fails — "Nobody at the market had food to sell" (MEDIUM)
- **Evidence:** Tick 24 — Traveled 35 ticks to Fog Market, action completed but no food purchased. "Nobody at the market had food to sell." No gold spent.
- **Impact:** Players waste significant time traveling to market for nothing. Action should not be listed if no food available.
- **Likely files:** `server/action/economy.go` (buy_food Execute and Candidates)

## Design/Balance Issues

### DESIGN-V1: Carpenter has no profession-specific actions (MEDIUM)
- 24 ticks with carpenter: 58 skill, zero carpenter actions available. Same issue as blacksmith (DESIGN-J1). Professions are cosmetic only.

### DESIGN-V2: Explore travel time fluctuates randomly every tick (LOW)
- Values observed: +10, +5, +25, +25, +30, +20, +30, +45, +60 min. No consistency.

### DESIGN-V3: Social drains constantly ~2.5/tick even when alone (OBSERVATION)
- Social went 31→19% over 24 ticks despite 4 talk actions (+20 each = +80 total). Net loss = -12%.

### DESIGN-V4: Sleep duration 240 min is impractical (OBSERVATION)
- At observed tick rates, sleep would lock NPC for 48+ ticks. Rest (-20 fatigue for 1g) is vastly preferable.

### DESIGN-V5: Fishing caught nothing at skill 0 (OBSERVATION)
- Despite 20 fish at location. Possibly skill-gated, making fishing useless for beginners.

### DESIGN-V6: NPCs at inn permanently miserable (OBSERVATION)
- All 3 NPCs at Drowned Lantern Inn were "miserable" across multiple ticks. Didn't improve after social interaction.

## Working Mechanics (Confirmed Good)

- **talk:** +20 social, +3 relationship, +3 happiness. Consistent across 4 NPCs.
- **gather_clay:** Local gathering works, +2 clay, "(here)" displays correctly.
- **craft_pottery:** clay x2 → ceramic x1. Good conditional gating by inventory.
- **rest:** -20 fatigue, -8 stress, 1 gold. Works correctly at inns.
- **drink:** Fully restores thirst to 100%. Free at wells.
- **explore:** Successfully discovers new locations. Moved between 3 locations.
- **Action gating:** rest/sleep by fatigue threshold, craft by inventory. Correct behavior.
- **Memory system:** "Recent" and "Defining moments (vivid)" categories. Well-structured.
- **Resource regeneration:** Clay regenerated between ticks.
- **Travel system:** NPC successfully moved across 4 locations.
- **Token persistence:** Token worked for full 120-minute session (24 ticks). BUG-J3 FIXED.

## Regression Check (vs Juna Ironhand 2026-03-24)

| Previous Bug | Status |
|---|---|
| BUG-J1: gather_clay no location names | **FIXED** — Shows full location names now |
| BUG-J2: repair_building with no locations | UNABLE TO TEST — never appeared |
| BUG-J3: Token expiry after ~35 min | **FIXED** — Token lasted 120+ min |
| BUG-J4: Duplicate world events | **STILL PRESENT** (BUG-V1) |
| DESIGN-J1: Blacksmith no profession actions | N/A — but carpenter also affected (DESIGN-V1) |

## Priority Ranking

| ID | Issue | Priority | Impact |
|----|-------|----------|--------|
| BUG-V5 | Action durations 3-5x longer than listed | **HIGH** | All planning broken |
| BUG-V4 | Sleep missing at inn | **MEDIUM** | Can't sleep where expected |
| BUG-V6 | buy_food fails at market | **MEDIUM** | Wasted travel, no food |
| DESIGN-V1 | Professions have no unique actions | **MEDIUM** | Identity broken |
| BUG-V1 | Duplicate world events | **LOW** | Display clutter |
| BUG-V2 | Resource mismatch gather_clay | **LOW** | Economy tracking |
| BUG-V3 | Profession name inconsistency | **LOW** | Display confusion |
| DESIGN-V2 | Explore travel time random | **LOW** | Confusing UX |

## Session Summary

Played 24 ticks as Vala Greenleaf the carpenter. Tested 9 unique actions across 4 locations. Found 6 bugs (1 HIGH, 2 MEDIUM, 3 LOW) and 6 design issues. Key finding: action durations are systematically 3-5x longer than listed, which is the most impactful bug. Token expiry from previous playtest is fixed. Duplicate world events persist.
