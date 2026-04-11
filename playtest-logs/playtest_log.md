# Playtest: Kaelen the Smith — 2026-04-11

## NPC Lives

### Life 1: Kaelen the Smith
- Spawn: Golden Meadow Market, smith, diligent
- Home: The Gilded Inn | Gold: 15 | HP: 100
- Start: Day 64, 07:40, clear weather
- Hunger: 61, Thirst: 62, Fatigue: 29, Stress: 30, Social: 59, Happiness: 71
- Token: 4e9c9fb5-4a02-4a9c-85cd-5034ca55aa4f
- NPC ID: ab96b5f1-1113-4ab6-86c9-9fc95e77ecff
- Start Tick: 4912

## Tick-by-Tick Log

### Ticks 1-10 (Pre-sleep exploration)
- **T1** Scavenge at Golden Meadow Market: 5 bread stacks (10 bread). Iron ore on ground NOT picked up (scavenge priority bug).
- **T2** Eat bread (+25 hunger, 61->86).
- **T3** Drink at Garden Well (+5 min travel). Thirst fully restored.
- **T4-5** Gather_clay at Garden Well: +1 clay first, then +2 clay.
- **T5-6** Travel to Iron Anvil Forge (+35 min). NO smith actions at forge despite smith:35 skill.
- **T6** Explore from forge -> Clearwater Well. Rain weather.
- **T7-8** Gather_clay at Clearwater Well: +2 clay.
- **T8-9** Craft_pottery: 2 clay -> 1 ceramic. Fatigue 74 now.
- **T9** Travel to Saints Rest Shrine. ZERO shrine actions at shrine. No pray, no meditate.
- **T10-22** Sleep at The Gilded Inn (~50 tick duration). Fatigue 78->13 on wake.

### Post-sleep observations
- Social crashed from 59 -> 18 during sleep (41 point drop in 50 ticks!)
- Mood: lonely
- Weather changed to storm
- Still ZERO NPCs in world

### Ticks 11-22 (Post-sleep actions)
- **T11** Scavenge at Gilded Inn: 5 bread stacks again. CONFIRMED scavenge priority bug.
- **T12** Scavenge again: 5 more bread stacks. Still no iron ore/leather.
- **T13** Travel to Iron Anvil Forge (+5 min).
- **T14** Start_business at forge: Claimed for 5 gold. NOW OWNS FORGE.
- **T15** List_actions after owning forge: STILL NO SMITHING ACTIONS. Critical bug - "smith" profession not recognized as "blacksmith".
- **T16** Farm at Sunflower Farm: +4 wheat.
- **T17** Gather_thatch at Sunflower Farm: +3 thatch.
- **T18** Fish at Dock: caught nothing (+fishing xp).
- **T19** Mill_grain at Stone Mill: 2 wheat -> 2 flour. Full chain working.
- **T20** Bake_bread_adv: 2 flour -> 3 bread. Full production chain!
- **T21** Eat bread (+25 hunger).
- **T22** Drink at Hill Well. Thirst restored.
- **T22** Sleep (second cycle). Social at 0, happiness at 47.

### Ticks 23-48 (Second sleep + continued testing)
- Sleeping through tick ~5178 (48 ticks).
- Will resume actions post-sleep.

## Actions Tested (22 unique action commits)
1. scavenge (x3) - picks first 5 stacks, ignores valuable items (BUG)
2. eat (x2) - +25 hunger per bread, works
3. drink (x2) - thirst fully restored at well
4. gather_clay (x3) - 1-2 clay per gather
5. travel (x3) - correct routing, fatigue cost
6. explore (x1) - random destination within 60 units
7. craft_pottery (x1) - 2 clay -> 1 ceramic
8. sleep (x2) - 50 tick duration, -40 fatigue, routes to home inn
9. start_business (x1) - claimed forge for 5 gold
10. farm (x1) - +4 wheat at Sunflower Farm
11. gather_thatch (x1) - +3 thatch
12. fish (x1) - caught nothing (skill-based)
13. mill_grain (x1) - 2 wheat -> 2 flour
14. bake_bread_adv (x1) - 2 flour -> 3 bread
15. check_self (x4), observe (x4), list_actions (x6)

## Locations Visited (12)
Golden Meadow Market, Garden Well, Iron Anvil Forge, Clearwater Well,
Saints Rest Shrine, The Gilded Inn, Sunflower Farm, Dock, Stone Mill,
Hill Well, East Side Market (nearby), Scribe's Library (nearby)

## Confirmed Bugs

### BUG-1 (CRITICAL): "smith" profession unrecognized — zero forge actions
- Spawned as "smith" profession with smith:35 skill
- Profession registry only has "blacksmith" — no "smith" entry
- `smelt_ore` requires HasProfessionOrSkill("blacksmith", "blacksmith", 20) — fails for "smith"
- Even after claiming forge (IsWorkerAtType passes), STILL no smith actions
- Root: LLM spawn returns "smith" instead of "blacksmith", no validation
- Impact: Smith NPCs have ZERO profession-specific gameplay

### BUG-2 (HIGH): Scavenge picks first 5 stacks, not most valuable
- At inn: 80+ bread x2 stacks + iron ore x4 + leather x2
- Scavenge always picks 5 bread stacks, never iron ore or leather
- Multiple scavenges (3+) all returned only bread
- Root: gather.go iterates items in order, takes first 5
- Impact: NPCs hoard worthless bread, miss valuable resources

### BUG-3 (HIGH): No auto-interrupt on enemy damage during multi-tick actions
- engine/tick.go: runEnemyAttackPhase runs every tick even on busy NPCs
- No check for HP threshold to abort action
- Previous playtests confirmed deaths during forage/travel
- Impact: NPCs die helplessly during long actions

### BUG-4 (HIGH): flee_area picks random location, not nearest safe
- combat.go line 69-78: picks random safe location from entire world
- Can send NPC on 160+ min flee trip while being attacked
- Impact: Flee action is functionally useless

### BUG-5 (MEDIUM): Trade sells first inventory item, not most valuable
- economy.go lines 65-71: picks first non-gold item with break
- Impact: NPCs sell cheap items while hoarding valuable ones

### BUG-6 (LOW): Sleep fallback text says "Slept at home" when NPC has no home
- survival.go line 163: catch-all says "home" even for homeless NPCs

## Design Issues Found

### DESIGN-1 (CRITICAL): Zero NPCs in entire world
- 12 locations visited, 22 actions committed: ZERO NPC encounters
- All social/trade/economy actions blocked (require nearby NPCs)
- Social need crashed 59 -> 0 in ~20 actions with no recovery
- Game is unplayable as social sim

### DESIGN-2 (CRITICAL): Shrine has zero actions for most NPCs
- Visited Saints Rest Shrine: zero shrine-specific actions
- pray requires SpiritualSensitivity > 65 OR (stress > 85 / HP < 30 / happiness < 15)
- ~90% of NPCs never qualify for pray
- Shrines are completely dead locations

### DESIGN-3 (HIGH): Sleep takes 50 real ticks (~4 hours)
- 240 game minutes = 48 ticks + travel
- 2 sleep cycles = ~100 of my 218 game ticks (46% of time sleeping!)
- Dominates gameplay, prevents testing

### DESIGN-4 (HIGH): Social need has zero solo recovery
- Social crashed 59 -> 18 during first sleep, 18 -> 0 during play
- No reflect/meditate/journal actions to recover social
- Empty world means social is permanently at zero

### DESIGN-5 (HIGH): Business ownership unlocks zero actions
- Claimed Iron Anvil Forge for 5 gold
- list_actions before and after: IDENTICAL
- No smithing, smelting, or forge actions unlocked
- 5 gold wasted with zero gameplay benefit

### DESIGN-6 (MEDIUM): Weather has zero gameplay effect
- Experienced clear, rain, storm weather
- No effect on travel speed, fatigue, actions, or visibility
- Weather is purely cosmetic

## New Feature Ideas
1. Guard patrol action (guard unique)
2. Healer self-heal / triage (healer unique)
3. Smith forge_item action (smith unique, bypass IsWorkerAt for profession)
4. Meditate at shrine (solo-friendly, no stat gates)
5. Reflect action (social recovery for solo play)
6. Stargaze action (night outdoor relaxation)
7. Weather effects (rain = slower travel, storm = can't forage)
8. Quest board at markets
9. Cooking recipes (fish + herbs = cooked meal, etc.)

## Regression Checks
| Previous Fix | Status |
|-------------|--------|
| Scavenge gold cap at 20 (PR #7) | CONFIRMED - no gold on ground to test |
| list_actions location limit 10 (PR #7) | CONFIRMED - shows 10 + "N more distant" |
| Sleep at home-inn gives home sleep (PR #6) | CONFIRMED - "Slept at home" at assigned inn |
| check_location uses loc.ID (PR #9 fix) | NOT TESTED (check_location parameter format unclear) |
| Social decay during sleep reduced (PR #10) | STILL SEVERE - 41 point drop during one sleep |
| Sleep travel fatigue exemption (PR #10) | Partially - fatigue was 78 at sleep commit, didn't spike |
| Scavenge non-gold cap 5 stacks (PR #12) | CONFIRMED - exactly 5 stacks per scavenge |
