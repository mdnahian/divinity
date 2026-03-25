# Playtest Notes — Vala Greenleaf

**Session:** 2026-03-24
**NPC:** Vala Greenleaf (Carpenter)
**Starting Location:** Soggy Keel
**Personality:** curious, outgoing

---

## Tick Log

### Tick 1 — Game Tick 1097, Day 5 01:25 (night, rain)
**Location:** Soggy Keel (dock)
**Stats:** HP 100 | Hunger 71% | Thirst 66% | Fatigue 32% | Social 31% | Happiness 54% | Stress 18% | Hygiene 57%
**Inventory:** gold x13, iron ore x1, logs x2, rope x2
**Equipped:** iron axe

**People here:** Ulric Deeproot (healer, lonely), Osric Meadowlark (potter, lonely), Rowan Underhill (farmer, neutral)

**Actions available (local):** fish (here), gather_clay (here), talk, explore, unequip_item
**Actions available (nearby):** drink (Fen Well +5min), buy_food (Fog Market +5min), buy_supplies (Fog Market +5min)

**Decision:** Talk to Osric Meadowlark — Social is critically low at 31%, and Osric is nearby and lonely. Tried Ulric first but he left ("no longer here" error).

**Observations:**
- BUG-J4 STILL PRESENT: Duplicate "Strange omens in the sky" events in observe output (two different texts)
- BUG-J1 POSSIBLY FIXED: gather_clay now shows full location names from non-clay locations (previously showed "+5 min travel" with no names)
- Interesting: NPC target can leave between observe and commit — got "Ulric Deeproot is no longer here" error. Not a bug per se, but worth noting for agent design.
- Carpenter has no profession-specific actions (no craft/build). Similar to DESIGN-J1 (blacksmith).

### Tick 2 — Game Tick 1101, Day 5 01:45 (night, rain)
**Location:** Soggy Keel (dock)
**Stats:** HP 100 | Hunger 71% | Thirst 66% | Fatigue 33% | Social 49% | Happiness 57% | Stress 18% | Hygiene 57%
**Inventory:** gold x13, iron ore x1, logs x2, rope x2 (unchanged)
**Memory:** "Chatted with Osric Meadowlark (social +20, relationship +3)"
**Relationship:** Osric Meadowlark: slight positive (talk) [+3]

**People here:** Ulric Deeproot (healer, lonely), Osric Meadowlark (potter, calm, +3), Rowan Underhill (farmer, neutral)

**Actions available (local):** fish (here), gather_clay (here), talk, explore (+5 min travel), unequip_item
**Actions available (nearby):** drink (Fen Well +5min), buy_food (Fog Market +5min), buy_supplies (Fog Market +5min)

**Decision:** Fish at Soggy Keel — available here with no travel, haven't tried it yet. Testing action diversity.

**Observations:**
- Talk results confirmed: social +20, relationship +3, happiness +3 (54→57). Good social action.
- No duplicate world events in observe this time (events may have rotated/expired). Can't confirm BUG-J4 regression.
- Explore travel time changed from +10 min (tick 1) to +5 min (tick 2) — worth monitoring.
- Osric's mood changed from "lonely" to "calm" after our talk — NPC mood reacts to social interaction (working correctly).
- Stat changes between ticks: hunger -0.28, thirst -0.46, fatigue +0.4, social +18 (from talk action).

**Actions tried so far:** talk (1), fish (2)

### Tick 3 — Game Tick 1106, Day 5 (night, rain)
**Location:** Soggy Keel (dock)
**Stats:** HP 100 | Hunger 71% | Thirst 65% | Fatigue 33% | Social ~49%
**Status:** BUSY — fishing, completes at tick 1107 (1 tick remaining)
**Skipped:** NPC busy, no action taken.

### Tick 4 — Game Tick 1111, Day 5 02:35 (night, clear)
**Location:** Soggy Keel (dock)
**Stats:** HP 100 | Hunger 70% | Thirst 65% | Fatigue 42% | Social 44% | Happiness 56% | Stress 17% | Hygiene 57%
**Inventory:** gold x13, iron ore x1, logs x2, rope x2 (unchanged — fishing caught nothing)
**Skills:** carpenter: 58, fisher: 0 (NEW — gained from fishing)
**Memory:** "Sat by the water but caught nothing (+fishing experience)"

**People here:** Ulric Deeproot (healer, lonely), Osric Meadowlark (potter, lonely, +3), Rowan Underhill (farmer, neutral)
**Weather:** Changed from rain to clear

**NEW actions this tick:** rest (Drowned Lantern Inn +30 min)
**Actions available (local):** fish (here), gather_clay (here), talk, explore (+25 min travel), unequip_item
**Actions available (nearby):** drink (Fen Well +5min), buy_food (Fog Market +5min), rest (Drowned Lantern Inn +30min)

**Decision:** Gather clay at Soggy Keel — available locally (clay: 10), testing resource gathering at current location. Verifying BUG-J1 fix.

**Observations:**
- Fishing caught nothing despite fish: 20 at location. Gained fisher: 0 skill. Fatigue +9 (33→42). Seems like fishing success might be skill-dependent.
- Explore travel time keeps fluctuating: +10 (tick 1) → +5 (tick 2) → +25 (tick 4). Possible BUG or randomized per tick.
- "rest" action appeared this tick — wasn't available in ticks 1-2. May be fatigue-gated (appeared when fatigue hit 42%).
- Osric's mood reverted from "calm" back to "lonely" despite our conversation 2 ticks ago.

**Actions tried so far:** talk (1), fish (2), gather_clay (4)

### Tick 5 — Game Tick 1116, Day 5 (night)
**Location:** Soggy Keel (dock)
**Stats:** HP 100 | Hunger 70% | Thirst 64% | Fatigue 42% | Social ~44%
**Status:** BUSY — gathering clay, completes at tick 1117 (1 tick remaining)
**Skipped:** NPC busy, no action taken.

### Tick 6 — Game Tick 1121, Day 5 03:25 (night, clear)
**Location:** Soggy Keel (dock)
**Stats:** HP 100 | Hunger 70% | Thirst 63% | Fatigue 53% | Social 39% | Happiness 54% | Stress 17% | Hygiene 57%
**Inventory:** gold x13, iron ore x1, logs x2, rope x2, clay x2 (NEW +2 clay from gather_clay)
**Memory:** "Gathered 2 clay near the water"

**People here:** Ulric Deeproot (healer, lonely), Osric Meadowlark (potter, lonely, +3), Rowan Underhill (farmer, neutral)
**Resources:** fish: 20, clay: 9 (was 10, decreased by 1 — but we got 2 clay? Resource tracking mismatch?)

**NEW actions this tick:**
- **sleep** (~240 min + 30 min travel) — appeared when fatigue hit 53%
- **craft_pottery** — appeared after getting clay x2 in inventory! No location listed.

**Actions available:** drink, sleep, forage, hunt, farm, fish, chop_wood, gather_thatch, gather_clay, craft_pottery, buy_food, buy_supplies, talk, rest, unequip_item, explore, travel

**Decision:** Craft pottery — rare crafting action just unlocked by having clay. Testing crafting mechanics.

**Observations:**
- gather_clay gave +2 clay, fatigue +11 (42→53). Clay at location decreased 10→9 (only -1, but we got 2 — mismatch?). POSSIBLE BUG.
- craft_pottery unlocked by inventory (clay x2) — good conditional action gating.
- sleep appeared — likely fatigue threshold (>50%?). Duration 240 min (~16 ticks!) — very long.
- New world event: divine wrath about neglect/dead/granaries. Only 1 event this time (no duplicates).
- Social dropped 44→39 over 2 ticks (~2.5/tick passive drain).

**Actions tried so far:** talk (1), fish (2), gather_clay (4), craft_pottery (6)

### Tick 7 — Game Tick 1126, Day 5 (night)
**Location:** Soggy Keel (dock)
**Stats:** HP 100 | Hunger 69% | Thirst 63% | Fatigue 53% | Social ~34%
**Status:** BUSY — crafting pottery, completes at tick 1134 (8 ticks remaining!)
**Skipped:** NPC busy, no action taken.
**Note:** craft_pottery takes ~8 ticks (45 game min = 8 ticks at ~5.6 min/tick). This will keep us busy for several loop iterations.

### Tick 8 — Game Tick 1131, Day 5
**Location:** Soggy Keel (dock)
**Stats:** HP 100 | Hunger 69% | Thirst 62% | Fatigue 54% | Social ~29%
**Status:** BUSY — crafting pottery, completes at tick 1134 (3 ticks remaining)
**Skipped:** NPC busy, no action taken.
**Note:** Fatigue barely changing during crafting (+0.5 over 5 ticks). Thirst/hunger declining steadily. Social now critically low ~29%.

### Tick 9 — Game Tick 1136, Day 5 04:40 (night, clear)
**Location:** Soggy Keel (dock)
**Stats:** HP 100 | Hunger 69% | Thirst 62% | Fatigue 62% (CRITICAL) | Social 32% | Happiness 53% | Stress 13% | Hygiene 57%
**Inventory:** gold x13, iron ore x1, logs x2, rope x2, ceramic x1 (NEW — clay x2 → ceramic x1)
**Memory:** "Shaped clay into a ceramic piece"

**People here:** Ulric Deeproot (healer, lonely), Osric Meadowlark (potter, lonely, +3), Rowan Underhill (farmer, lonely)
**Resources:** fish: 20, clay: 10 (regenerated from 9!)

**Actions available:** drink, sleep, forage, hunt, farm, fish, chop_wood, gather_thatch, gather_clay, buy_food, buy_supplies, talk, rest, unequip_item, explore (+30 min), travel
**Gone:** craft_pottery (no more clay in inventory)

**Decision:** Rest at Drowned Lantern Inn (+30 min travel) — fatigue is critical at 62%, testing rest mechanics. Haven't tried rest yet.

**Observations:**
- craft_pottery result: clay x2 consumed → ceramic x1 produced. Fatigue +9 (53→62). Stress -4 (17→13). Duration ~8 ticks.
- Clay resource at location regenerated: 9 → 10. Resources replenish over time.
- craft_pottery disappears when no clay in inventory — correct conditional gating.
- Rowan Underhill mood changed: neutral → lonely. All 3 NPCs now lonely.
- Memory system shows "Defining moments" category — conversation with Osric marked as "(vivid)".
- Explore travel time: +30 min (previously +25, +5, +10). Still fluctuating.

**Actions tried so far:** talk (1), fish (2), gather_clay (4), craft_pottery (6), rest (9)

### Tick 10 — Game Tick 1147, Day 5
**Location:** Soggy Keel (dock) — still here, traveling to Drowned Lantern Inn
**Stats:** HP 100 | Hunger 68% | Thirst 60% | Fatigue 65% (UP from 62%!) | Social ~27%
**Status:** BUSY — resting, completes at tick 1155 (8 ticks remaining)
**Skipped:** NPC busy, no action taken.
**IMPORTANT:** Fatigue INCREASED 62%→65% while traveling to rest. Expected during travel, but need to verify fatigue drops after rest completes. Related to known bug: "fatigue rises during sleep."

### Tick 11 — Game Tick 1151, Day 5
**Location:** Soggy Keel (still traveling)
**Stats:** HP 100 | Hunger 67% | Thirst 60% | Fatigue 66% (still rising!) | Social ~24%
**Status:** BUSY — resting, completes at tick 1155 (4 ticks remaining)
**Skipped:** NPC busy, no action taken.
**Note:** Fatigue tracking during rest: 62% (tick 9) → 65% (tick 10) → 66% (tick 11). Continuously rising. If this doesn't drop after rest completes, it's a confirmed bug.

### Tick 12 — Game Tick 1156, Day 5 06:20 (morning, clear)
**Location:** Drowned Lantern Inn (inn) — NEW LOCATION! "A warm hearth crackles inside a house built atop a submerged tower."
**Stats:** HP 100 | Hunger 67% | Thirst 59% | Fatigue 46% (DOWN from 66%!) | Social 22% (CRITICAL) | Happiness 50% | Stress 2% | Hygiene 57%
**Inventory:** gold x12 (was 13, -1 for rest), iron ore x1, logs x2, rope x2, ceramic x1
**Mood:** Changed to "lonely" (was "calm")
**Memory:** "Paid 1 gold and sat by the hearth at the inn (-8 stress, -20 fatigue)"

**People here:** Ulric Deeproot (herbalist — was "healer" at Soggy Keel!), Hilda Greenleaf (scribe, miserable), Owen Baker (baker, miserable)
**Owned by:** Molly Sprig

**Actions available:** drink, forage, hunt, farm, fish, chop_wood, gather_thatch, gather_clay, buy_food, buy_supplies, talk, rest (here!), unequip_item, explore (+20 min), travel
**Gone:** sleep (was available at Soggy Keel, not here — wait, this IS an inn?)

**Decision:** Talk to Hilda Greenleaf — Social is critical at 22%, and Hilda shares our surname. Testing social interaction at new location.

**Observations:**
- REST WORKS: fatigue dropped 66%→46% (-20). Stress dropped 17%→2% (-8 from rest + passive). Cost: 1 gold. Duration: ~19 ticks total (travel+rest).
- Fatigue rose during travel (62→66, +4 over 3 ticks), then dropped -20 on rest completion. Net: -16 effective.
- Ulric Deeproot's profession changed from "healer" to "herbalist" — POSSIBLE BUG or display inconsistency.
- sleep action DISAPPEARED despite being at an inn. Was available at Soggy Keel (dock) but not at Drowned Lantern Inn (inn). POSSIBLE BUG.
- Everyone at this inn is "miserable" — harsh location?
- Hilda Greenleaf shares our surname — potential family system?
- It's now morning (06:20) — time of day progressing.

**Actions tried so far:** talk (1, 12), fish (2), gather_clay (4), craft_pottery (6), rest (9)

### Tick 13 — Game Tick 1161, Day 5
**Location:** Drowned Lantern Inn
**Stats:** HP 100 | Hunger 67% | Thirst 59% | Fatigue 47% | Social ~22%
**Status:** BUSY — talking, completes at tick 1162 (1 tick remaining)
**Skipped:** NPC busy, no action taken.

### Tick 14 — Game Tick 1166, Day 5 07:10 (morning, rain)
**Location:** Drowned Lantern Inn
**Stats:** HP 100 | Hunger 66% | Thirst 58% | Fatigue 47% | Social 37% | Happiness 52% | Stress 1% | Hygiene 57%
**Inventory:** gold x12, iron ore x1, logs x2, rope x2, ceramic x1 (unchanged)
**Memory:** "Chatted with Hilda Greenleaf (social +20, relationship +3)"
**Relationships:** Osric Meadowlark: slight positive (+3), Hilda Greenleaf: slight positive (+3)

**People here:** Ulric Deeproot (herbalist, miserable), Hilda Greenleaf (scribe, miserable, +3), Owen Baker (baker, miserable)

**Actions available:** drink, forage, hunt, farm, fish, chop_wood, gather_thatch, gather_clay, buy_food, buy_supplies, talk, rest (here), unequip_item, explore (+30 min), travel
**Still missing:** sleep (at an inn!)

**Decision:** Explore — haven't tested exploration yet. Will discover new location and test travel time mechanics.

**Observations:**
- Talk with Hilda: social +20, relationship +3, happiness +2 (50→52). Consistent with Osric results.
- Hilda remained "miserable" despite the conversation — our mood went calm→lonely→calm but hers didn't change. NPCs at this location seem stuck in miserable state.
- Weather changed: clear → rain
- Still no sleep action at inn — BUG-V4 confirmed across multiple ticks.

**Actions tried so far:** talk (1, 12), fish (2), gather_clay (4), craft_pottery (6), rest (9), explore (14)

### Tick 15 — Game Tick 1171, Day 5
**Location:** Drowned Lantern Inn (still here, exploring outward)
**Stats:** HP 100 | Hunger 66% | Thirst 58% | Fatigue 49% | Social ~32%
**Status:** BUSY — exploring, completes at tick 1175 (4 ticks remaining)
**Skipped:** NPC busy, no action taken.
**Note:** Explore from inn with +30 min travel takes ~9 ticks total. Thirst dropping toward 57%.

### Tick 16 — Game Tick 1176, Day 5 08:00 (morning, rain)
**Location:** Echo Well (well) — NEW! "The sound of the water falling echoes strangely in the narrow shaft."
**Stats:** HP 100 | Hunger 66% | Thirst 57% | Fatigue 54% | Social 32% | Happiness 51% | Stress 0%
**Inventory:** gold x12, iron ore x1, logs x2, rope x2, ceramic x1 (unchanged)
**Memory:** "Wandered to Echo Well, taking in the sights"

**Resources here:** water: 150, clay: 8
**People here:** Dorin Deeproot (farmer, lonely), Hilda Root (herbalist, lonely), Thomas Baker (baker, lonely), Bessa Oakshield (farmer, calm)

**Actions available:** drink (HERE!), sleep (+15 min), forage, hunt, farm, fish, chop_wood, gather_thatch, gather_clay (here), buy_food, buy_supplies, talk, rest, unequip_item, explore (+5 min), travel

**Decision:** Drink at Echo Well — thirst at 57%, we're at a well. Testing drink mechanics.

**Observations:**
- Explore moved us from Drowned Lantern Inn → Echo Well. Took ~10 ticks total (listed as 15+30=45 min).
- BUG-V4 CONFIRMED HARDER: sleep available HERE at Echo Well (+15 min to inn) but was MISSING when actually AT Drowned Lantern Inn (an inn). Sleep candidates function seems to only list inns you need to TRAVEL to, not the one you're at.
- drink shows "Echo Well" (here) correctly — good location awareness.
- "The Echo Well" also appears as a SEPARATE location (+140 min travel). Duplicate location names? Or different places?
- New NPCs here: 4 people, mostly lonely. Different from Soggy Keel crowd.
- Stress dropped to 0% (was 1%).

**Actions tried so far:** talk (1, 12), fish (2), gather_clay (4), craft_pottery (6), rest (9), explore (14), drink (16)

### Tick 17 — Game Tick 1187, Day 5
**Location:** Echo Well
**Stats:** HP 100 | Hunger 65% | Thirst 56% | Fatigue 55% | Social ~27%
**Status:** BUSY — drinking, completes at tick 1190 (3 ticks remaining)
**Skipped:** NPC busy, no action taken.
**POSSIBLE BUG:** Drink started ~tick 1176, ends tick 1190 = ~14 ticks for a "~15 min" action at "(here)" location. At ~5 min/tick, should be ~3 ticks. Duration seems 4-5x too long. POSSIBLE BUG — drink duration mismatch.

### Tick 18 — Game Tick 1191, Day 5 09:15 (morning, rain)
**Location:** Echo Well
**Stats:** HP 100 | Hunger 65% | Thirst 100% (FULL!) | Fatigue 56% | Social 24% (CRITICAL) | Happiness 49% | Stress 0%
**Inventory:** gold x12, iron ore x1, logs x2, rope x2, ceramic x1 (unchanged)
**Memory:** "Drank fresh water at the well (thirst fully restored)"
**Mood:** lonely

**Resources:** water: 149 (was 150, -1 from drinking)
**People here:** Dorin Deeproot (farmer, lonely), Hilda Root (herbalist, lonely), Thomas Baker (baker, lonely), Bessa Oakshield (farmer, calm)

**Decision:** Talk to Thomas Baker — Social critical at 24%. Tried Bessa first but she left ("no longer here"). Second NPC to leave mid-tick.

**Observations:**
- DRINK WORKS: thirst fully restored 56%→100%. Free, uses well water. Resource decreased 150→149.
- BUG-V5 CONFIRMED: Drink duration ~15 game ticks (~75 game-min) for a listed "~15 min" action at "(here)". Memory shows 07:55→09:10 = 75 game-min. 5x duration mismatch.
- NPCs leave between observe and commit (Bessa Oakshield "no longer here"). Second time this happened (Ulric at tick 1). Race condition with NPC movement.
- Explore travel time now +45 min (was +30, +25, +5, +10). Still fluctuating every tick.

**Actions tried so far:** talk (1, 12, 18), fish (2), gather_clay (4), craft_pottery (6), rest (9), explore (14), drink (16)

### Tick 19 — Game Tick 1199, Day 5
**Location:** Echo Well
**Stats:** HP 100 | Hunger 64% | Thirst 99% | Fatigue 56% | Social ~24%
**Status:** BUSY — talking to Thomas Baker, completes at tick 1200 (1 tick remaining)
**Skipped:** NPC busy, no action taken.
**Note:** Talk started ~tick 1191, ends tick 1200 = ~9 ticks for a "~15 min" action. Further confirms BUG-V5 duration mismatch (~45 game-min vs listed 15 min, 3x longer).

### Tick 20 — Game Tick 1201, Day 5 10:00 (morning)
**Location:** Echo Well
**Stats:** HP 100 | Hunger 64% | Thirst 99% | Fatigue 57% | Social 39% | Happiness 51% | Stress 0%
**Inventory:** gold x12, iron ore x1, logs x2, rope x2, ceramic x1 (unchanged)
**Memory:** "Chatted with Thomas Baker (social +20, relationship +3)"
**Relationships:** Osric (+3), Hilda Greenleaf (+3), Thomas Baker (+3)

**Decision:** Buy food at Fog Market (+40 min travel) — testing market/economy system, hunger declining to 64%.

**Observations:**
- Talk with Thomas: social +20, relationship +3. Duration: 09:10→10:00 = 50 game-min for "~15 min" action (3.3x mismatch). Confirms BUG-V5 pattern.
- Explore travel time now +60 min (was +45 last tick). Still fluctuating (DESIGN-V2).
- Token still working at tick 20 (~100 min). BUG-J3 token expiry appears FIXED.

**Actions tried so far:** talk (1, 12, 18), fish (2), gather_clay (4), craft_pottery (6), rest (9), explore (14), drink (16), buy_food (20)

### Tick 21 — Game Tick 1225, Day 5
**Location:** Echo Well (traveling to Fog Market)
**Stats:** HP 100 | Hunger 62% | Thirst 96% | Fatigue 61% (CRITICAL) | Social ~34%
**Status:** BUSY — buying food, completes at tick 1236 (11 ticks remaining)
**Skipped:** NPC busy, no action taken.
**Note:** buy_food started ~tick 1201, ends tick 1236 = 35 ticks for "~15 min + 40 min travel" (55 min listed). At ~5 min/tick that should be ~11 ticks, but it's 35. Another BUG-V5 data point: ~3.2x longer than expected.

### Tick 22 — Game Tick 1226, Day 5
**Location:** Echo Well (traveling to Fog Market)
**Stats:** HP 100 | Hunger 62% | Thirst 96% | Fatigue 61% | Social ~29%
**Status:** BUSY — buying food, completes at tick 1236 (10 ticks remaining)
**Skipped:** NPC busy, no action taken.

### Tick 23 — Game Tick 1231, Day 5
**Location:** Echo Well (traveling to Fog Market)
**Stats:** HP 100 | Hunger 62% | Thirst 95% | Fatigue 62% | Social ~24%
**Status:** BUSY — buying food, completes at tick 1236 (5 ticks remaining)
**Skipped:** NPC busy, no action taken.

### Tick 24 (FINAL) — Game Tick 1236, Day 5 13:00 (afternoon)
**Location:** Fog Market — NEW LOCATION (4th visited!)
**Stats:** HP 100 | Hunger 62% | Thirst 95% | Fatigue 62% | Social ~19%
**Inventory:** gold x12, iron ore x1, logs x2, rope x2, ceramic x1 (UNCHANGED — no food bought!)
**Memory:** "Nobody at the market had food to sell."

**Observations:**
- BUG-V6: buy_food FAILED — "Nobody at the market had food to sell" despite being listed as available action. Traveled ~35 ticks to Fog Market for nothing. No gold spent, no food gained.
- buy_food duration: 10:00→13:00 = 180 game-min for listed "~15 min + 40 min travel" (55 min). 3.3x mismatch. Confirms BUG-V5.
- Token still working at tick 24 (~120 min). BUG-J3 token expiry is FIXED.
- Visited 4 locations total: Soggy Keel → Drowned Lantern Inn → Echo Well → Fog Market.

**Final stats comparison (tick 1 → tick 24):**
| Stat | Start | End | Change |
|------|-------|-----|--------|
| HP | 100 | 100 | 0 |
| Hunger | 71% | 62% | -9% |
| Thirst | 66% | 95% | +29% (drank) |
| Fatigue | 32% | 62% | +30% |
| Social | 31% | ~19% | -12% (despite 4 talks) |
| Happiness | 54% | ~49% | -5% |
| Stress | 18% | 0% | -18% |

**Actions tested (9 unique):** talk (x4), fish, gather_clay, craft_pottery, rest, explore, drink, buy_food

---

## SESSION COMPLETE — 24 ticks finished. Proceeding to Phase 3-5.

