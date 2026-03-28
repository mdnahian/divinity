# Playtest Notes — Eldric (Miller)

**Session:** 2026-03-28
**NPC:** Eldric (Mill worker at The Shivering Hearth)
**Personality:** empathetic, diligent, addictive personality
**HP:** 100

---

## Tick Log

### Ticks 17-20 (Game: Day 20, 21:15-23:00)
- **Tick 17:** Gift bread to Eamon Ironhand (miserable miner). +12 relationship, +3 reputation. Works correctly.
- **Tick 18:** Comforted Elara the Silent (hunter). -15 stress, -3 trauma, +10 relationship. Works correctly.
- **Tick 19:** Claimed Whispering Peak Inn as own business for 5 gold. Now business owner.
- **Tick 20:** Foraging at Whispering Pine Forest. Routing took very long (~47 ticks) despite +5 min travel listed.
- **State at tick 20:** HP 100, Hunger 88, Thirst 86, Fatigue 38, Social 44

### Tick 16 (Sleep completed - Game: Day 20, 20:45)
- **State:** HP 99, Hunger 90, Thirst 89, Fatigue 21, Social 35, Happiness 37, Stress 31
- **Location:** Whispering Peak Inn (home)
- **Memory:** "Slept rough on the ground" despite being at home (Whispering Peak Inn = HomeID)
- **BUG-E3 (HIGH): Sleep at home-inn gives rough sleep.** NPC traveled home to Whispering Peak Inn (HomeID), but sleep treated it as rough sleep because the inn is unowned. needsInn logic on line 108 of survival.go incorrectly includes "at home inn" as needing inn service.
  - Lost: -3 HP, -10 happiness, +10 stress, -10 hygiene
  - Should have gotten: free sleep at home (-40 fatigue, -5 stress)
  - Fatigue decay during travel+sleep brought fatigue from 74 to ~41, then rough sleep only -20 -> 21
  - Net fatigue reduction was ~53 (74->21), mostly from DecayNeeds during the long sleep duration
- **Notes:**
  - Sleep took ~97 ticks total (travel 49 + action 48). Very long action.
  - Fatigue decayed naturally during sleep (good, confirmed fix from previous playtests)
  - HP dropped to 99 from rough sleep penalty
  - Happiness dropped from 51 to 37 (rough sleep -10)
  - Stress increased from 17 to 31 (rough sleep +10)

### Ticks 11-15 (Game: Day 20, 10:15-12:40)
- **Tick 11:** Talked to Alaric (fellow miller). +6 relationship, +20 social. Dialogue stored.
- **Tick 12:** Ate bread (+25 hunger, 70->95).
- **Tick 13:** Explored, wandered to Golden-Root Farm, found old coin.
- **Tick 14:** Brewed healing potion from herbs (herbs x2 -> healing potion).
- **Tick 15:** Sleeping. Routing to home (Whispering Peak Inn, +245 min travel). Fatigue 74->84 during travel.
- **State at tick 15 start:** HP 100, Hunger 93, Thirst 94, Fatigue 74, Social 56
- **DESIGN NOTE:** Fatigue INCREASES during travel to sleep location. Went from 74 to 84 while en route. If fatigue hits 100, NPC collapses. This means NPCs with homes far away may collapse before reaching bed.

### Ticks 8-10 (Game: Day 20, 07:35-08:55)
- **Tick 8:** Drank at Frost-Well of the Sage (+5 min travel from mill). Thirst restored to 100%.
- **Tick 10:** Baking bread from flour (bake_bread_adv). Testing full crafting chain: wheat -> flour -> bread.
- **State at tick 10:** HP 100, Hunger 71, Thirst 100, Fatigue 54, Social 63, Happiness 51, Stress 16
- **Inventory:** bread x2, gold x1197, flour x2, berries x3, herbs x3, saw x3, ruler x2
- **Notes:**
  - Drink fully restores thirst to 100% (confirmed from previous playtests)
  - Fatigue accumulated from travel: 49 -> 54 (travel fatigue from mill to well)
  - go_home still correctly hidden (daytime, fatigue 54 < 60)
  - DESIGN: Travel fatigue adds up significantly. Going to well consumed ~5 fatigue just from walking.

### Tick 7 (Game: Day 20, 07:20, day, rain)
- **State:** HP 100, Hunger 73, Thirst 60, Fatigue 49, Social 66, Happiness 51, Stress 17
- **Location:** Mist-Flow Mill (mill)
- **Memory:** "Milled 2 wheat into 2 flour" -- REGRESSION TEST PASSED (mill at non-first mill)
- **Inventory:** bread x2, gold x30, flour x2
- **Observe:** Nobody else here. Ground items: berries x3, gold x11, herbs x3, **gold x1156**, saw x3, ruler x2
- **Action:** scavenge — pick up 1156+ gold from ground
- **Notes:**
  - REGRESSION VERIFIED: mill_grain works at Mist-Flow Mill (not the first mill). PR #4 fix confirmed.
  - REGRESSION VERIFIED: go_home NOT available during daytime with fatigue 49 (<60). PR #4 fix confirmed.
  - bake_bread_adv appeared because we have flour x2. Good.
  - start_business available (Mist-Flow Mill unowned).
  - **DESIGN BUG (SEVERE): 1156 gold on the ground at Mist-Flow Mill.** This is the same DESIGN-M2 issue but MUCH worse. NPCs dying are dropping massive gold. Scavenging breaks the economy.
  - World event text changed to new spiritual crisis message about sharing bread and building homes.

### Tick 5 (Game: Day 20, 05:30, night, clear)
- **State:** HP 100, Hunger 74, Thirst 62, Fatigue 37, Social 76, Happiness 52, Stress 19
- **Location:** Golden-Root Farm
- **Memory:** "Worked the fields and harvested 2 wheat"
- **Inventory:** bread x2, gold x30, wheat x2
- **Action:** mill_grain at Mist-Flow Mill (+60 min travel) - REGRESSION TEST for PR #4 (mill at non-first mill)
- **Notes:**
  - mill_grain IS available now that we have 2 wheat. Shows 6 mill locations. Mist-Flow Mill is nearest at +60 min.
  - Farm action list appears to show duplicate location entries (docks, wells mixed in) -- need to verify if this is a display bug
  - Golden-Root Farm is "(here)" -- no travel needed for farm. Good.
  - Farmer skill gained: 0.3
  - REGRESSION: go_home still available even though not night (it's 05:30 but technically still night). Need to test during daytime when fatigue < 60 to see if it hides properly.

### Tick 3 (Game: Day 20, 04:40, night, clear)
- **State:** HP 100, Hunger 75, Thirst 63, Fatigue 23, Social 44 (talk dropped it by 20), Happiness 54, Stress 21
- **Location:** The Shivering Hearth (inn)
- **Memory:** Talk with Elara of the Stone worked, +6 relationship (charisma bonus), +20 social
- **Relationship:** Elara of the Stone: slight positive
- **Action:** farm at Golden-Root Farm (+10 min travel) to get wheat for mill_grain testing

### Tick 2 (Game: Day 20, 04:10, night, clear)
- **State:** HP 100, Hunger 76, Thirst 64, Fatigue 23, Social 64, Happiness 51, Stress 21
- **Location:** The Shivering Hearth (inn)
- **Observe:** Same 12 NPCs listed, no items on ground (scavenged)
- **Inventory:** bread x2, gold x30 (scavenge got us 15 more gold)
- **Action:** talk to "Elara of the Stone" (miner)
- **Notes:**
  - BUG CANDIDATE: Tried to talk to "Elara" and "Eldrin" but got "no longer here" errors despite observe listing them. Unique name "Elara of the Stone" worked. Possible FindNPCByName resolves to wrong NPC when multiple share a name.
  - start_business still available (The Shivering Hearth unowned)
  - No ground items after scavenging
  - offer_counsel available (stressed NPCs present)

### Tick 1 (Game: Day 20, 03:40, night, clear)
- **State:** HP 100, Hunger 76, Thirst 65, Fatigue 22, Social 68, Happiness 52, Stress 21
- **Location:** The Shivering Hearth (inn)
- **Observe:** 12 NPCs present (miners, herbalist, tailor, hunter, merchant, shepherd). 15 gold on ground. World event: "Strange omens in the sky" (spiritual crisis).
- **Inventory:** bread x2, gold x15
- **Skills:** mill: 41
- **Home:** Whispering Peak Inn (+230 min travel)
- **Action:** scavenge — pick up 15 gold from ground
- **Notes:**
  - go_home IS available at night (expected). Need to test during day when fatigued.
  - Lots of miserable/exhausted/lonely NPCs — many are miners.
  - Nearby frost region locations (5-60 min travel). Farther regions 155-385 min.
  - gather_clay shows location names properly (BUG-J1 regression: FIXED)
  - No mill_grain action available yet — need to check if I need wheat/grain in inventory or need to be at a mill.
  - Two NPCs named "Eldric" (miners) — potential confusion with our character name.
  - start_business available — could test business ownership.

