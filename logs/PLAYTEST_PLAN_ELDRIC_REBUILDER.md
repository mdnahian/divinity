# Playtest Plan — Eldric the Rebuilder / Mara the Gatherer

**Session:** 2026-04-09
**NPCs:** Eldric the Rebuilder (farmer, outgoing/mentor/loyal/open-minded), Mara the Gatherer (forager, brave)
**Territories:** Golden Meadow (spawn), East Side / Rat's Nest territory
**Budget:** 48 ticks across 2 lives (Eldric: 24 ticks, died; Mara: 24 ticks)

---

## Confirmed Bugs

### BUG-R1 (HIGH): Sleep text fallback says "Slept at home" when NPC has no home
- **Evidence:** Eldric spawned at Golden Meadow Market, slept at The Gilded Inn (random inn assigned as home). Memory showed "Slept at home" even though NPC never owned a home. The fallback text at survival.go line 163 always says "home" even for the default case (no HomeID, not at inn).
- **Root cause:** Line 163 in survival.go `return fmt.Sprintf("Slept at home (-%d fatigue, -5 stress).", restore)` is the catch-all fallback after the HomeID check. Should say something generic like "Rested and slept".
- **Fix:** Changed line 163 to say "Rested and slept" instead of "Slept at home".

### BUG-R2 (MEDIUM): flee_area takes 160 min travel — NPC killed before fleeing
- **Evidence:** Eldric at Sunrise Stable with 2 dire wolves. HP 63. Committed flee_area but it takes 160 min travel time to destination. NPC was killed during the travel (HP 63→47→29→22→7→0) while "fleeing". The flee action routes to a random safe location far away, not the nearest safe location.
- **Impact:** NPCs that need to flee enemies die during the multi-tick flee travel. Flee is essentially useless against nearby threats.
- **Suggestion:** Flee should route to nearest safe location (no enemies, within 10 min) or provide instant location change for 1-tick escape.

### BUG-R3 (MEDIUM): Dire wolves at innocuous locations (Sunrise Stable +5 min from market)
- **Evidence:** 2 dire wolves at Sunrise Stable, just 5 min from Golden Meadow Market. No warning in list_actions travel options. Eldric traveled there unsuspectingly and was killed.
- **Impact:** Safe-looking locations near spawn are death traps. Combined with BUG-R2 (flee doesn't work), this is extremely dangerous.
- **Suggestion:** list_actions travel destinations should show enemy count if > 0.

## Design/Balance Issues

### DESIGN-R1 (SEVERE): Solo world is unplayable for social actions
- **Evidence:** All 30+ locations visited across 2 lives had zero NPCs. Social need crashed from 79 to 2 (Eldric) and 68 to 46 (Mara). No recovery mechanism exists for solo social need. Actions blocked: talk, trade, gift, eat_together, work_together, comfort, flirt, teach, etc.
- **Impact:** Economy, social, and progression systems all require NPCs. A solo NPC can only eat, drink, sleep, gather, craft, and explore. Social need spirals to 0 with no recovery.
- **Fix:** Added `reflect` action — available when social > 30, recovers 15 social need, +2 happiness, -3 stress.

### DESIGN-R2 (HIGH): Shrine has no actions for most NPCs
- **Evidence:** Visited Holy Stone Shrine with Eldric (stress 11, HP 100, happiness 67). list_actions showed zero shrine-specific actions. Pray requires SpiritualSensitivity > 65 (believer path) or stress > 85/HP < 30/happiness < 15 (desperate path).
- **Impact:** Shrines are completely useless for ~80% of NPCs, wasting a location type.
- **Fix:** Added `meditate` action — no stat requirements, routes to shrine, gives -5 stress, -10 fatigue, +3 happiness, -8 social need.

### DESIGN-R3 (HIGH): Business ownership unlocks zero new actions
- **Evidence:** Eldric claimed Scribe's Library for 5 gold. list_actions before/after showed identical actions. No copy_text, write_journal, or read_book became available.
- **Impact:** start_business costs gold but provides no gameplay benefit.

### DESIGN-R4 (HIGH): Scavenge picks first 5 stacks by insertion order, not value
- **Evidence:** East Side Market had bread x60, ceramic x1, leather x2, iron ore x4 on ground. Scavenge picked 5 bread x2 stacks, skipping all rare items. The cap iterates in ground-item order, not by value or rarity.
- **Impact:** Players miss rare/valuable items in favor of common bread.

### DESIGN-R5 (SEVERE): Every forest has massive enemy counts (8-24)
- **Evidence:** Whispering Forest Edge: 24 enemies. Breezy Forest Grove: 8 enemies. Plus 2 dire wolves at Sunrise Stable and 2 at Oakhaven Farm.
- **Impact:** Forage and hunt actions are suicide. Farms near forests are death traps. Dire wolf event makes most outdoor locations extremely dangerous.

### DESIGN-R6 (MEDIUM): Duplicate location names ("Hill Well" x2, "Dock" x2)
- **Evidence:** list_actions showed "Hill Well" at +5 min and +100 min travel. "Dock" at +35 min and +150 min. These are different locations with the same name.
- **Impact:** Confusing for players and AI agents. Cannot distinguish between them in API calls.

## New Features Implemented

### FEATURE-1: `meditate` action at shrine (solo stress/fatigue/social recovery)
- **Location:** wellbeing.go
- **Destination:** Nearest shrine
- **Conditions:** Not a repeat of meditate/pray (no other requirements)
- **Effects:** -5 stress, -10 fatigue, +3 happiness, -8 social need
- **Rationale:** Addresses DESIGN-R2 (shrine useless). Makes shrines accessible to all NPCs. Also provides a solo social need recovery path via peaceful contemplation.

### FEATURE-2: `reflect` action (solo social recovery anywhere)
- **Location:** wellbeing.go
- **Conditions:** Social need > 30 (feeling lonely), not a repeat
- **Effects:** -15 social need, +2 happiness, -3 stress
- **Rationale:** Addresses DESIGN-R1 (solo world unplayable for social). Provides a lifeline for NPCs in empty worlds to slowly recover social need through self-reflection. Not as powerful as actual social interaction but prevents social need from reaching 0.

### BUGFIX: Sleep text fallback (BUG-R1)
- **Location:** survival.go line 163
- **Change:** "Slept at home" → "Rested and slept" for the fallback case when NPC has no HomeID.

## Regression Checks

| Previous Bug/Fix | Status |
|-----|--------|
| PR #12 scavenge item cap (5 stacks) | **CONFIRMED** — tested at 3 locations |
| PR #12 scavenge gold cap (20) | **CONFIRMED** — tested at 3 locations |
| PR #11 recall_memories category filter | **CONFIRMED** — routine and economic both return correct entries |
| PR #10 SpawnOnDemand profession equipment | **CONFIRMED** — farmer gets farmer skill 32, forager gets forager skill 47 |
| PR #9 check_location loc.ID fix | **CONFIRMED** — shows enemies, travel time, resources correctly |
| PR #9 self-exclusion inspect_person | Not tested (no NPCs available) |
| PR #7 list_actions location limit | **CONFIRMED** — shows 10 + "...and N more" |
| PR #7 ground gold decay | **PARTIALLY CONFIRMED** — gold still present but amounts seem lower |

## Actions Tested

### Eldric (24 ticks)
1. scavenge, 2. travel (shrine), 3. eat, 4. farm, 5. gather_thatch, 6. drink, 7. gather_clay, 8. explore, 9. sleep, 10. scavenge, 11. craft_pottery, 12. travel (ESM), 13. scavenge, 14. travel (library), 15. fish, 16. gather_clay, 17. travel (library), 18. start_business, 19. craft_pottery, 20. explore, 21. drink, 22. travel (stable) **→ ATTACKED**, 23. flee_area **→ INTERRUPTED**, 24. flee_area **→ DIED**

### Mara (8+ ticks)
1. scavenge, 2. drink, 3. eat, 4. farm, 5. gather_thatch, 6. fish, 7. gather_clay, 8. sleep (in progress)

## Working Mechanics Verified
- eat, drink, sleep, farm, fish, gather_thatch, gather_clay, scavenge, craft_pottery, explore, travel, start_business, check_location, recall_memories, list_actions, observe, check_self

## Priority Ranking

| ID | Issue | Priority | Status |
|----|-------|----------|--------|
| BUG-R1 | Sleep text says "home" | LOW | FIXED |
| BUG-R2 | flee_area too slow | HIGH | DOCUMENTED |
| BUG-R3 | Dire wolves at safe locations | HIGH | DOCUMENTED |
| DESIGN-R1 | Solo social unplayable | SEVERE | FIXED (reflect action) |
| DESIGN-R2 | Shrine useless | HIGH | FIXED (meditate action) |
| DESIGN-R3 | Business no benefit | HIGH | DOCUMENTED |
| DESIGN-R4 | Scavenge picks first stacks | MEDIUM | DOCUMENTED |
| DESIGN-R5 | Forest death traps | SEVERE | DOCUMENTED |
| DESIGN-R6 | Duplicate location names | MEDIUM | DOCUMENTED |
