# Playtest Notes — Elara (Miner)

**Session:** 2026-03-31
**NPC:** Elara
**Profession:** Miner
**Location:** East Side Market
**Personality:** greedy, empathetic, neurotic, unscrupulous, introverted, addictive personality
**NPC ID:** 9d05bf6f-b7ee-4ac1-9def-d0a6265e0ac3

---

## Tick Log

### Tick 1 (Game Tick 1554) — Scavenge Test
- **State:** HP 100, Hunger 65, Thirst 60, Fatigue 20, Social 74, Happiness 59, Stress 12
- **Location:** East Side Market (market)
- **Weather:** cloudy, Day 34 13:10
- **Observed:** 15 NPCs here! Including 7 other Elaras (guard, potter, miner, herbalist, Elara Thorne guard, miller, miner). Also 3 Eldrics, 2 Eamons, 1 Eldrin, 1 Kaelen Thorne.
- **Items on ground:** Many bread x2 piles + gold x15 piles (8+ gold piles = 120+ gold on ground)
- **DUPLICATE NAME BUG:** 8 Elaras at one market (including us). Even worse than PR #9 which found 6 at one inn.
- **check_location "The Gilded Inn":** Shows "~0 min" travel time despite being +5 min. BUG-G1 from PR #9 CONFIRMED (not merged).
- **inspect_person "Eldrin":** Works correctly - shows potter, mood desperate, relation 0.
- **Available actions:** eat, drink, forage, farm, fish, gather_thatch, gather_clay, scavenge, trade, talk, gift, eat_together, work_together, comfort, steal, explore, travel, start_business
- **NOTE:** trade action IS available at market! Previous playtests couldn't test this.
- **NOTE:** No mine_ore or mine_stone action despite miner profession (skill 37). Need to be at a mine?
- **NOTE:** steal action available (unscrupulous personality trait?)
- **NOTE:** eat_together action available - new social action to test
- **Action:** scavenge — Testing gold cap with 120+ gold on ground
- **Inventory before:** bread x2, gold x15

### Tick 2 (Game Tick 1556) — Busy (scavenge in progress)
- **State:** Busy until tick 1558. HP 100, Hunger 65, Thirst 60, Fatigue 20.
- **NOTE:** Scavenge takes 3 ticks (15 min at 5 min/tick) at same location.
- **Code review findings while waiting:**
  - **BUG: SpawnOnDemand gives no profession equipment!** Miners need pickaxes but spawn with only bread+gold. See `server/god/spawn_on_demand.go` line 111 vs `genesis.go` line 220-221. This means miners can NEVER mine until they somehow get a pickaxe.
  - **check_location bug confirmed in code:** `npc_tools.go` lines 452, 461, 466-467 use `params.LocationID` (name string) instead of `loc.ID` (resolved ID) for NPCsAtLocation, EnemiesAtLocation, and TravelTicks. PR #9 fixes this but is not merged.
  - **Sleep travel fatigue (DESIGN-E1) root cause:** `turns.go` line 342-343 applies travel fatigue immediately when action is submitted, even for sleep actions. Fix: skip travel fatigue when actionID is "sleep".

### Tick 3 (Game Tick 1558) — Trade Test
- **State:** HP 100, Hunger 65, Thirst 59, Fatigue 21, Social 72
- **Scavenge Result:** Picked up 24 bread + 20 gold. **GOLD CAP AT 20 CONFIRMED (PR #7).**
- **Inventory:** bread x46, gold x35
- **Ground items remaining:** gold x100 (10+15x6 piles). Gold left behind correctly.
- **REGRESSION CHECK:** PR #7 scavenge gold cap CONFIRMED WORKING
- **REGRESSION CHECK:** PR #7 list_actions location limit CONFIRMED WORKING (shows 10 + "... and N more")
- **NOTE:** No mine_ore or mine_stone actions visible despite miner profession (skill 37). Confirmed in code: requires pickaxe equipped. SpawnOnDemand doesn't give pickaxes.
- **NOTE:** trade action available at market! First time tested in any playtest.
- **Action:** trade with Eldrin — Testing trade mechanics with bread

### Tick 4 (Game Tick 1573) — Trade Result + Eat Together
- **State:** HP 100, Hunger 64, Thirst 58, Fatigue 22, Social 65, Happiness 59, Stress 10
- **Trade Result:** Sold 1 bread to Eldrin for 3 gold. **TRADE ACTION WORKING!**
- **Inventory:** bread x45, gold x38 (sold 1 bread for 3 gold)
- **Relationship:** Eldrin now at 5 (was 0). Trade builds relationships.
- **Merchant skill:** 0 (just unlocked from trade)
- **BUG: Trade doesn't store memory for seller!** Only buyer (target) gets economic memory. Seller's trade is not categorized, so recall_memories with "economic" filter misses it.
- **inspect_person "Elara":** Returned the guard Elara (not self). FindNPCByNameAtLocation returns first match in NPC list.
- **NOTE:** 16 NPCs now at market (new Elara miner spawned). 8+ Elaras at one location!
- **NOTE:** recall_memories works — shows scavenge and trade results.
- **World event:** "Strange omens in the sky" with dramatic description.
- **Social decay:** 74 -> 65 over ~19 ticks = ~0.47/tick (matches config 0.5/tick)
- **Action:** eat_together with Eamon — Testing social + hunger combo action

### Tick 5 (Game Tick 1577) — Eat Together Result + Steal Test
- **State:** HP 100, Hunger 78, Thirst 57, Fatigue 22, Social 78, Happiness 58, Stress 10
- **eat_together Result:** Shared bread with Eamon. +15 hunger, +15 social (displayed as ~+14/+13 after decay). 1 bread consumed. Relationship +8. Dialogue stored in memory.
- **CONFIRMED:** eat_together action working correctly. Good combo of social + hunger recovery.
- **Inventory:** bread x44, gold x38
- **Relationship:** Eamon: +8 (slight positive)
- **Action:** steal from Eldric Swiftwind — Testing steal/crime mechanic

### Tick 6 (Game Tick 1583) — Steal Result + Travel to Shrine
- **State:** HP 100, Hunger 78, Thirst 57, Fatigue 23, Social 75, Happiness 58, Stress 9
- **Steal Result:** "Quietly stole 2 gold from Eldric Swiftwind." SUCCESS - wasn't caught!
- **Gold:** 38 -> 40 (+2 stolen). Reputation: 49 -> 46 (-3 crime penalty).
- **Relationship:** Eldric Swiftwind now "slight negative (steal)"
- **check_location bugs confirmed:** Saints Rest Shrine shows "~0 min" (should be 5 min). Iron Anvil Forge shows "~0 min" (should be 10 min). Neither shows NPCs present. BUG-G1 from PR #9 confirmed.
- **Action:** travel to Saints Rest Shrine — Testing shrine area

**6-tick checkpoint update:** Key findings so far:
1. **Scavenge gold cap WORKING (PR #7 confirmed)** - 20 gold max
2. **list_actions location limit WORKING (PR #7 confirmed)** - shows 10 + "...and N more"
3. **Trade action WORKING** - Sold bread for 3 gold at market
4. **eat_together WORKING** - +15 hunger, +15 social, relationship +8
5. **Steal action WORKING** - Stole 2 gold, reputation -3, relationship negative
6. **check_location "~0 min" BUG** - PR #9 fix not merged, confirmed present
7. **BUG: SpawnOnDemand missing profession equipment** - Miners get no pickaxe, can't mine
8. **BUG: Trade memory not stored for seller** - Only buyer gets CatEconomic memory
9. **DUPLICATE NAMES: 8+ Elaras at one market**
10. **Social decay rate:** ~0.5/tick, matching config

### Tick 7 (Game Tick 1589) — Shrine Visit + Travel to Forge
- **Location:** Saints Rest Shrine (shrine) — empty, no NPCs, no items on ground
- **State:** HP 100, Hunger 78, Thirst 56, Fatigue 27, Social 73 (est), Happiness 58, Stress 9
- **Travel fatigue:** +4 fatigue from travel (23 -> 27, 1 tick travel = ~3 fatigue per turns.go formula)
- **NO PRAY ACTION at shrine!** Despite being AT a shrine. Conditions require SpiritualSensitivity > 65 (believer) or stress > 85 / HP < 30 / happiness < 15 (desperate). Our NPC meets neither.
- **DESIGN: pray conditions too restrictive** — most NPCs can't use shrines at all, making them pointless locations
- **No scavenge available** (nothing on ground). No social actions (nobody here).
- **Memory system:** Storing defining/vivid moments correctly. Scavenge and trade marked as vivid.
- **Action:** travel to Iron Anvil Forge — Testing forge area for craft/mining actions

### Tick 8 (Game Tick 1594) — Forge Visit + Drink
- **Location:** Iron Anvil Forge (forge) — empty, no NPCs, no items
- **State:** HP 100, Hunger 77, Thirst 55, Fatigue 31
- **No forge-specific actions!** No forge_weapon, forge_tool, smelt_ore despite being AT a forge. Need blacksmith skill >= 25 AND worker status AND materials.
- **No mine actions** — still no pickaxe (SpawnOnDemand bug)
- **DESIGN: Forge useless to non-blacksmiths** — miners who need tools can't use the forge
- **Action:** drink at Old Well (+5 min travel) — restoring thirst

### Tick 9 (Game Tick 1600) — Drink Result + Gather Clay
- **Location:** Old Well — water 149, clay 8. Nobody here.
- **State:** HP 100, Hunger 77, Thirst 100 (restored!), Fatigue 32, Social 66, Happiness 56, Stress 6
- **Drink Result:** "Drank fresh water at the well (thirst fully restored)." CONFIRMED WORKING.
- **Social decay:** 78 -> 66 over ~23 ticks since eat_together = ~0.52/tick (matches config)
- **Action:** gather_clay at Old Well (here) — Testing gathering at current location

### Tick 10 (Game Tick 1607) — Gather Clay Result + Travel to Inn
- **State:** HP 100, Hunger 76, Thirst 99, Fatigue 43, Social 63, Happiness 56, Stress 5
- **gather_clay Result:** "Gathered 2 clay near the water." Fatigue +11 (32->43). CONFIRMED WORKING.
- **Inventory:** bread x44, gold x40, clay x2
- **Action:** travel to The Gilded Inn (+5 min from well) — Testing inn mechanics

### Tick 11 (Game Tick 1613) — Inn Visit + Work Together
- **Location:** The Gilded Inn — 14 NPCs! 7 Elaras (baker x2, scribe, tailor, shepherd x2, miner).
- **State:** HP 100, Hunger 76, Thirst 98, Fatigue 47, Social 60 (est)
- **Items on ground:** 79 gold + 17x15 gold = 334 gold total + lots of bread. MASSIVE gold accumulation!
- **New actions at inn:** craft_pottery, write_journal, work_together, scavenge
- **NOTE:** "Stranger" NPC at inn — spawn_on_demand name fallback when LLM parse fails
- **NOTE:** write_journal action available — untested in previous playtests!
- **NOTE:** No eat_together — hunger > 70% threshold blocks it
- **DESIGN: Gold accumulation at inns** — 334 gold at this single inn despite PR #7 faster decay fix
- **Action:** work_together with Thalric Ironhand (miner) — Testing same-profession cooperation

### Tick 12 (Game Tick 1620) — Work Together Result + Talk
- **State:** HP 100, Hunger 75, Thirst 97, Fatigue 59, Social 56, Happiness 55, Stress 4
- **work_together Result:** "Worked together with Thalric Ironhand - improved miner skills and bond." +12 fatigue, +0.6 miner skill (37->38), relationship +5.
- **DESIGN: work_together doesn't recover social!** Despite being categorized as "social" action, it doesn't modify SocialNeed. Working with someone should help loneliness.
- **Fatigue climbing:** 59% now, approaching sleep threshold (>50 for sleep to appear).
- **Action:** talk to Stranger — Testing talk + social recovery + Stranger NPC

### Tick 13 (Game Tick 1626) — Talk Result + Craft Pottery
- **State:** HP 100, Hunger 75, Thirst 97, Fatigue 60, Social 74, Happiness 57, Stress 3
- **Talk Result:** "Chatted with Stranger (social +20, relationship +3)." Social: 56 -> 74 after decay. CONFIRMED.
- **Sleep available!** Shows +160 min travel to home (Rat's Nest). Fatigue >= 50 threshold met.
- **write_journal disappeared** — social need dropped below 30 after talk. Condition: stress > 40 OR socialNeed > 30.
- **Action:** craft_pottery — Testing 2 clay -> ceramic crafting

### Tick 14 (Game Tick 1637) — Craft Pottery Result + Comfort
- **State:** HP 100, Hunger 74, Thirst 95, Fatigue 69, Social 68, Happiness 56, Stress 2
- **craft_pottery Result:** "Shaped clay into a ceramic piece." 2 clay -> 1 ceramic. +9 fatigue. CONFIRMED WORKING.
- **Inventory:** bread x44, gold x40, ceramic x1
- **inspect_person Eldrin:** Mood desperate, stressed, thirsty, lonely. (different Eldrin than market one - this is a miner)
- **Action:** comfort Eldrin — Testing comfort action on stressed NPC

### Tick 15 (Game Tick 1643) — Comfort Result + Gift
- **comfort Result:** "Comforted Eldrin (-15 stress, -3 trauma, relationship +10)." CONFIRMED WORKING.
- **Relationship:** Eldrin now "slight positive (comfort)". +10 relationship.
- **Dialogue stored:** "It will be alright, friend. The mines will be waiting for us tomorrow."
- **State:** HP 100, Hunger 74, Thirst 95, Fatigue 70
- **Action:** gift bread to Eldrin — Testing gift food action

### Tick 16 (Game Tick 1648) — Gift Result + Sleep (DESIGN-E1 Test)
- **State before sleep:** HP 100, Hunger 73, Thirst 94, Fatigue 70, Social 72, Happiness 61, Stress 0, Rep 49
- **Gift Result:** "Gave bread to Eldrin who was hungry (rep +3, relationship +12)." Bread 44->43. Rep 46->49. Relationship +12.
- **Gift CONFIRMED WORKING** — consumes food, grants rep and relationship.
- **BUG CONFIRMED: DESIGN-E1 Sleep travel fatigue!** Fatigue jumped from 70 to 79.7 (+9.7) INSTANTLY on sleep commit! Travel to home (160 min = 32 ticks). Formula: 3.0 * 32 / 10.0 = 9.6 fatigue.
- **If fatigue was 90+, NPC would collapse during transit to bed!**
- **Sleep duration:** 80 ticks (32 travel + 48 sleep = 400 min game time). Very long.
- **Action:** sleep at home (Rat's Nest, +160 min travel)

### Ticks 17-Sleep (Game Ticks 1649-1731) — Sleep Journey
- **Sleep trajectory:** Fatigue 79.7 (commit) -> 74.2 (t1660) -> 59.2 (t1690) -> 43.7 (t1721) -> 0.2 (t1731 with -40 from sleep)
- **Fatigue decay during sleep:** ~0.5/tick (matches config)
- **Sleep result:** "Slept at home (-40 fatigue, -5 stress)." FREE home sleep at Rat's Nest inn!
- **REGRESSION CHECK: PR #6 inn-home sleep fix CONFIRMED** — Free home sleep at inn, not rough sleep.
- **Eldrin died during sleep!** "Eldrin has died. I feel grief." World event during sleep.
- **DESIGN: Social decays during sleep!** Social: 72 -> 31 over 80 ticks. -0.5/tick even while sleeping.
- **Total sleep time:** 80 ticks (32 travel + 48 sleep = ~400 game min)

### Tick 17 (Game Tick 1731) — Post-Sleep + Eat
- **Location:** Rat's Nest (inn/home) — 15 NPCs! 7 Elaras, 2 Eldrins (duplicates!), 3 Eamons, 2 Eldrics.
- **State:** HP 100, Hunger 70, Thirst 89, Fatigue 0, Social 31 (!), Happiness 46, Stress 0
- **Ground items:** 165 gold + lots of bread at this inn
- **DESIGN: Social still drops during sleep** — NPCs wake up lonely after long sleep. Social 72 -> 31.
- **Action:** eat — Restoring hunger from 70%

### Tick 18 (Game Tick 1736) — Eat Result + Explore
- **Eat Result:** "Ate bread (+25 hunger)." Hunger 70 -> 95. Bread 43->42. CONFIRMED WORKING.
- **Action:** explore — Wandered to Hill Well from Rat's Nest. +9.7 fatigue from travel.

### Tick 19 (Game Tick 1749) — Explore Result + Travel to Market
- **Location:** Hill Well — arrived from explore
- **State:** HP 100, Hunger 94, Thirst 87, Fatigue 10, Social 25 (est), Happiness ~44
- **Explore Result:** "Wandered to Hill Well, taking in the sights." CONFIRMED WORKING.
- **Action:** travel to East Side Market — returning for final tests

### Tick 20 (Game Tick 1800) — Trade + Talk at Market
- **Trade Result:** Sold 1 bread to Caelum for 3 gold. **DESIGN: trade always sells first non-gold item in inventory** — no choice of what to sell. Ceramic (value 4) never sold because bread comes first.
- **Talk Result:** "Chatted with Caelum (social +20, relationship +3)." Relationship with Caelum: trade + talk = +8 total.
- **Dynamic pricing confirmed:** Bread base price 2, sold for 3 (higher demand).

### Tick 21 (Game Tick 1810) — Final Scavenge + Session Summary
- **Scavenge Result:** 20 gold (15+5, gold cap again!) + 19 bread. **GOLD CAP CONFIRMED** for third time.
- **Final State:** HP 100, Hunger 90, Thirst 80, Fatigue 28, Social 16 (!!), Happiness 40, Stress 0
- **Inventory:** bread x60, gold x63, ceramic x1
- **Mood:** lonely (Social at 16%)
- **Skills:** miner 38, merchant 1
- **Relationships:** Eldrin +22, Caelum +8, Eamon +8, Thalric Ironhand +5, Stranger +3
- **Reputation:** 49

---

## Session Summary

### Actions Tested (21 unique actions/observations)
1. observe - WORKING
2. check_self - WORKING
3. inspect_person - WORKING (but self-targeting risk with duplicate names)
4. list_actions - WORKING (PR #7 location limit confirmed)
5. check_location - BUGGY (0 min travel, no NPCs shown)
6. recall_memories - WORKING (but missing seller trade category)
7. scavenge - WORKING (gold cap at 20 confirmed 3x)
8. trade - WORKING (sold bread for 3 gold with dynamic pricing)
9. eat_together - WORKING (+15 hunger, +15 social, +8 relationship)
10. steal - WORKING (stole 2 gold, rep -3, relationship negative)
11. travel - WORKING (correct routing and fatigue)
12. drink - WORKING (thirst fully restored)
13. gather_clay - WORKING (+2 clay, +11 fatigue)
14. work_together - WORKING (+skill, +5 relationship, but NO social recovery)
15. talk - WORKING (+20 social, +3 relationship, dialogue stored)
16. craft_pottery - WORKING (2 clay -> 1 ceramic)
17. comfort - WORKING (-15 stress, -3 trauma, +10 relationship)
18. gift - WORKING (bread given, +3 rep, +12 relationship)
19. sleep - WORKING (free at home inn, -40 fatigue, -5 stress, PR #6 confirmed)
20. eat - WORKING (+25 hunger from bread)
21. explore - WORKING (random location, ~10 fatigue)

### Actions NOT Tested
- hunt, forage, fish, chop_wood, farm (resource gathering - no time)
- mine_ore, mine_stone (BLOCKED - no pickaxe from SpawnOnDemand)
- start_business, serve_customer, hire_employee (economy actions)
- forge_weapon, forge_tool, smelt_ore (need blacksmith skill)
- write_journal (conditions not met - need stress > 40 or social > 30)
- go_home (tested sleep instead which routes home)
- pray (conditions too restrictive at shrine)
- Any guard-specific actions (NONE EXIST)

### Locations Visited (8 total)
1. East Side Market (market) — spawn point, 16 NPCs, trade working
2. Saints Rest Shrine (shrine) — empty, no pray action available
3. Iron Anvil Forge (forge) — empty, no craft actions for non-blacksmiths
4. Old Well (well) — water 149, clay 8, nobody here
5. The Gilded Inn (inn) — 14 NPCs, 334 gold on ground
6. Hill Well (well) — from explore
7. Rat's Nest (inn/home) — 15 NPCs, free home sleep
8. Back to East Side Market

