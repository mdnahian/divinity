# Playtest Notes — Merchant Elian

**Session:** 2026-03-26
**NPC:** Merchant Elian (Merchant at Golden Meadow Market)
**Personality:** empathetic, curious, brave, creative, unscrupulous, introverted
**NPC ID:** f609ba99-9a9b-4a06-889f-139d779d1d4d

---

## Tick Log

### Tick 1 — Game Tick 2565, Day 10, 03:45 (night)
**Location:** Golden Meadow Market (market)
**Stats:** HP 100 | Hunger 79% | Thirst 60% | Fatigue 28% | Social 51% | Happiness 59% | Stress 17% | Hygiene 65% | Sobriety 100%
**Inventory:** bread x2, gold x15
**Weather:** Storm (thunder and lightning)
**NPCs nearby:** Torin Silverbrook (carpenter, miserable, stressed, lonely), Gareth the Merchant (trader, neutral)
**Items on ground:** bread x1, gold x19, bread x1, gold x29 (48 gold and 2 bread on the ground!)
**Actions available:** eat, drink, forage, hunt, farm, fish, gather_thatch, gather_clay, mill_grain, scavenge, trade, buy_supplies, talk, comfort, steal, explore, travel, go_home, start_business
**Action taken:** `comfort` targeting Torin Silverbrook — Test the comfort mechanic on a stressed NPC
**Reasoning:** Torin is miserable, stressed and lonely. Testing new `comfort` action which hasn't been tested in any previous playtest.

**Key observations:**
- **REGRESSION CHECK: gather_clay now shows location names!** Lists all docks and wells with travel times. BUG-J1 appears FIXED.
- **New actions discovered:** `scavenge` (pick up items from ground), `buy_supplies` (market action), `comfort` (help stressed NPCs)
- `trade` shows "(here)" at Golden Meadow Market — confirms market-based action
- `buy_supplies` shows "(here)" at Golden Meadow Market — merchant-specific?
- **48 gold + 2 bread on the ground** — seems like a lot. Economic leak? NPCs dropping items?
- `check_self` shows "Satiation" and "Hydration" labels — matches Sable's playtest, not Juna's "Hunger"/"Thirst". The API state still uses hunger/thirst fields though.
- Merchant skill at 37. Employed? Not clear from check_self output.
- `start_business` available — will test later
- `go_home` shows +40 min travel — home is not at the market

### Tick 2 — Game Tick 2567, Day 10, 03:55 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 79% (-0) | Thirst 60% (-0) | Fatigue 28% (+0) | Busy
**Busy:** YES — comfort, busy_until_tick 2568 (1 tick remaining)
**Action:** Waiting for comfort action to complete
**Notes:**
- Stats have barely changed. Hunger dropped by ~0.14, thirst by ~0.23 per tick while busy. Minimal passive drain.

### Tick 3 — Game Tick 2568, Day 10, 04:00 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 79% | Thirst 60% | Fatigue 28% | Social 59% (+8 from comfort) | Happiness 59% | Stress 16% (-1) | Hygiene 65% | Sobriety 100%
**Inventory:** bread x2, gold x15
**Weather:** Cloudy
**NPCs nearby:** Torin Silverbrook (carpenter, miserable, relation: +10), Gareth the Merchant (trader, calm, relation: 0)
**Items on ground:** bread x1, gold x19, bread x1, gold x29 (still there)
**Memory:** "Comforted Torin Silverbrook (-15 stress, -3 trauma, relationship +10)"
**Relationship:** Torin Silverbrook: +10 (slight positive, comfort)
**Actions available:** eat, drink, forage, hunt, farm, fish, gather_thatch, gather_clay, scavenge, trade, buy_supplies, talk, comfort, steal, explore, travel, go_home, start_business
**Action taken:** `scavenge` — Pick up the gold and bread on the ground
**Reasoning:** 48 gold + 2 bread on the ground is a massive haul. Testing scavenge mechanic.

**Key observations:**
- Comfort action result: +8 social (NPC had -10 SocialNeed), -15 stress to target, -3 trauma to target, +10 relationship. 
- Comfort action requires Empathy >= 45 (code confirms: `n.Stats.Empathy < 45`). Our NPC has "empathetic" personality.
- Torin is STILL miserable despite being comforted. Stress likely still > 50 since comfort only reduces by 15.
- `comfort` still available — could keep comforting Torin. But testing scavenge first.
- `trade` shows "(here)" at Golden Meadow Market — we ARE at a market with another NPC. Will test trading after scavenge.
- `go_home` available with +40 min travel. We have a home somewhere.
- `start_business` available — can claim an unowned establishment. But needs 5 gold.

### Tick 4 — Game Tick 2569, Day 10, 04:05 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 79% | Thirst 59% | Fatigue 29% | Busy
**Busy:** YES — scavenge, busy_until_tick 2571 (2 ticks remaining)
**Notes:**
- Scavenge is taking 3 ticks. BaseGameMinutes = 0 (defaults to 15 min). Interesting that it's taking this long.
- Passive stat drain: Hunger -0.069/tick, Thirst -0.116/tick, Fatigue +0.1/tick. Very slow.

### Tick 5 — Game Tick 2570, Day 10, 04:10 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 79% | Thirst 59% | Fatigue 29% | Busy
**Busy:** YES — scavenge, busy_until_tick 2571 (1 tick remaining)

### Tick 6 — Game Tick 2573, Day 10, 04:25 (night) — 6-TICK CHECKPOINT
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 78% | Thirst 59% | Fatigue 29% | Social 57% | Happiness 58% | Stress 14% | Hygiene 65% | Sobriety 100%
**Inventory:** bread x4 (+2), gold x63 (+48!)
**Weather:** Cloudy
**NPCs nearby:** Torin Silverbrook (carpenter, miserable, +10), Gareth the Merchant (trader, calm, 0)
**Items on ground:** EMPTY (scavenged everything)
**Memory:** "Picked up loot from the ground: 1x bread, 19x gold, 1x bread, 29x gold."
**Actions available:** eat, drink, forage, hunt, farm, fish, gather_thatch, gather_clay, trade, buy_supplies, talk, comfort, steal, explore, travel, go_home, start_business
**Action taken:** `trade` with Gareth the Merchant — Test trading at the market
**Reasoning:** We're at a market with another trader. Testing core merchant mechanic.

**Key observations:**
- **Scavenge successful!** Picked up ALL items from ground: 2 bread + 48 gold. Scavenge picks up everything in one action.
- `scavenge` no longer shows in actions since ground is empty. Correct behavior!
- **DESIGN OBSERVATION:** 48 gold on the ground at a market seems like a lot. Where did it come from? Could be economic leak from NPCs dying or dropping items. 
- **Trade mechanic check:** trade Conditions in code check `n.LocationID != markets[0].ID` — only works at FIRST market. Since Golden Meadow Market might be the first market, this could work for us. BUG-E2 still relevant.
- Torin still miserable. Comfort seems to have reduced his stress but he's still in bad shape.
- Thirst dropping below 60% — will need to drink soon.

### Tick 7 — Game Tick 2574, Day 10, 04:30 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 78% | Thirst 59% | Fatigue 29% | Social 56% | Happiness 58% | Stress 14%
**Inventory:** bread x4, gold x63
**Busy:** YES — trade, busy_until_tick 2578 (4 ticks remaining)
**Notes:**
- Trade takes ~25 min base (5 ticks at ~5 min/tick). Merchant skill at 37 reduces this slightly.
- Passive drain continues: Hunger -0.069/tick, Thirst -0.116/tick, Fatigue +0.1/tick

### Ticks 8-9 — Busy (trade)
**Busy with trade until tick 2578.**

### Tick 10 — Game Tick 2580, Day 10, 05:00 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 78% | Thirst 58% | Fatigue 30% | Social 53% | Happiness 58% | Stress 12% | Hygiene 65% | Sobriety 100%
**Inventory:** bread x3 (-1), gold x67 (+4)
**Weather:** Cloudy
**NPCs nearby:** Torin Silverbrook (carpenter, miserable, +10), Gareth the Merchant (trader, calm, +5)
**Memory:** "Sold 1 bread to Gareth the Merchant for 4 gold (market price)."
**Action taken:** `talk` to Torin Silverbrook with dialogue "How are you feeling now?"
**Reasoning:** Testing talk mechanic with custom dialogue. Also building relationship.

**Key observations:**
- **Trade successful!** Sold 1 bread for 4 gold. Merchant skill gained 0.3. Gareth relationship +5.
- `trade` no longer shows in actions — because we need sellable non-gold items AND another NPC must afford the price. We have bread (worth ~4g) and Gareth might be running low.
- **Note:** Trade Conditions check `n.LocationID != markets[0].ID` but we're at Golden Meadow Market which IS the first market. So BUG-M2 doesn't block us here, but would block at East Side Market.
- Thirst now at 58% — dropping. Will need to drink at a well soon.
- Social at 53% — talk should help.

### Tick 11 — Talk completed
**Stats after talk:** Social 68% (+15), Relationship Torin +13 (+3)
**Memory:** "Chatted with Torin Silverbrook (social +20, relationship +3). Said: ..."
**Note:** talk correctly used our custom dialogue in the memory.

### Tick 12 — Game Tick 2591, Day 10, 05:55 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 77% | Thirst 57% | Fatigue 31% | Social 68% | Happiness 58% | Stress 12%
**Inventory:** bread x3, gold x67
**Action taken:** `eat` — Eat bread to restore hunger
**Reasoning:** Testing eat action. Will drink water next.

**Key observations:**
- Talk mechanic works well. +20 social (displayed as -20 SocialNeed), +3 relationship, +3 happiness.
- Custom dialogue was stored in memory correctly.
- Torin's relationship progressed from "neutral" to "acquaintance" label at +13.
- Torin STILL miserable. His stress must be very high.

### Tick 13 — Game Tick 2602, Day 10, 06:10 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 99% (+22!) | Thirst 56% | Fatigue 32% | Social 62% | Happiness 57% | Stress 12%
**Inventory:** bread x2 (-1), gold x67
**Memory:** "Ate bread (+25 hunger)."
**Action taken:** `drink` at Garden Well (+5 min travel) — Restore thirst
**Reasoning:** Thirst at 56%, need to drink. Testing drink action with travel.

**Key observations:**
- **Eat action confirmed working!** Bread restores +25 hunger (not +8 like wheat). Different food types have different restore values.
- Bread consumed correctly (3 -> 2).
- Time advanced significantly: 05:55 -> 06:10 in game time. ~15 game minutes for eat action.

### Tick 14 — Game Tick 2617, Day 10, 08:05 (morning)
**Location:** Garden Well (well)
**Stats:** HP 100 | Hunger 98% | Thirst 99% (fully restored!) | Fatigue 34% | Social 55% | Happiness 56% | Stress 11%
**Inventory:** bread x2, gold x67
**Weather:** Clear (changed from cloudy!)
**NPCs nearby:** None
**Resources here:** water: 150, clay: 8
**World events:** "Strange omens in the sky" — great boar with iron tusks stalking Whispering Reeds forest
**Memory:** "Drank fresh water at the well (thirst fully restored)."
**Action taken:** `gather_clay` at Garden Well (here) — Test gather_clay with local clay resources
**Reasoning:** Garden Well has clay: 8. Testing gather_clay at a location with actual clay resources. This tests the fix for BUG-J1.

**Key observations:**
- **Drink action confirmed working!** Thirst fully restored (60% -> 100%). Uses 1 well water.
- Successfully traveled to Garden Well (+5 min travel). Location correctly shows "(here)" for actions.
- Garden Well has clay: 8 — good test for gather_clay.
- Memory system shows "Defining moments" with (vivid) markers for important events.
- Time is now morning (08:05). Weather changed from cloudy to clear. Day 10.
- World event text is atmospheric and well-written. Mentions specific location (Whispering Reeds).
- Alone at the well — no social actions available.

### Tick 15 — Game Tick 2633, Day 10, 09:25 (morning)
**Location:** Garden Well
**Stats:** HP 100 | Hunger 97% | Thirst 97% | Fatigue 45% (+11!) | Social 47% | Happiness 54% | Stress 9%
**Inventory:** bread x2, gold x67, clay x1 (+1 new!)
**Memory:** "Gathered 1 clay near the water."
**Action taken:** `explore` — Wander and discover new locations
**Reasoning:** Testing explore mechanic. Want to discover new areas.

**Key observations:**
- **Gather clay confirmed working at Garden Well!** +1 clay, +10 fatigue (from code), total fatigue increase +11 (includes travel fatigue from market to well).
- Clay resource at well shows "clay: 8" still — but code should have decremented to 7. Possible display issue? Or the observe might be cached.
- **DESIGN NOTE:** Fatigue jumped from 34% to 45% from one gather_clay action. That's significant. If actions keep adding fatigue, NPCs need sleep frequently.

### Tick 16 — Game Tick 2648, Day 10, 10:40 (morning)
**Location:** South Gate Forge
**Stats:** HP 100 | Hunger 96% | Thirst 95% | Fatigue 55% | Social 39% | Happiness 53% | Stress 9%
**Inventory:** bread x2, gold x67, clay x1
**Weather:** Rain
**NPCs nearby:** None
**World event:** "Strange omens in the sky" — pale moon, frost, voices of hunger
**Actions available:** drink, sleep(!), forage, hunt, farm, fish, gather_thatch, gather_clay, buy_supplies, explore, travel, start_business
**Action taken:** `forage` at Breezy Forest Grove (+5 min travel) — Test foraging
**Reasoning:** Testing forage mechanic. Close forest available.

**Key observations:**
- **Explore worked!** Traveled from Garden Well to South Gate Forge. Fatigue +10 from explore.
- **No find from explore** — "Wandered to South Gate Forge, taking in the sights." (20% chance to find item, didn't trigger)
- `sleep` now available! Fatigue at 55% (above 50% threshold). Sleep shows "+60 min travel" suggesting home is ~60 min away.
- `start_business` available at the forge — could claim it. But not testing that now.
- `eat` NOT available — hunger at 96% (above 90% threshold). Correct behavior.
- `go_home` NOT available — it's daytime (10:40 AM). go_home requires `w.IsNight()`. **DESIGN NOTE:** This seems overly restrictive — NPCs should be able to go home during the day too, especially if tired.
- No social actions — nobody else at the forge.

### Tick 17 — Game Tick 2664, Day 10, 11:15 (day)
**Location:** Breezy Forest Grove
**Stats:** HP 100 | Hunger 95% | Thirst 93% | Fatigue 65% | Social 31% | Happiness 50% | Stress 9%
**Inventory:** bread x2, gold x67, clay x1, berries x1 (+1 new!)
**Memory:** "Foraged in the woods and found 1 berries."
**Action taken:** `sleep` — Test sleep mechanic with travel to home/inn
**Reasoning:** Fatigue at 65% - high. Testing sleep destination routing (should go home per code fix).

**Key observations:**
- **Forage confirmed working!** Found 1 berries. Fatigue +8 from forage.
- Social dropping fast (39% -> 31% in 16 ticks). No NPCs to interact with in wilderness.
- Fatigue at 65% - past the 50% threshold for sleep.

### Ticks 18-24 — Sleep in progress
**Action:** sleep — traveling to home then sleeping
**Location:** Breezy Forest Grove (in transit)
**Notes:**
- Sleep committed at tick 2664, busy until tick 2725 (~61 ticks total).
- That's 240 game min (sleep) + ~60 min travel = 300 game min.
- At tick 2668: fatigue at 66.8% (up from 65%). This is because the NPC is still TRAVELING (walk fatigue).
- Once NPC arrives at home, DecayNeeds will reduce fatigue by 0.5/tick during sleep.
- Session still active at tick 2668 (~17 ticks played). BUG-J3 (token expiry) not triggered this session.

## Session Summary (17 ticks played, sleep in progress)

### Actions Tested (9 unique)
| Action | Tick | Result | Notes |
|--------|------|--------|-------|
| comfort | 1-2 | SUCCESS | -15 stress, -3 trauma, +10 relationship to target |
| scavenge | 3-5 | SUCCESS | Picked up 48 gold + 2 bread from ground |
| trade | 6-9 | SUCCESS | Sold 1 bread for 4 gold to Gareth |
| talk | 10-11 | SUCCESS | +20 social, +3 relationship, custom dialogue stored |
| eat | 12-13 | SUCCESS | Bread +25 hunger, consumed 1 |
| drink | 14 | SUCCESS | Thirst fully restored to 100% |
| gather_clay | 14-15 | SUCCESS | +1 clay, +10 fatigue |
| explore | 15-16 | SUCCESS | Traveled to South Gate Forge |
| forage | 16-17 | SUCCESS | Found 1 berries |
| sleep | 17+ | IN PROGRESS | 240 min + travel |

### Locations Visited
1. Golden Meadow Market (spawn) — market
2. Garden Well — well (water, clay)
3. South Gate Forge — forge
4. Breezy Forest Grove — forest

### Stat Trends (over 17 ticks)
- HP: 100 (stable)
- Hunger: 79% -> 99% (ate bread) -> 95% (passive drain)
- Thirst: 60% -> 56% (drain) -> 100% (drank) -> 93% (drain)
- Fatigue: 28% -> 65% (actions + travel) -> sleeping
- Social: 51% -> 68% (talk) -> 31% (passive drain, alone)
- Happiness: 59% -> 50% (gradual decline)
- Stress: 17% -> 9% (passive reduction)

