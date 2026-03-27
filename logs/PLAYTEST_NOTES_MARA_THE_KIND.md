# Playtest Notes — Mara the Kind

**Session:** 2026-03-27
**NPC:** Mara the Kind (Healer at Golden Meadow Market)
**Personality:** brave, mentor
**NPC ID:** be2d337d-97cf-46a1-86db-0d383854c2ef

---

## Tick Log

### Tick 1 — Game Tick 4004, Day 15, 03:40 (night)
**Location:** Golden Meadow Market (market)
**Stats:** HP 100 | Hunger 59% | Thirst 71% | Fatigue 15% | Social 62% | Happiness 71% | Stress 13% | Hygiene 82% | Sobriety 100%
**Inventory:** bread x2, gold x15
**Home:** The Gilded Inn (+40 min travel from here)
**Weather:** Clear
**NPCs nearby:** Eldric (priest, mood: lonely, relation: neutral [0])
**Skills:** healer: 45
**Reputation:** 52
**Actions available:** eat, drink, forage, hunt, farm, fish, gather_thatch, gather_clay, trade, buy_supplies, talk, eat_together, steal, explore, travel, go_home, start_business
**Action taken:** `talk` to Eldric — Socialize with lonely priest

**Key Observations:**
- **Healer profession:** No healer-specific actions visible (no heal, brew_potion, etc.). Healer skill at 45 but no way to use it. Similar to BUG DESIGN-J1 (blacksmith no profession actions).
- **Market location:** `trade` and `buy_supplies` both available here with "(here)" tag. Golden Meadow Market is our home market.
- **gather_clay now shows locations!** Unlike Juna's playtest where gather_clay showed no locations, now it lists docks and wells. Possible fix or regression?
- **drink:** Lists 50+ wells including duplicates (multiple "Garden Well", "Hill Well", "Market Well" entries at different distances). Are these truly different locations or display bug?
- **go_home:** Shows +40 min travel to The Gilded Inn. NPC "lives" at an inn — interesting for testing sleep/home mechanics.
- **start_business:** Available! Can claim unowned establishment. Will test later.
- **No heal action:** A healer with no heal action. This is a design gap similar to blacksmith.
- **Reputation 52:** New stat not seen in previous playtests? Or just not noted.

### Tick 2 — Game Tick 4011, Day 15, 04:20 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 58% | Thirst 70% | Fatigue 16% | Social 78% (+16!) | Happiness 73% (+2) | Stress 12% (-1) | Hygiene 82% | Sobriety 100%
**Inventory:** bread x2, gold x15
**Weather:** Clear
**NPCs nearby:** Eldric (priest, mood: lonely, relation: 3)
**Memory:** "Chatted with Eldric (social +20, relationship +3)"
**Relationship:** Eldric — "slight positive (talk)"
**Action taken:** `eat` — Eat bread to restore satiation before it drops further

**Key Observations:**
- Talk results: Social 62 -> 78 (+16 displayed as +20 raw, dampened?), Happiness +2 (from +3 base), Stress -1.
- Relationship with Eldric now 3 (slight positive).
- **REGRESSION CHECK: BUG-2 (no solo eat) — eat action still available, FIXED confirmed.**
- Hunger dropped 59% -> 58% over 7 ticks (very slow decay). Thirst 71% -> 70%.
- Fatigue 15% -> 16% (slow natural increase).

### Tick 3 — Game Tick 4016, Day 15, 04:40 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 83% (+25!) | Thirst 69% | Fatigue 16% | Social 76% | Happiness 72% | Stress 11% | Hygiene 82%
**Inventory:** bread x1 (-1), gold x15
**Weather:** Clear
**NPCs nearby:** Eldric (priest, lonely, relation: 3)
**Memory:** "Ate bread (+25 hunger)"
**Action taken:** `trade` — Sell bread to Eldric at the market to test trade mechanics

**Key Observations:**
- Bread gives +25 hunger. Very good food item.
- Trade conditions check: We are at Golden Meadow Market. Need to verify if this is `markets[0]`.
- If trade works here, it means Golden Meadow Market IS the first market. Need to test at East Side Market later.

### Tick 4 — Game Tick 4021, Day 15, 05:15 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 83% | Thirst 69% | Fatigue 17% | Social 76% | Happiness 72% | Stress 11%
**Busy:** YES — trading, busy_until_tick 4023 (2 ticks remaining)
**Action:** Still executing trade with Eldric

### Tick 5 — Game Tick 4027, Day 15, 05:35 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 82% | Thirst 68% | Fatigue 17% | Social 70% | Happiness 71% | Stress 10% | Hygiene 82%
**Inventory:** gold x18 (+3 from trade, bread x0)
**Weather:** Rain (changed from clear!)
**NPCs nearby:** Eldric (priest, lonely, relation: 8)
**Memory:** "Sold 1 bread to Eldric for 3 gold (market price)"
**Action taken:** `forage` at Whispering Forest Edge (+5 min travel) — Get herbs for healing potions

**Key Observations:**
- **Trade SUCCESS at Golden Meadow Market!** Sold bread for 3 gold. Confirms trade works at this market.
- Trade gave: +3 gold, +5 relationship with Eldric (now 8), gained merchant skill.
- **BUG NOTE:** Trade condition checks `markets[0].ID` — this market works because it's likely the first. Need to test at East Side Market to confirm bug.
- Weather changed from clear to rain. Dynamic weather system working.
- Now have 0 food items. Need to forage for berries (food) and herbs (medicine for healer profession).
- **NEW FINDING:** Merchant skill gained at 0 level from one trade. check_self now shows "merchant: 0" — is this intentional to show 0 or a display bug?

### Tick 6 — Game Tick 4032, Day 15, 06:00 (night) — 6-TICK CHECKPOINT
**Location:** Whispering Forest Edge (traveled from market)
**Stats:** HP 69 (!!!) | Hunger 82% | Thirst 68% | Fatigue 18%
**Busy:** YES — foraging, busy_until_tick 4034
**Weather:** Rain
**COMBAT:** Attacked by wolves 3 times!
- Wolf attack 1: -17 HP (100 -> 83)
- Wolf attack 2: -8 HP (83 -> 77) (wrong - actually 83-8=75? memory says HP 77)
- Wolf attack 3: -9 HP (77 -> 69) (wrong - memory says from 77 to 69 = -8?)
**Memory:** "A wolf attacked me for 17/8/9 damage!"

**Key Observations:**
- **Wolf attacks are working!** 3 attacks over 3 ticks while foraging at forest edge.
- HP damage is real and dangerous: 100 -> 69 in 3 ticks. Need healing.
- **Note:** HP values in memories seem slightly inconsistent. 100-17=83 (correct), 83-8=75 but memory says 77, 77-9=68 but state shows 69. Possibly due to natural HP regen (+1/tick from DecayNeeds).
- Actually: HP regens +1/tick. So 83+1=84 at next tick, 84-8=76... still doesn't match. Minor rounding issue?
- No interrupt triggered because HP > 40 (threshold is HP<40 with enemies, or HP<15).
- **CRITICAL:** We're a healer with no medicine. Can't heal ourselves. Need to forage for herbs, brew potion, then heal.

### Tick 7 — Game Tick 4038, Day 15, 06:30
**Location:** Whispering Forest Edge (forest)
**Stats:** HP 23 (!!!) | Hunger 82% | Thirst 67% | Fatigue 27% | Social 64% | Happiness 70% | Stress 100% (!!) | Hygiene 82%
**Inventory:** gold x18, berries x2 (gained from forage!)
**Weather:** Rain
**Enemies:** Wolf (HP 54, STR 51), Bear (HP 71, STR 70)
**Ground loot:** Massive pile of loot from dead NPCs - gold x350+, herbs, berries, weapons, tools, journals, etc.
**Location history:** 5+ NPCs killed here! (Commoners, Torin Frostborn, Elian, Eldric, Vala Duskwalker)
**Mood:** Anxious (was content)
**Action taken:** `flee_area` — ESCAPE from deadly forest

**Key Observations:**
- **DEADLY FOREST!** Multiple NPCs have died here. Wolf and bear are permanent threats.
- Forage completed successfully: gained berries x2. But the 15% herb chance did not trigger.
- HP dropped from 69 to 23 over multiple attacks: wolf -19, bear -9, bear -17 = -45 more HP.
- HP regen (+1/tick) not enough to offset 10-20 damage/tick from enemies.
- **Stress maxed at 100%** from repeated attacks. Mood changed to "anxious".
- **MASSIVE LOOT ON GROUND** from previous deaths: 350+ gold, weapons, herbs, etc. The scavenge action would be very valuable here... if you could survive.
- **DESIGN NOTE:** This forest is essentially a death trap. Any NPC going there to forage/hunt will be attacked repeatedly. No warning shown in actions list about enemy presence.
- **BUG CANDIDATE:** `forage` Conditions (line 79-81, gather.go) checks `forests[0].Resources["berries"]` — should check the actual destination forest, not the first one.

### Tick 8 — Game Tick 4044, Day 15, 06:40 — DEATH
**Location:** Whispering Forest Edge
**HP:** 0 — DEAD
**Cause:** Killed by wolves while trying to flee
**Attack sequence:** Wolf -14 (HP 23->10), Wolf -18 (HP 10->0)
**resume_action:** flee_area (32 ticks remaining — she was mid-flee when killed)

**Key Observations:**
- **MARA THE KIND IS DEAD.** Survived 8 ticks total. Killed by wolves at Whispering Forest Edge.
- **DESIGN ISSUE:** Fleeing takes too long! 32 ticks remaining on flee when she died. Enemies attack every tick while you're fleeing. Flee should be faster or provide damage reduction.
- **DESIGN ISSUE:** No warning about enemies at destination. forage/hunt actions show forests but don't warn about enemy presence. NPCs walk into death traps.
- **DESIGN ISSUE:** HP regen (+1/tick) is vastly insufficient against enemy damage (8-19/tick). A healer with no medicine has no way to survive.
- **BUG:** NPC died but `alive: false` state is correct. `resume_action: flee_area` still set despite death.
- The forest is a GRAVEYARD — 5+ NPCs already dead there.

---

## Continuing as Elara the Compassionate

**NPC:** Elara the Compassionate (Healer at Golden Meadow Market)
**Personality:** conformist
**NPC ID:** d6e4efd6-4b54-4d9c-91e9-1913a5c8c92b
**Home:** The Gilded Inn

### Ticks 12-19: Garden Well, Farm, Dock
- Tick 12: `drink` at Garden Well — Thirst restored to 100%
- Tick 12: `gather_clay` at Garden Well — Gained 1 clay (well has clay:8, water:149)
- Tick 14: `farm` at Oakhaven Farm — Harvested 4 wheat, +12 fatigue
- Tick 16: `eat` bread — +25 hunger (63% -> 88%)
- Tick 17: `fish` at Dock — Caught nothing (+fishing experience). 60% failure rate confirmed.
- Tick 19: `scavenge` at Dock — MASSIVE haul! 8 fish, 17 gold, 1 herbs, 1 berries, 2 nets, 3 hooks
- Tick 20: `sleep` at The Gilded Inn (home) — testing sleep/home mechanics, fatigue at 57%
- Ticks 21-24: Still sleeping at The Gilded Inn. Fatigue decreasing: 57% -> 50% over 20 ticks.
  - Confirmed: NPC traveled to HomeID (The Gilded Inn) for sleep.
  - Confirmed: Fatigue decreases during sleep (~0.5/tick from DecayNeeds).
  - **REGRESSION CHECK BUG-5: FIXED!** Fatigue no longer rises during sleep.

**Key Observations:**
- gather_clay works correctly at a location with clay resources
- Farm produced 4 wheat as expected
- Fish has 40% success rate (missed this time)
- Scavenge is incredibly valuable — picked up everything on the ground
- **NOTE:** The dock ground items came from dead NPCs who fished here and died
- Weather: rain -> cloudy transition during these ticks

### Tick 9 — Game Tick 4044, Day 15, 07:00
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 66% | Thirst 71% | Fatigue 13% | Social 62% | Happiness 50% | Stress 16% | Hygiene 76%
**Inventory:** bread x2, gold x15
**Weather:** Rain
**NPCs nearby:** Eldric (priest, lonely, relation: 0)
**Action taken:** `explore` — Test explore mechanics
**Notes:**
- Explore action shows "+20 min travel" from list_actions. Dynamic based on random destination.

### Ticks 10-11 — Game Tick 2, Day 15, 07:55
**Location:** Silvermine Outpost (mine) — explored here
**Stats:** HP 100 | Hunger 65% | Thirst 70% | Fatigue 20% | Social 56% | Happiness 47% | Stress 15% | Hygiene 76%
**Inventory:** bread x2, gold x15
**Resources here:** stone: 20, iron_ore: 12
**Weather:** Rain
**Memory:** "Wandered to Silvermine Outpost, taking in the sights."
**Action taken:** `drink` at Garden Well (+10 min travel)
**Notes:**
- Explore successfully moved us to Silvermine Outpost. No items found (80% chance of nothing).
- **Tick counter reset** from 4050 to 2 — possible server restart or world reset between ticks.
- **start_business available** at Silvermine Outpost (mine with no owner). Interesting.
- **go_home NOT available** — source shows `w.IsNight()` check. It's 07:55 (daytime), so go_home blocked. This is correct.
- **No trade available** since we're not at a market.
- Fatigue increased normally (13%->20% over ~7 ticks).

