# Playtest: Elara/Thomas/Mara — 2026-04-01

## NPC Lives
### Life 1: Elara the Builder
- Spawn: Golden Meadow Market, builder, brave/generous/cold/addictive
- Home: The Gilded Inn | Gold: 15 | Builder skill: 42
- DIED at tick 2688 (Day 38, 11:40) — killed by dire wolves while traveling from Oakhaven Farm
- Total damage from wolves: 104 HP (8+10+13+21+18+7+7+20)
- Cause: Social hit 0%, stress 64%, wolves attacking during farm and travel

### Life 2: Thomas the Farmer
- Spawn: Golden Meadow Market, farmer, mentor/neurotic/resilient/generous
- DIED at tick 2710 (Day 38, 13:30) — killed at Whispering Forest Edge
- 9 enemies at forest (4 wolves, 4 dire wolves, 1 bear)
- HP 100 -> 61 -> 0 within 2 ticks

### Life 3: Mara the Farmer
- Spawn: Golden Meadow Market, farmer, outgoing/creative/wise/resilient
- Home: The Gilded Inn | Farmer skill: 35
- Status: ALIVE at end of playtest (sleeping at home)
- Owned business: Silvermine Outpost (claimed for 5 gold)

## Actions Tested (27 total)
1. scavenge (x2) — picked up items, gold cap 20 working
2. eat (x2) — solo eat from inventory works, +25 hunger per bread
3. travel (x6) — to shrine, barracks, inn, market, mine, stable
4. go_home (x1) — traveled home but didn't include sleep
5. explore (x3) — random destination, found pretty stone once
6. drink (x2) — thirst fully restored at well
7. gather_clay (x1) — got 1 clay
8. gather_thatch (x1) — got 1 thatch
9. sleep (x2) — at home inn, -40 fatigue, -5 stress (correct)
10. forage (x1) — died during action
11. farm (x2) — harvested 3-4 wheat
12. fish (x1) — caught 1 fish
13. start_business (x1) — claimed mine for 5 gold
14. check_self (x8), observe (x8), list_actions (x6), recall_memories (x4)

## Locations Visited (12+)
Golden Meadow Market, Holy Stone Shrine, Royal Barracks, The Gilded Inn, East Side Market,
West Gate Inn, Old Well, Rose Garden, Garden Well, Crystal Well, Oakhaven Farm, Sunflower Farm,
Silvermine Outpost, Dock, Whispering Forest Edge, Sunrise Stable

## Bugs Found

### BUG-B1 (HIGH): check_location uses raw input string instead of resolved loc.ID
- File: server/gametools/npc_tools.go lines 452, 461, 466
- NPCsAtLocation, EnemiesAtLocation, and TravelTicks all use params.LocationID (user input)
  instead of loc.ID (resolved location ID)
- Causes: NPC list empty, enemy count wrong, travel time shows ~0 min
- Same bug identified in PR #9 (unmerged)
- STATUS: Still present, will fix

### BUG-B2 (HIGH): Social need decays at full rate during sleep
- File: server/npc/npc.go line 667
- DecayNeeds only reduces hunger/thirst/fatigue rates during sleep, but social need
  (line 667) decays at the full rate regardless of sleep state
- Observed: Social dropped from 22% to 2% during one long sleep session
- Same bug identified in PR #10 (unmerged)
- STATUS: Will fix

### BUG-B3 (MEDIUM): Sleep travel fatigue spike
- File: server/engine/turns.go lines 330-344
- Travel fatigue is added IMMEDIATELY on action submission, even for sleep
- NPC with 67% fatigue spikes higher when walking to bed
- Same bug identified in PR #10 (unmerged)
- STATUS: Will fix

## Design Issues Found

### DESIGN-1 (SEVERE): Entire territory devoid of NPCs
- Visited 12+ locations across Golden Meadow territory — zero NPCs found at any location
- Markets, inns, barracks, farms, mines, shrines — all empty
- Makes social interaction testing impossible
- Worse than "NPC concentration at inns" — there are NO NPCs at all

### DESIGN-2 (HIGH): Forest death trap near spawn
- Whispering Forest Edge (5 min from market): 9 enemies (4 wolves, 4 dire wolves, 1 bear)
- Oakhaven Farm (10 min from market): 2 dire wolves
- New NPCs die within 2-3 ticks of entering these areas
- Wolves attack during non-combat actions (farming, foraging, traveling) with no escape

### DESIGN-3 (HIGH): Builder profession has zero unique actions
- Builder with skill 42 has no build, repair, construct actions
- Same generic actions as any other profession at barracks, forge, market, etc.

### DESIGN-4 (MEDIUM): Shrine has no actions
- Holy Stone Shrine: only generic actions (eat, drink, travel, explore)
- No pray, meditate, or worship action available

### DESIGN-5 (MEDIUM): Business ownership has no gameplay benefit
- Claimed Silvermine Outpost as business (5 gold)
- No new actions unlocked (no mine_ore, manage, hire_employee, etc.)

### DESIGN-6 (LOW): Explore shows misleading travel time
- Travel time for explore is calculated from a random destination (changes each call)
- Shows "+55 min travel" or "+40 min travel" but actual explore goes to random nearby loc

### DESIGN-7 (LOW): Duplicate location names confuse players
- "Dock" appears twice in fish destinations (+5 min and +110 min)
- "Hill Well" appears twice in drink destinations (+35 min and +100 min)
- No way to distinguish which one you're selecting

### DESIGN-8 (LOW): Ground items (bread) accumulate massively
- Markets and inns littered with dozens of bread x2 stacks
- One scavenge picks up 100+ bread — no cap on non-gold items

## Regression Checks

| Previous Fix | Status |
|-------------|--------|
| Scavenge gold cap at 20 (PR #7) | CONFIRMED — picked up exactly 20 gold from larger stacks |
| list_actions location limit 10 (PR #7) | CONFIRMED — shows 10 + "N more distant" |
| Ground gold decay (PR #7) | CONFIRMED — gold stacks reduced between visits |
| Sleep at home-inn gives home sleep (PR #6) | CONFIRMED — "Slept at home (-40 fatigue, -5 stress)" |
| FindNPCByNameAtLocation (PR #6) | UNTESTABLE — no NPCs found to interact with |
| check_location 0 min travel (PR #9 unmerged) | CONFIRMED STILL PRESENT |
| Sleep travel fatigue (PR #10 unmerged) | CONFIRMED STILL PRESENT |
| Social decay during sleep (PR #10 unmerged) | CONFIRMED STILL PRESENT |
