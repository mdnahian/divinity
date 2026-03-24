# Playtest Notes — Juna Ironhand

**Session:** 2026-03-24
**NPC:** Juna Ironhand (Farmer at Juna Ironhand's Home)
**Personality:** virtuous, diligent, loyal, resilient, open-minded, introverted
**NPC ID:** 1e3cb962-f488-49b7-b41b-0cb04c043215

---

## Tick Log

### Tick 1 — Game Tick 182, Day 1, 22:50 (night)
**Location:** Juna Ironhand's Home
**Stats:** HP 100 | Hunger 76% | Thirst 68% | Fatigue 23% | Social 68% | Happiness 61% | Stress 30% | Hygiene 87% | Sobriety 100%
**Inventory:** gold x13, wheat x3
**Weather:** Clear
**NPCs nearby:** None
**Actions available:** eat, drink, forage, hunt, farm, fish, gather_thatch, gather_clay, mill_grain, explore, travel
**Action taken:** `eat` — Test the new solo eat action (was BUG-2 in previous playtest)
**Notes:**
- **REGRESSION CHECK: BUG-2 (No solo eat) → FIXED!** The `eat` action is now available: "Eat food from inventory [~15 min]"
- `gather_clay` shows "+5 min travel" but lists NO specific locations — different from previous bug (BUG-1). Need to investigate further.
- Employed by Tomlin Gravel at 2 gold/day. Farmer skill at 54.
- Closest farm: The Gritroot Farm (+10 min travel). Will farm there next.

### Tick 2 — Game Tick 187, Day 1, 23:15 (night)
**Location:** Juna Ironhand's Home
**Stats:** HP 100 | Hunger 84% (+8) | Thirst 67% (-1) | Fatigue 24% (+1) | Social 65% (-3) | Happiness 61% | Stress 30% | Hygiene 87% | Sobriety 100%
**Inventory:** gold x13, wheat x2 (-1)
**Weather:** Clear
**NPCs nearby:** None
**Memory:** "Ate wheat (+8 hunger)" — eat action confirmed working!
**Actions available:** eat, drink, forage, hunt, farm, fish, gather_thatch, gather_clay, mill_grain, explore, travel
**Action taken:** `farm` at The Gritroot Farm (+10 min travel) — Work the wheat fields
**Notes:**
- Eat action result: +8 hunger, consumed 1 wheat. Action works correctly.
- `explore` travel time changed from "+10 min" (Tick 1) to "+55 min" (Tick 2) — seems dynamic/random. Related to BUG-3?
- `gather_clay` still shows "+5 min travel" with no locations. Consistent oddity.

### Tick 3 — Game Tick 193, Day 1, 23:45 (night)
**Location:** The Gritroot Farm
**Stats:** HP 100 | Hunger 83% (-1) | Thirst 67% (0) | Fatigue 25% (+1) | Social 62% (-3) | Happiness 61% | Stress 30% | Hygiene 87%
**Inventory:** gold x13, wheat x2
**Weather:** Clear
**NPCs nearby:** None
**Busy:** YES — farming, busy_until_tick 195 (2 ticks remaining)
**Farm resources:** wheat: 30, thatch: 20. Owned by Callum Thornash.
**Notes:**
- Successfully traveled to The Gritroot Farm. Farm correctly shows "(here)" for farm and gather_thatch actions.
- New action appeared: `go_home` — Head home [~15 min + 10 min travel]. Nice context-dependent action.
- `gather_clay` STILL shows "+5 min travel" with no locations listed. This is consistent across home and farm.

### Tick 4 — Game Tick 199, Day 2, 00:15 (night)
**Location:** The Gritroot Farm
**Stats:** HP 100 | Hunger 83% (0) | Thirst 66% (-1) | Fatigue 37% (+12!) | Social 59% (-3) | Happiness 61% | Stress 25% (-5) | Hygiene 84% (-3)
**Inventory:** gold x15 (+2 wage), wheat x6 (+4 harvested)
**Weather:** Cloudy (changed from clear)
**NPCs nearby:** None
**Events:** Wolves spotted at The Wandering Grove and The Shivering Pines. Strange blood-moon omens.
**Memory:** "Worked the fields and harvested 4 wheat"
**New actions:** `repair_building` — Repair a building you own [~45 min] (no locations listed)
**Action taken:** `gather_clay` — Test BUG-1 from previous playtest
**Notes:**
- Farming results: +4 wheat, +2 gold (wage), +12 fatigue, -5 stress. Good farming loop.
- `repair_building` appeared but has no locations. Does NPC own a building? check_self doesn't mention one. Possibly a bug or unlisted condition.
- `gather_clay` still shows "+5 min travel" with no specific locations. Submitting to see what happens.

### Tick 5 — Game Tick 204, Day 2, 00:40 (night)
**Location:** The Boneridge Well (traveled from Gritroot Farm)
**Stats:** HP 100 | Hunger 82% (-1) | Thirst 65% (-1) | Fatigue 38% (+1) | Social 56% (-3) | Happiness 61% | Stress 25%
**Inventory:** gold x15, wheat x6
**Weather:** Cloudy
**NPCs nearby:** Corden Flint (carpenter, neutral mood, relation: 0) — FIRST NPC ENCOUNTER!
**Busy:** YES — gather_clay, busy_until_tick 206 (2 ticks remaining)
**Resources here:** clay: 8, water: 150
**Notes:**
- `gather_clay` traveled to The Boneridge Well (+5 min from farm). The "+5 min travel" in list_actions was correct!
- BUG-1 REGRESSION CHECK: The behavior is different now. gather_clay doesn't list specific locations (unlike drink/farm/fish), but it does find and travel to the nearest clay source automatically. The "+5 min travel" seems to reflect actual travel to nearest clay. This is a DIFFERENT behavior than what Ulric saw — possibly improved but UX is confusing since no location name is shown.
- Will try to interact with Corden Flint next tick if still here (test talk/social actions).

### Tick 6 — Game Tick 210, Day 2, 01:10 (night) — 6-TICK CHECKPOINT
**Location:** The Boneridge Well
**Stats:** HP 100 | Hunger 82% (-1) | Thirst 65% (0) | Fatigue 49% (+11!) | Social 54% (-2) | Happiness 61% | Stress 25% | Hygiene 84%
**Inventory:** gold x15, wheat x6, clay x1 (+1 new!)
**Weather:** Cloudy
**NPCs nearby:** Corden Flint (carpenter, neutral, relation: 0)
**Memory:** "Gathered 1 clay near the water"
**New actions:** `talk`, `teach` (2 gold fee, ~30 min) — both require nearby NPC
**Action taken:** `talk` to Corden Flint — Test social interaction
**Notes:**
- gather_clay results: +1 clay, +11 fatigue. Action works correctly at well with clay resources.
- **BUG-1 REGRESSION: IMPROVED.** gather_clay now shows "(here)" when at a clay location (The Boneridge Well with clay: 8). No more incorrect travel time. However, from non-clay locations it shows "+5 min travel" with no location names — UX could still be improved by showing the destination name.
- `drink` correctly shows The Boneridge Well "(here)" — location awareness working well.
- `teach` is new — earn 2 gold by teaching skills. Interesting economy action.
- Fatigue at 49% — getting high. Will need to address soon.

### Tick 7 — Game Tick 216, Day 2, 01:40 (night)
**Location:** The Boneridge Well
**Stats:** HP 100 | Hunger 82% (0) | Thirst 64% (-1) | Fatigue 49% (0) | Social 70% (+16!) | Happiness 64% (+3) | Stress 25% | Hygiene 84%
**Inventory:** gold x15, wheat x6, clay x1
**Weather:** Cloudy
**NPCs nearby:** Corden Flint (carpenter, relation: 3 now)
**Memory:** "Chatted with Corden Flint (social +20, relationship +3)"
**Relationship:** Corden Flint — "slight positive (talk)"
**Action taken:** `drink` at The Boneridge Well (here) — Restore thirst
**Notes:**
- Talk results: +20 social (raw), +3 relationship, +3 happiness. Great social boost.
- **DESIGN-1 REGRESSION: FIXED!** Memory now shows "social +20" (positive) instead of "social -20". The confusing negative sign is gone.
- `talk` is still available immediately after talking — **DESIGN-2 REGRESSION CHECK:** No cooldown observed yet! talk still shows in actions list. Will check after drink completes.
- Wolf events no longer showing in observe. Events cleared?

### Tick 8 — SESSION LOST
**Token returned "unknown token" error.** NPC session expired or server restarted between ticks (5 min gap).
**BUG-J3: NPC session token expired unexpectedly (HIGH)** — After ~35 minutes of active play, the token became invalid. No error or warning before this happened. This is a critical issue for agent-based gameplay.

---
## Continuing as Sable Brightwater (blacksmith) from Tick 8 onwards

**NPC:** Sable Brightwater (Blacksmith at The Sandlord's Manor)
**Personality:** charming, outgoing, clever, generous
**NPC ID:** 6885f66e-56da-421c-9346-ab908230a63b

### Tick 8 — Game Tick 4, Day 2, 00:40 (night)
**Location:** The Sandlord's Manor (manor)
**Stats:** HP 100 | Hunger 65% | Thirst 57% | Fatigue 25% | Social 62% | Happiness 56% | Stress 10% | Hygiene 67% | Sobriety 100%
**Inventory:** gold x11, wheat x3, iron ore x4, iron ingot x2
**Weather:** Cloudy
**NPCs nearby:** Rowan Hillcrest (scribe, neutral), Cora Coldstream (herbalist, calm), Willa Ashford (count, lonely!)
**Events:** Wolves at Wandering Grove and Shivering Pines. Strange omens (duplicate event text!).
**Actions available:** eat, drink, forage, hunt, farm, fish, gather_thatch, gather_clay, mill_grain, talk, eat_together, steal(!), explore, travel, start_business
**Action taken:** `eat_together` with Willa Ashford — Share meal with lonely NPC
**Notes:**
- **NEW BUG-J4: Duplicate world event text.** "Strange omens in the sky" appears TWICE in observe output with identical text. Likely a display/dedup bug.
- `steal` action available — RISKY tag. Haven't seen this before. Will test later.
- `start_business` available — Can claim The Sandlord's Manor? Or nearby unowned location?
- check_self shows "Satiation" and "Hydration" instead of "Hunger" and "Thirst" (Juna showed Hunger/Thirst). Naming inconsistency? Or NPC-dependent?
- Blacksmith has iron ore and iron ingot but NO smelt/forge action visible. **DESIGN-J3: Blacksmith has no profession-specific actions.** A blacksmith should be able to smelt ore or forge items.
- Thirst at 57% — concerning. Will drink next tick.
- Employed by Brennan Ashford (same surname as Willa Ashford — related?)

