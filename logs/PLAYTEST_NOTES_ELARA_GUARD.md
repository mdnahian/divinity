# Playtest Notes — Elara (Guard)

**Session:** 2026-03-30
**NPC:** Elara
**Profession:** Guard
**Location:** Mistwell Village
**Personality:** virtuous, generous
**NPC ID:** 5c214915-9b1c-49e6-9216-2ac2cea2e53b

---

## Tick Log

### Tick 1 (Game Tick 114) — Scavenge Test
- **State:** HP 100, Hunger 78, Thirst 62, Fatigue 22, Social 66, Happiness 68, Stress 22
- **Location:** Mistwell Village (market)
- **Observed:** Nobody here. 34 gold on ground. Weather clear, Day 29 13:10.
- **Available Actions:** eat, drink, forage, hunt, farm, fish, gather_thatch, gather_clay, scavenge, explore, travel, start_business
- **NOTE:** No guard-specific actions (patrol, defend, etc.) visible despite guard profession!
- **NOTE:** list_actions location limit working — shows "... and N more distant locations" (PR #7 fix confirmed)
- **Action:** scavenge — Testing gold cap. 34 gold on ground, expecting cap at 20.
- **Inventory before:** bread x2, gold x15

### Tick 2 (Game Tick 119) — Eat Test
- **State:** HP 100, Hunger 78, Thirst 61, Fatigue 23, Social 63, gold 35
- **Scavenge Result:** Picked up 20 gold (cap at 20 working!). 14 gold remains on ground.
- **REGRESSION CHECK:** PR #7 scavenge gold cap CONFIRMED WORKING
- **REGRESSION CHECK:** PR #7 list_actions location limit CONFIRMED WORKING (shows 10 + "... and N more")
- **Weather:** Changed from clear to storm
- **NOTE:** No trade action available despite being at market - possibly need specific trade goods?
- **NOTE:** Still no guard-specific actions visible
- **Action:** eat — Testing eat action with bread

### Tick 3 (Game Tick 125) — Travel to Castle
- **State:** HP 100, Hunger 99.9 (eat worked +25!), Thirst 61, Fatigue 23
- **Eat Result:** bread +25 hunger (78 -> 99.9). Consumed 1 bread. CONFIRMED WORKING.
- **NOTE:** check_location shows "~0 min" travel time for all locations - BUG? Should show travel time from current position.
- **Action:** travel to Peakwatcher Citadel (+10 min from list_actions) - checking for guard-specific actions

### Tick 4 (Game Tick 130) — Scavenge at Castle
- **State:** HP 100, Hunger 99.5, Thirst 60, Fatigue 27 (+4 from travel)
- **Location:** Peakwatcher Citadel (castle)
- **Observed:** Nobody here. Ground has 185 gold total (63+15+12+59+14+10+12) plus berries x1, herbs x1. Storm weather.
- **BUG/DESIGN: Guard profession has NO special actions** — no patrol, defend, guard_gate despite guard skill 47 at a castle.
- **BUG/DESIGN: Gold accumulation at castle** — 185 gold on ground despite PR #7 gold decay fix. Castle may have high NPC death rate or insufficient decay.
- **NOTE:** check_location bug confirmed — shows "~0 min" travel for all locations regardless of actual distance.
- **Action:** scavenge — Testing gold cap with multiple gold piles

### Tick 5 (Game Tick 136) — Travel to Market
- **State:** HP 100, Hunger 99, Thirst 59, Fatigue 28, Social 54, gold 55
- **Scavenge Result:** 20 gold again (cap confirmed). Also got berries x1, herbs x1 (non-gold items not capped).
- **Ground gold at castle:** 165 remaining (185 - 20). Multiple piles handled correctly.
- **Social dropping:** 66 -> 63 -> 54 over 5 ticks. Need NPCs urgently.
- **Action:** travel to Aethelgard Market (+10 min) — looking for NPCs

### Tick 6 (Game Tick 141) — Explore
- **State:** HP 100, Hunger 99, Thirst 59, Fatigue 32, Social 54 (still dropping)
- **Location:** Aethelgard Market — still nobody here! 20 gold on ground.
- **NOTE:** Both markets visited (Mistwell Village, Aethelgard Market) are completely empty of NPCs.
- **NOTE:** No trade action available at this market either.
- **Action:** explore — Trying to discover new areas and find populated locations

**6-tick checkpoint update:** Key findings so far:
1. Scavenge gold cap WORKING (PR #7 confirmed)
2. list_actions location limit WORKING (PR #7 confirmed)
3. Guard profession has NO special actions (design issue)
4. check_location shows 0 min travel for all locations (bug)
5. Markets are deserted - no NPCs found in 6 ticks
6. Gold accumulation still significant at castle (185 gold)

### Tick 7 (Game Tick 147) — Busy (exploring)
- **State:** Busy until tick 152. Still at Aethelgard Market during explore.
- **Stats:** HP 100, Hunger 98, Thirst 58, Fatigue 35

### Tick 8 (Game Tick 152) — Drink Water
- **State:** HP 100, Hunger 98, Thirst 57, Fatigue 40, Social 54 (likely lower)
- **Location:** Emberforge (forge) — nobody here, gold x23 + fish on ground
- **Explore result:** Wandered from Aethelgard Market to Emberforge. +8 fatigue from explore.
- **NOTE:** Still no NPCs found anywhere! 4 locations visited, all empty.
- **Action:** drink at Ridge Top Well (+5 min travel)

### Tick 9 (Game Tick 158) — Gather Clay
- **State:** HP 100, Hunger 98, Thirst 99.8 (drink worked!), Fatigue 41
- **Location:** Ridge Top Well — resources: water 149, clay 8. Still nobody here.
- **Drink Result:** Thirst fully restored to 99.8%. CONFIRMED WORKING.
- **NOTE:** "(here)" label correctly shown for drink and gather_clay at Ridge Top Well
- **NOTE:** 5 locations visited, ALL empty of NPCs. Potential population density issue.
- **Action:** gather_clay at Ridge Top Well (here) — no travel needed

### Tick 10 (Game Tick 163) — Busy (gather_clay in progress)
- Busy until tick 164

### Tick 11 (Game Tick 168) — Farm
- **State:** HP 100, Hunger 97, Thirst 98, Fatigue 52, Social 38 (!), Happiness 59
- **gather_clay Result:** 2 clay gathered, +11 fatigue. Resource went from 8 to 7 (possible regen issue or display).
- **Social crisis:** 66 -> 38 in 11 ticks with zero NPC interaction. Declining ~2.5/tick.
- **World events:** Dire wolf at Shelter Caves, strange omens, divine protection at farms/caves/wells.
- **check_location BUG:** Always shows "~0 min" travel time regardless of actual distance. Also doesn't show NPCs at locations.
- **Weather:** storm -> rain
- **Inventory:** bread x1, gold x55, berries x1, herbs x1, clay x2
- **Action:** farm at Obsidian Farm (+15 min travel) — testing farming

### Tick 12 (Game Tick 174) — Busy (farming)
- Busy until tick 178

### Tick 13 (Game Tick 180) — Craft Pottery
- **State:** HP 100, Hunger 96, Thirst 97, Fatigue 66 (\!), Social 38 (est)
- **Location:** Obsidian Farm — nobody here. Ground: gold x10, fish x1, iron ore x4
- **Farm Result:** 4 wheat harvested, +14 fatigue (including travel). CONFIRMED WORKING.
- **New actions unlocked:** sleep (fatigue >= 60), go_home (fatigue >= 60), mill_grain (have wheat), craft_pottery (have clay)
- **REGRESSION CHECK:** go_home available during day with fatigue >= 60 — PR #4 fix CONFIRMED
- **NOTE:** go_home shows +240 min travel — home is very far\! Sleep travel fatigue collapse risk (DESIGN-E1).
- **NOTE:** sleep shows +240 min travel to home — would add massive fatigue during transit.
- **Action:** craft_pottery — testing new craft action with 2 clay

TICK12-13 < /dev/null
### Tick 12 (Game Tick 174) — Busy (farming)
- Busy until tick 178

### Tick 13 (Game Tick 180) — Craft Pottery
- **State:** HP 100, Hunger 96, Thirst 97, Fatigue 66, Social 38 (est)
- **Location:** Obsidian Farm - nobody here. Ground: gold x10, fish x1, iron ore x4
- **Farm Result:** 4 wheat harvested, +14 fatigue (including travel). CONFIRMED WORKING.
- **New actions unlocked:** sleep, go_home, mill_grain, craft_pottery
- **REGRESSION CHECK:** go_home available during day with fatigue >= 60 - PR #4 fix CONFIRMED
- **NOTE:** go_home shows +240 min travel - home is very far. Sleep travel fatigue collapse risk.
- **Action:** craft_pottery - testing new craft action with 2 clay

### Tick 14 (Game Tick 185) — Busy (craft_pottery)
- Busy until tick 189

### Tick 15 (Game Tick 191) — Sleep
- **State:** HP 100, Hunger 95, Thirst 96, Fatigue 75, Social 27 (!!)
- **craft_pottery Result:** 2 clay -> 1 ceramic. +9 fatigue. Farmer skill unlocked.
- **Social CRISIS:** 27% and dropping. 15 ticks, zero NPC interactions across 7 locations.
- **Inventory:** bread x1, gold x55, berries x1, herbs x1, wheat x4, ceramic x1
- **Action:** sleep - Testing sleep with home 240 min away. This tests DESIGN-E1 sleep travel fatigue.

### Tick 16 (Game Tick 196) — Busy (sleeping/traveling)
- **Fatigue rose to 87.2 from 75** during sleep! +12 fatigue while traveling to home for sleep.
- **BUG CONFIRMED: DESIGN-E1 Sleep travel fatigue** - fatigue increases during travel to sleep location.
- Busy until tick 287 (91 more ticks = very long sleep action)
- If fatigue had been 88+, the NPC would collapse during transit to bed!

### Ticks 16-Sleep (Game Ticks 196-287) — Sleep at Home
- **Sleep journey:** Fatigue went from 75 -> 87.2 (peak during travel) -> 1.7 (after sleep)
- **Travel fatigue during sleep:** +12.2 fatigue added during transit to home! CONFIRMS DESIGN-E1.
- **Slept at home (Rat's Nest inn):** Free home sleep, -40 fatigue, -5 stress. NOT rough sleep.
- **REGRESSION CHECK:** PR #6 inn-home sleep fix - NEED MORE DATA. This was "Slept at home" not rough.
- **Memory insights at midnight:** Generated reflections on loneliness and crafting.
- **Total sleep duration:** ~91 ticks (tick 196-287). Extremely long due to 240 min travel.

### Tick 17 (Game Tick 287) — Talk to Eldrin
- **Location:** Rat's Nest (inn) - our HOME!
- **FINALLY FOUND NPCs!** 6 NPCs present.
- **DUPLICATE NAME BUG EXTREME:** 5 other "Elara" NPCs at same location (potter, 2 priests, herbalist, scribe)! Plus our guard = 6 Elaras!
- **State:** HP 100, Hunger 92, Thirst 90, Fatigue 2 (great!), Social 0 (!), Happiness 43
- **Mood:** lonely (Social at 0%)
- **Gold on ground:** 564 gold at this inn! Massive accumulation despite PR #7 decay fix.
- **inspect_person Eldrin:** Works correctly, shows farmer, mood lonely, relationship 0.
- **Guard Post nearby:** +10 min travel - might unlock guard actions
- **Action:** talk to Eldrin - desperately need social recovery

### Tick 18 (Game Tick 294) — Talk to "Elara" (Duplicate Name Test)
- **Talk result:** Social +20, relationship +3 with Eldrin. Talk works correctly.
- **inspect_person "Elara" BUG:** Returned OUR OWN NPC (guard) instead of another Elara at location!
- **BUG: inspect_person resolves self as target** when multiple NPCs share a name, FindNPCByName returns the requesting NPC itself.
- **DUPLICATE NAME EXTREME:** 6 NPCs named "Elara" at same location (guard=us, potter, 2 priests, herbalist, scribe)
- **Action:** talk to "Elara" - testing if talk also targets self or finds another Elara

### Tick 19 (Game Tick 299) — Scavenge (Gold Cap Test at 564 Gold)
- **Talk to Elara result:** Social +20, relationship +3. But WHICH Elara? System tracks as single entity.
- **Memory echo:** Memory shows "Elara said..." and "heard from Elara" - confusing self-referential loop.
- **DESIGN: Relationship tracking merges all same-named NPCs** into one entry ("Elara: slight positive")
- **Action:** scavenge - Testing gold cap with 564 gold on ground

### Tick 20 (Game Tick 305) — Travel to Guard Post
- **Scavenge result:** 20 gold (cap), 1 quill, 3 ink pot. Gold cap CONFIRMED again at massive pile.
- **Inventory now:** bread x1, gold x75, berries x1, herbs x1, wheat x4, ceramic x1, quill x1, ink pot x3
- **Action:** travel to Guard Post (+10 min) - testing guard-specific actions

### Tick 21 (Game Tick 310) — Guard Post Result + Travel to Shrine
- **Location:** Guard Post (barracks) - nobody here, water 40, wheat x3 + iron ore x1 on ground
- **CONFIRMED BUG: Guard profession has NO unique actions even at barracks.** No patrol, defend, guard_gate.
- Guard skill 47 provides zero gameplay benefit. Major design gap.
- **Action:** travel to Silent Shrine (+5 min) - testing shrine actions

### Tick 22 (Game Tick 316) — Travel to Inn
- **Location:** Silent Shrine - no unique actions (no pray, worship, meditate)
- **DESIGN: Shrines have no unique actions** - purely decorative locations
- **eat reappeared** in action list (hunger dropped to 90)
- **Action:** travel to Drowned Lantern Inn (+15 min) - looking for NPCs and inn mechanics

### Tick 23 (Game Tick 321) — Busy (traveling)
- Busy until tick 322

### Tick 24 (Game Tick 326) — Final Tick: Work Together at Inn
- **Location:** Drowned Lantern Inn - 8 NPCs present!
- **NPCs:** Kaelen (hunter), Elian the Silent (scribe), Eldric x2 (carpenter/baker), Silas (fisher), Elara x3 (guard/herbalist/fisher)
- **State:** HP 100, Hunger 89, Thirst 86, Fatigue 16, Social 32, Happiness 44
- **MORE DUPLICATE NAMES:** 3 Elaras + 2 Eldrics at this inn
- **ANOTHER ELARA GUARD:** Duplicate of our exact name AND profession!
- **Gold on ground:** 698 gold! Even worse accumulation at this inn.
- **New action: work_together** - appears when same-profession NPC is at location
- **World events:** Golden light marketplace omens, dire wolf at Whispering Reeds
- **Action:** work_together with Elara (guard) - testing new action

## End of Tick Loop (24 ticks played)


### Post-Loop: work_together Result
- **Result:** "Worked together with Elara - improved guard skills and bond." +20 fatigue.
- **Relationship:** work_together tracked in relationships
- **NOTE:** Memory duplicates the event ("Worked together" appears twice with slightly different text)

---

## Session Summary

### Actions Tested
1. observe - WORKING
2. check_self - WORKING
3. inspect_person - WORKING (but name resolution issue)
4. list_actions - WORKING (PR #7 location limit confirmed)
5. check_location - BUGGY (0 min travel, no NPC info)
6. recall_memories - NOT TESTED
7. scavenge - WORKING (gold cap at 20 confirmed)
8. eat - WORKING (+25 hunger from bread)
9. drink - WORKING (thirst fully restored)
10. travel - WORKING
11. explore - WORKING (moved to new location)
12. farm - WORKING (+4 wheat)
13. gather_clay - WORKING (+2 clay)
14. craft_pottery - WORKING (2 clay -> 1 ceramic)
15. sleep - WORKING (at home, free, -40 fatigue)
16. talk - WORKING (+20 social, +3 relationship)
17. work_together - WORKING (improved skills + bond)

### Actions NOT Tested (unavailable or no opportunity)
- hunt, fish, chop_wood, mine_stone, mine_ore, forage
- trade, gift, comfort
- start_business, serve_customer, hire_employee
- mill_grain, bake_bread, brew_ale, forge_weapon, etc.
- Any guard-specific actions (NONE EXIST)

### Locations Visited (10 total)
1. Mistwell Village (market) - empty
2. Peakwatcher Citadel (castle) - empty, 185 gold on ground
3. Aethelgard Market (market) - empty
4. Emberforge (forge) - empty
5. Ridge Top Well (well) - empty
6. Obsidian Farm (farm) - empty
7. Rat's Nest (inn/home) - 6 NPCs! All Elaras except Eldrin
8. Guard Post (barracks) - empty
9. Silent Shrine (shrine) - empty
10. Drowned Lantern Inn (inn) - 8 NPCs

