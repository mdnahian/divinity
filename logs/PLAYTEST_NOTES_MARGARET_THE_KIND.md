# Playtest Notes — Margaret the Kind

**Session:** 2026-03-29
**NPC:** Margaret the Kind (Innkeeper at Golden Meadow Market)
**Personality:** creative, dominant
**NPC ID:** 1243a51c-1074-4895-8394-fee2846150a4

---

## Tick Log

### Tick 1 (Game Tick 114, Day 25 03:35, Night, Cloudy)
- **Location:** Golden Meadow Market (market)
- **State:** HP 100, Hunger 67%, Thirst 57%, Fatigue 27%, Social 64%, Happiness 58, Stress 11
- **Inventory:** bread x2, gold x15
- **Observation:** Nobody here. World event: "Strange omens in the sky" with dramatic vision of NPCs dying alone. Night time.
- **Action:** eat (bread) - reason: hunger at 67%, need to maintain health
- **Notes:**
  - go_home available (+40 min travel) -- home is NOT at Golden Meadow Market
  - start_business available -- can claim an establishment
  - No scavenge action visible -- no items on ground? Or not available?
  - No profession-specific actions (serve_customer, brew_ale) -- not at an inn yet
  - Massive list of wells for drink/gather_clay -- showing ALL wells across entire world (50+)
  - **POTENTIAL BUG:** The list_actions drink list shows 50+ wells including distant ones (300+ min travel). This is an extremely long list that clutters the output. Should probably be limited to nearby locations only.

### Tick 2 (Game Tick 120, Day 25 04:05, Night, Cloudy)
- **Location:** Golden Meadow Market
- **State:** HP 100, Hunger 92% (+25 from eat), Thirst 56%, Fatigue 28%, Social 61%
- **Inventory:** bread x1, gold x15
- **Observation:** Still nobody here. Same world event active.
- **Action:** drink at Garden Well (+5 min travel)
- **Notes:**
  - "eat" action no longer shows up -- likely because hunger is at 92% (threshold?)
  - "trade" NOT available at market -- we're at a market but no trade action. Maybe need items to sell?
  - No profession-specific actions still (serve_customer, brew_ale, cook)
  - chop_wood, mine_stone, mine_ore not in action list at all -- maybe need specific conditions?
  - **POTENTIAL ISSUE:** eat action disappearing when hunger > 90% makes sense as a threshold, but should still be available for strategic eating

### Tick 3 (Game Tick 126, Day 25 04:35, Night, Cloudy)
- **Location:** Garden Well (well)
- **State:** HP 100, Hunger 91%, Thirst 100% (fully restored!), Fatigue 29%, Social 61%
- **Observation:** Resources: water 149, clay 8. Nobody here. drink correctly restored thirst to ~100%.
- **Action:** forage at Whispering Forest Edge (+5 min travel)
- **Notes:**
  - "buy_supplies" action now appears -- available at market locations. Interesting that it shows from Garden Well (market is 5 min away).
  - Garden Well shows "(here)" correctly in location lists
  - go_home still available (+35 min travel) -- home is within moderate distance
  - **CONFIRMED:** drink action fully restores thirst (56% -> 100%). Working as expected.

### Tick 4 (Game Tick 132, Day 25 05:05, Night, Cloudy)
- **Location:** Whispering Forest Edge (forest)
- **State:** HP 34 (DOWN FROM 100!), Hunger 91%, Thirst 99%, Fatigue 29%
- **COMBAT:** Wolf attacked for 16, bear attacked for 20, wolf for 15, wolf for 10 = 61 damage total! HP 100->34
- **Observation:** 7 enemies: 4 wolves (str 35-51), 2 dire wolves (str 55-59), 1 bear (str 70-71). This location is a death trap!
- **Ground items:** MASSIVE: gold x172, gold x64, gold x43, gold x32, gold x26, gold x19, gold x18, gold x15 (x7), gold x12 (x2), gold x11, gold x10 (x2), gold x9, gold x6, gold x14 + many items. Total ~500+ gold on ground.
- **Dead NPC records:** Elara killed by wolf, Eldric Stoneheart killed by bear, Elara killed by bear, Eamon killed by bear, Elian the Unseen killed by wolf.
- **Action:** scavenge (risky with enemies, but testing mechanics)
- **Resume action:** forage (1 tick / 5 min remaining) -- forage was INTERRUPTED by combat
- **Notes:**
  - **DESIGN ISSUE (SEVERE):** This forest has become a death zone. 7 enemies camping the only nearby forest. NPCs are dying and dropping all their gold, which attracts more NPCs to scavenge, who then also die.
  - **DESIGN ISSUE:** The ground item accumulation is far worse than previous reports -- 500+ gold at this one location.
  - **CONFIRMED:** resume action works correctly -- shows "Resume Forage for food in the forest (~5 min remaining)"
  - **CONFIRMED:** attack_enemy and flee_area appear when enemies present
  - **NOTE:** scavenge is risky here but testing if it picks up ALL items (previous playtests confirmed it does)
  - **BUG RISK:** By choosing scavenge instead of resume, the forage resume state should be preserved. Testing BUG-M5 (force-interrupt clears resume state).

### Tick 5 (Game Tick 137, Day 25 05:15, Night, Cloudy)
- **DEAD** -- HP 0. Killed by wolves during scavenge action.
- Wolf attacked for 18 (HP 34->17), wolf attacked for 23 (HP 17->0).
- **NPC Death after 5 ticks. Session ended early.**
- **Notes:**
  - Scavenge was interrupted by death. resume_action still shows "scavenge" with 2 ticks left.
  - **DESIGN ISSUE (CRITICAL):** The forest nearest to spawn (Whispering Forest Edge, only 5 min travel) is packed with 7 enemies including dire wolves and a bear. This is an extreme difficulty spike for new NPCs. An innkeeper with no combat skills had no chance.
  - **DESIGN ISSUE:** Non-combat actions (forage, scavenge) don't protect from enemy attacks. NPC has to choose between useful actions and survival. Should at least get a warning or automatic flee when HP is critical.
  - **DESIGN ISSUE:** The scavenge-death cycle is self-reinforcing: NPCs die, drop loot, more NPCs come to scavenge, they die too.

## NPC DEATH - Respawning New Character

**New NPC:** Thalric the Builder (Innkeeper at Golden Meadow Market)
**Personality:** empathetic, generous, timid, introverted
**NPC ID:** 57ebaee5-caf1-400f-80dd-a072f1a831c8
**Strategy:** Avoid Whispering Forest Edge. Travel to inn to test innkeeper actions, try start_business, social interactions.

---

### Tick 6 (Game Tick 138, Day 25 05:35, Night, Cloudy) [Thalric]
- **Location:** Golden Meadow Market
- **State:** HP 100, Hunger 73%, Thirst 67%, Fatigue 27%, Social 79%, Happiness 54, Stress 31
- **Inventory:** bread x2, gold x15
- **Observation:** Nobody here. No world event visible (the omens event may have ended).
- **Action:** travel to The Gilded Inn (+40 min travel)
- **Notes:**
  - Same spawn point as Margaret. All innkeepers spawn at Golden Meadow Market?
  - No world event this time (Margaret saw "Strange omens" but Thalric doesn't). Events may be time-limited.
  - Stress starts at 31 (higher than Margaret's 11). Personality affects starting stats?
  - Social starts at 79 (higher than Margaret's 64). Same observation.

### Tick 7 (Game Tick 144) [Thalric]
- **BUSY** traveling to The Gilded Inn. Busy until tick 149 (5 ticks remaining).
- Fatigue increased from 27 to 30 during travel (+3 from travel fatigue).

### Tick 8 (Game Tick 149, Day 25 06:30, Cloudy)
- **Location:** The Gilded Inn (inn)
- **State:** HP 100, Hunger 72%, Thirst 66%, Fatigue 34%, Social 79%
- **People here:** Martha Hearthward (innkeeper), Elara (hunter), Mara Hearthkeeper (innkeeper), Elara (shepherd), Elara (barmaid) -- 3 ELARAS!
- **Ground items:** gold x631, fishing net x1, hook x2 -- MORE gold hoarding
- **Action:** start_business -- claiming the inn
- **Notes:**
  - **CONFIRMED:** Duplicate NPC names -- 3 different "Elara" NPCs at same location (hunter, shepherd, barmaid).
  - **CONFIRMED:** Gold on ground issue continues -- 631 gold at The Gilded Inn.
  - No innkeeper profession actions available YET (serve_customer, brew_ale, cook not listed). Expecting them after claiming the inn.
  - work_together available (same profession NPCs: Martha Hearthward and Mara Hearthkeeper are innkeepers).
  - Social actions (talk, gift, comfort) all available with NPCs present.

### Tick 9 (Game Tick 155) [Thalric]
- **BUSY** with start_business. Busy until tick 158 (3 ticks remaining).
- Stats slowly declining: fatigue 34%, hunger 72%, thirst 65%.

### Tick 10 (Game Tick 160, Day 25 07:25, Rain)
- **Location:** The Gilded Inn (OWNED!)
- **State:** HP 100, Hunger 71%, Thirst 64%, Fatigue 35%, Social 68%
- **Inventory:** bread x2, gold x10 (was 15, paid 5 for business)
- **"You own a business."** -- confirmed in check_self
- **New actions unlocked:** serve_customer, hire_employee, write_journal
- **People here:** Martha Hearthward (innkeeper), Elara (hunter, miserable), Mara Hearthkeeper (innkeeper, miserable), Elara (shepherd, miserable)
- **Action:** serve_customer
- **Notes:**
  - **CONFIRMED:** start_business costs 5 gold and claims the establishment
  - **CONFIRMED:** New actions unlock after owning: serve_customer, hire_employee, write_journal
  - brew_ale and cook still NOT available -- may need ingredients (ale ingredients? food items?)
  - hire_employee targets "nearby unemployed NPC" -- interesting
  - Weather changed to rain
  - Elara (barmaid) left the inn -- only 4 NPCs now
  - **POTENTIAL BUG:** serve_customer with target "Elara" returned "invalid request body" -- might be duplicate name issue?
  - serve_customer without target works fine

### Tick 11 (Game Tick 166, Day 25 07:55, Rain)
- **Location:** The Gilded Inn (OWNED)
- **State:** HP 100, Hunger 71%, Thirst 64%, Fatigue 35%, Social 65%
- **Inventory:** bread x1 (was 2), gold x13 (was 10)
- **serve_customer result:** Served bread to Martha Hearthward for 3 gold. Used 1 bread from inventory, gained 3 gold.
- **Action:** talk to Elara
- **Notes:**
  - **CONFIRMED:** serve_customer works, earns 3 gold, consumes 1 bread from owner inventory, gives +2 relationship.
  - **CONFIRMED:** "barmaid: 0" skill appeared after serving -- new skill unlocked by action.
  - **CONFIRMED (PR #6):** inspect_person with "Elara" correctly found a LOCAL Elara (hunter) instead of a random global NPC. FindNPCByNameAtLocation fix working!
  - Martha Hearthward: mood lonely, HP 100, relation +2 after being served.
  - Elara (hunter): mood miserable, HP 100, hungry/stressed/lonely.
  - **OBSERVATION:** serve_customer auto-selected Martha Hearthward (the one with highest relation?). Didn't specify target.
  - **DESIGN OBSERVATION:** Innkeeper must use own food inventory to serve customers. If you run out of bread, you can't serve. Need a supply chain (buy_supplies -> serve).

### Tick 12 (Game Tick 171, Day 25 08:25, Rain)
- **Location:** The Gilded Inn (OWNED)
- **State:** HP 100, Hunger 71%, Thirst 63%, Fatigue 36%, Social 65%
- **talk result:** Social +20, relationship +3 with Elara. Dialogue stored in memory.
- **Relationship context:** "Elara: slight positive (talk)"
- **Action:** scavenge (631 gold + items on ground)
- **Notes:**
  - **CONFIRMED:** talk works correctly: +20 social, +3 relationship, dialogue stored.
  - **CONFIRMED:** Relationship labels work: "slight positive (talk)" after 1 talk.
  - write_journal disappeared from action list this tick -- might be one-time or conditional?
  - Still no brew_ale or cook -- these may need specific ingredients.

### Tick 13 (Game Tick 177, Day 25 08:40, Rain)
- **Location:** The Gilded Inn (OWNED)
- **State:** HP 100, Hunger 70%, Thirst 62%, Fatigue 36%, Social 80%, Happiness 56
- **Inventory:** bread x1, gold x644, fishing net x1, hook x2
- **scavenge result:** Picked up 631 gold, 1 fishing net, 2 hooks. ALL ground items collected.
- **Action:** comfort Mara Hearthkeeper
- **Notes:**
  - **CONFIRMED:** scavenge picks up ALL items from ground in one action. 631 gold in single scavenge!
  - **CONFIRMED (DESIGN ISSUE):** Economy completely broken. Started with 15 gold, now have 644 from one scavenge.
  - Social jumped from ~65% to 80% -- the +20 from talk is significant.
  - Happiness went up slightly (54 -> 56).

### Tick 14 (Game Tick 182, Day 25 09:20, Rain) -- 6-tick checkpoint
- **Location:** The Gilded Inn (OWNED)
- **State:** HP 100, Hunger 70%, Thirst 62%, Fatigue 37%, Social 80%
- **comfort result:** -15 stress, -3 trauma, +10 relationship with Mara Hearthkeeper. Dialogue stored.
- **Relationship context:** Elara: slight positive (talk), Mara: slight positive (comfort)
- **Action:** gift bread to Elara (hunter)
- **Notes:**
  - **CONFIRMED:** comfort works: -15 stress, -3 trauma, +10 relationship. Very effective.
  - Ground is now clear after scavenge (no items).
  - All 4 NPCs still present, most still miserable.
  - Mara relationship jumped to 10 from comfort (+10 is much more than talk's +3).

### Tick 15 (Game Tick 188, Day 25 09:35, Rain)
- **Location:** The Gilded Inn (OWNED)
- **State:** HP 100, Hunger 70%, Thirst 61%, Fatigue 37%, Social 84%, Reputation 58
- **Inventory:** gold x644, fishing net x1, hook x2 (NO BREAD LEFT)
- **gift result:** Gave bread to Elara, rep +3, relationship +12. Very effective.
- **Action:** hire_employee targeting Elara
- **Notes:**
  - **CONFIRMED:** gift works: rep +3, relationship +12, consumes 1 food item from inventory.
  - eat, serve_customer, gift all disappeared from action list -- all require food items in inventory.
  - **OBSERVATION:** Without food, the innkeeper can't perform core profession actions (serve_customer). Need food supply chain.
  - Reputation increased from 55 to 58 (from gift +3).
  - Relationships: Elara +15 (talk +3, gift +12), Mara +10 (comfort +10).

### Tick 16 (Game Tick 193, Day 25 10:00, Rain)
- **Location:** The Gilded Inn (OWNED)
- **hire_employee result:** Hired Elara to work at The Gilded Inn for 2 gold/day.
- **Action:** farm at Oakhaven Farm (+15 min travel from inn, +10 from market)
- **Notes:**
  - **CONFIRMED:** hire_employee works, costs 2 gold/day, targets an NPC.
  - hire_employee disappeared from action list after hiring Elara -- may need more unemployed NPCs nearby.
  - **DESIGN NOTE:** The innkeeper gameplay loop is: farm -> mill -> bake -> serve. This requires traveling away from the inn. An innkeeper should be able to buy food supplies at the inn or have employees supply them.
  - Heading to farm to get wheat for bread supply chain.

### Ticks 17-18 (Game Ticks 199-204) [Thalric]
- **BUSY** farming at Oakhaven Farm. Fatigue 42% (was 38%, +4 from travel).

### Tick 19 (Game Tick 210, Day 25 11:35, Rain) -- at Oakhaven Farm
- **Location:** Oakhaven Farm (farm)
- **State:** HP 100, Hunger 68%, Thirst 59%, Fatigue 55%, Social 73%
- **Inventory:** gold x644, fishing net x1, hook x2, wheat x2
- **farm result:** Harvested 2 wheat. Fatigue 42% -> 55% (+13 from farming). farmer skill unlocked at 0.
- **Action:** mill_grain at Stone Mill (+60 min travel)
- **Notes:**
  - **NEW ACTIONS UNLOCKED:** brew_ale, mill_grain, sleep, fire_employee, eat
  - **CONFIRMED:** eat reappeared because we now have wheat (edible food in inventory)
  - **CONFIRMED:** brew_ale available -- "Brew ale from wheat or berries (at inn)" -- interesting, requires inn
  - **CONFIRMED:** mill_grain shows 2 mills: Stone Mill (+60 min) and Cloudmill (+130 min). PR #4 fix working!
  - sleep description: "at home: free; at inn: 3g+1/guest; rough: free but miserable" -- home is The Gilded Inn, +50 min travel
  - fire_employee appeared -- can fire the hired Elara
  - Elara relationship went from +15 to +20 -- relationship may be growing passively?
  - **DESIGN NOTE:** Farm only yielded 2 wheat. Previous playtests got 4 wheat. Random yield?

### Ticks 20-22 (Game Ticks 215-226) [Thalric]
- **BUSY** milling grain. Traveled to Stone Mill and milling wheat.
- Fatigue climbed from 55% to 60% during travel+mill.

### Tick 23 (Game Tick 231, Day 25 13:20, Rain) -- at Stone Mill
- **Location:** Stone Mill (mill)
- **State:** HP 100, Hunger 67%, Thirst 56%, Fatigue 66%, Social 73%
- **mill_grain result:** Milled 2 wheat into 2 flour. Fatigue 55% -> 66% (+11 from travel+milling).
- **Action:** bake_bread_adv
- **Notes:**
  - **CONFIRMED (PR #4):** mill_grain works at Stone Mill (NOT the first mill). Fix verified!
  - **CONFIRMED:** bake_bread_adv available at mill with flour in inventory.
  - go_home appeared (+10 min travel) -- home is The Gilded Inn, relatively close from Stone Mill.
  - sleep available (+10 min travel) -- will route to home (The Gilded Inn).
  - Fatigue at 66% -- need to sleep soon after baking.
  - **OBSERVATION:** The full innkeeper supply chain: farm (15 min travel) -> mill (60 min travel) -> bake -> travel to inn -> serve. This takes 4-5 actions and 100+ game min just to serve one customer.

### Tick 24-25 (Game Ticks 237-242, Day 25 13:55, Rain) -- at Stone Mill
- **Location:** Stone Mill
- **bake_bread_adv result:** 2 flour -> 3 bread. Fatigue 66% -> 73%.
- **Final state:** HP 100, Hunger 66%, Thirst 55%, Fatigue 73%, Social 57%
- **Inventory:** gold x644, fishing net x1, hook x2, bread x3
- **Skills:** innkeeper: 31, cook: 0, barmaid: 0, farmer: 0
- **Notes:**
  - **CONFIRMED:** bake_bread_adv works: 2 flour -> 3 bread. Previous playtest reported same ratio.
  - cook skill unlocked at 0 from baking.
  - Full supply chain tested: farm -> mill -> bake. 3 actions, ~4 hours game time, yields 3 bread.
  - Fatigue now at 73% -- need to sleep.

## END OF TICK LOOP (24 ticks complete, 2 NPC lives)
