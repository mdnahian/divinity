# Playtest Notes — Eldric Warmth

**Session:** 2026-03-25
**NPC:** Eldric Warmth (Innkeeper at Golden Meadow Market)
**Personality:** wise, neurotic
**NPC ID:** c1ddc340-0bfd-4e31-b821-83978a26a7e5

---

## Tick Log

### Tick 1 — Game Tick 1998, Day 8, 04:30 (night)
**Location:** Golden Meadow Market (market)
**Stats:** HP 100 | Hunger 64% | Thirst 62% | Fatigue 23% | Social 57% | Happiness 71% | Stress 11% | Hygiene 79% | Sobriety 100% | Reputation 46
**Inventory:** bread x2, gold x15
**Weather:** Clear
**NPCs nearby:** Soren Greenleaf (scribe, miserable), Countess Zarah (countess, anxious), Eamon Stonefist (miner, miserable), Lysandra Brightwater (healer, miserable), Serafim the Compassionate (healer, miserable), Mara Moonshadow (healer, anxious), Martha Hearthwarm (innkeeper, neutral)
**Items on ground:** gold x13
**Events:** Strange omens in the sky (fog/gathering message)
**Actions available:** eat, drink, forage, hunt, farm, fish, gather_thatch, gather_clay, scavenge, trade, buy_supplies, offer_counsel, talk, gift, eat_together, work_together, comfort, steal, explore, travel, go_home, start_business
**Action taken:** `scavenge` — Pick up gold x13 from the ground. Testing this action for the first time.
**Notes:**
- **REGRESSION CHECK BUG-J1 (gather_clay no locations): FIXED!** gather_clay now shows full location list with names and travel times, just like drink/fish/farm. This was a bug in the previous playtest where gather_clay showed "+5 min travel" with no location names.
- **REGRESSION CHECK BUG-J4 (Duplicate world events): POSSIBLY FIXED** — Only one event text this tick, not duplicated. Need more observation.
- `offer_counsel` is a new action not seen before — "Counsel a stressed NPC for gold (elder)" — profession/age-specific?
- `work_together` available — can work with Martha Hearthwarm (same profession innkeeper)
- `comfort` available — many stressed/miserable NPCs here to comfort
- Many NPCs are "miserable" — 4 out of 7. This seems like a lot. Balance issue?
- Market has trade and buy_supplies available "(here)" — good market functionality.
- Innkeeper skill at 33. Employed (unspecified employer).
- `start_business` available — will test this to check DESIGN-6 (business has no benefit)

### Tick 2 — Game Tick 2001, Day 8, 04:45 (night)
**Location:** Golden Meadow Market (market)
**Stats:** HP 100 | Hunger 64% | Thirst 62% | Fatigue 23% | Social 55% | Happiness 71% | Stress 10% | Hygiene 79% | Sobriety 100% | Reputation 46
**Inventory:** bread x2, gold x28 (+13 from scavenge!)
**Weather:** Clear
**NPCs nearby:** Same 7 NPCs as Tick 1
**Memory:** "Picked up loot from the ground: 13x gold"
**Actions available:** Same as Tick 1 (eat, drink, forage, hunt, farm, fish, gather_thatch, gather_clay, trade, buy_supplies, offer_counsel, talk, gift, eat_together, work_together, comfort, steal, explore, travel, go_home, start_business)
**Action taken:** `offer_counsel` to Soren Greenleaf — Test the elder/wisdom counsel action
**Notes:**
- Scavenge SUCCESS: Picked up all 13 gold from the ground. Gold now 28. Action works well.
- Stats changed since Tick 1: Hunger -0.2, Thirst -0.3, Fatigue +0.3, Social -2. Decay rates match config.
- `scavenge` no longer appears in actions list since items are gone from ground. Good.
- `offer_counsel` conditions: profession "elder" OR GetEffectiveWisdom >= 60. Eldric is "wise" personality, age 45. Interesting that non-elders can counsel.
- All 4 "miserable" NPCs still miserable. Many potential targets for comfort/counsel.

### Tick 3 — Game Tick 2002, Day 8, 04:50 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 64% | Thirst 61% | Fatigue 24% | Social ~55%
**Busy:** YES — offer_counsel, busy_until_tick 2007 (5 ticks remaining)
**Notes:**
- offer_counsel scheduled for ~30 min (6 ticks at 5 min/tick). Currently 5 ticks remaining. Matches expected ~30 min duration.
- Stats draining slowly as expected per config: Hunger -0.069/tick, Thirst -0.116/tick, Fatigue +0.1/tick.

### Tick 4 — Game Tick 2002, Day 8, 04:50 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 64% | Thirst 61% | Fatigue 24%
**Busy:** YES — offer_counsel, busy_until_tick 2007 (same as tick 3; server tick hasn't advanced)
**Notes:**
- Tick counter still at 2002 — server ticks appear to be 60s intervals, so 300s sleep covers ~5 ticks. Still waiting for action to finish at tick 2007.
- **OBSERVATION: The game tick counter didn't advance between my two checks despite 300s passing.** This could mean the server is paused or the tick interval is longer than expected. Will monitor.

### Tick 5 — Game Tick 2003, Day 8, 04:55 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 64% | Thirst 61% | Fatigue 24%
**Busy:** YES — offer_counsel, busy_until_tick 2007 (4 ticks remaining)
**Notes:**
- Game tick advanced from 2002 to 2003 (1 tick in ~300s). Server tick interval appears to be ~300s (5 min), not 60s as configured. This is much slower than config suggests.
- Decay per tick confirmed: Hunger -0.069, Thirst -0.116, Fatigue +0.1. All matching config values.

### Ticks 6-7 — Game Tick 2007, Day 8, 05:15 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 63% | Thirst 61% | Fatigue 24% | Social 52% | Happiness 71% | Stress 9% | Reputation 46
**Inventory:** bread x2, gold x29 (+1 from counsel)
**Memory:** "Counseled Soren Greenleaf (-15 stress, +5 happiness) for 1 gold"
**Relationship:** Soren Greenleaf: slight positive (relation 4)
**Action taken:** `work_together` with Martha Hearthwarm (fellow innkeeper)
**Notes:**
- offer_counsel SUCCESS: Earned 1 gold, Soren got -15 stress and +5 happiness. Relationship +4 with Soren.
- Server tick rate is ~5 min per tick in real time. offer_counsel took 6 ticks (from 2001 to 2007) = ~30 game minutes = 30 real minutes. This is very slow for playtesting.
- Soren still "miserable" even after counsel. His stress was reduced by 15 but mood label unchanged. Mood calculation may need higher threshold to change from "miserable".
- **OBSERVATION:** `offer_counsel` available because Eldric is "wise" personality + age 45. The condition checks `GetEffectiveWisdom` >= 60 OR profession "elder". Good that non-elders with wisdom can counsel.
- Countess Zarah now "miserable" (was "anxious" in Tick 1). Moods are deteriorating for many NPCs. 6 of 7 NPCs are now "miserable".
- **DESIGN OBSERVATION: Mass misery** — Almost all NPCs at the market are miserable. This suggests stats are draining too fast or there are insufficient ways for NPCs to self-sustain without player intervention.

### Tick 8 — Game Tick 2007, Day 8, 05:15 (night)
**Action committed:** `work_together` with Martha Hearthwarm
**Notes:**
- Testing same-profession collaboration. Should improve innkeeper skill for both.
- work_together: 30 base game minutes, ~6 ticks expected.

### Ticks 9-11 — Game Tick 2008, Day 8 (night)
**Location:** Golden Meadow Market
**Busy:** YES — work_together busy until 2013, currently 2008
**Notes:**
- Server tick rate has slowed significantly — sometimes only 1 tick per 10+ minutes.
- Force-interrupted work_together to test eat action and the interrupt/resume system.
- **Force interrupt test:** Successfully interrupted work_together. Response included `interrupted_action`, `resumable: true`, `resume_ticks_left: 5`. Interrupt/resume system working well!
- Committed eat action (force=true).
- **BUG FOUND: Force-interrupt loses resume state** — After force-interrupting work_together, the server's resume_action was empty. Code review confirmed: `SubmitExternalAction` in turns.go sets ResumeActionID on line 247 but unconditionally clears it on line 297-298. The API response to the client shows the resume info correctly (captured from NPC state before the function clears it), but the server state loses it, making the resume feature non-functional.

### Ticks 12-13 — Game Tick 2009+ (waiting for eat to complete)
**Pending action:** eat, busy_until_tick 2011
**Notes:**
- Token still valid after many ticks — **BUG-J3 (token expiry) NOT REPRODUCED.** Tokens appear to persist correctly now. May have been a server restart issue.
- Eat takes 3 ticks (~15 min game time). Waiting for completion.

### Summary of Tick Loop
Due to the slow server tick rate (~5 min per tick in real time), completed fewer game ticks than planned but gathered extensive observations from both gameplay and code review. Key actions tested:
- scavenge (SUCCESS)
- offer_counsel (SUCCESS)
- work_together (INTERRUPTED to test force system)
- eat (IN PROGRESS)
- Force interrupt/resume system (TESTED, bug found)
- talk (COMMITTED, waiting)

### Tick 14 — Game Tick 2013, Day 8, 05:35 (night)
**Location:** Golden Meadow Market
**Stats:** HP 100 | Hunger 88% (+25 from bread!) | Thirst 60% | Fatigue 25% | Social ~50%
**Inventory:** bread x1 (was x2), gold x29
**Memory:** "Ate bread (+25 hunger)"
**Action taken:** `talk` to Martha Hearthwarm
**Notes:**
- Eat SUCCESS: Consumed 1 bread, restored +25 hunger (63% -> 88%). Matches bread item `hunger_restore: 25`.
- Token STILL VALID after ~90+ real minutes. BUG-J3 definitely not reproduced.
- Committed talk to Martha Hearthwarm as final action.

### Session End Summary
- **Total game ticks observed:** ~15 (1998-2013)
- **Total real time:** ~120 minutes
- **Actions tested:** scavenge, offer_counsel, work_together (interrupted), eat, talk
- **Bugs found from gameplay:** 1 (force-interrupt resume loss)
- **Bugs found from code review:** 3 (gather [0] index, trade first-market-only, IsWorkerAtType [0])
- **Total bugs documented:** 4 (BUG-E1 through BUG-E4)
- **Token stability:** Excellent, no expiry issues

