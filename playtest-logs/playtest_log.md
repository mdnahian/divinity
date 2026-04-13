# Playtest: Theron/Thalric/Lyra — 2026-04-10

## NPC Lives

### Life 1: Theron the Builder
- Spawn: Golden Meadow Market, builder, neurotic/resilient/addictive
- Home: The Gilded Inn | Gold: 15 | HP: 100
- Start: Day 59, 22:50 (night), clear weather
- DIED at Day 60, 16:45 — killed by wolves at Whispering Forest Edge during forage
- 7 wolf/bear attacks during one forage action, HP 100 → 0
- Ticks: 23 (9 actions submitted + 1 sleep)

### Life 2: Thalric (Farmer)
- Spawn: Golden Meadow Market, farmer, virtuous/neurotic/resilient
- Start: Day 60, 16:55, HP 100, farmer skill 35
- DIED at Day 61, 06:45 — killed by wolves at Whispering Forest Edge while fleeing
- Explore randomly sent to wolf zone, flee_area picked distant location (25 ticks travel)
- Ticks: 21 (12 actions submitted + 1 sleep)

### Life 3: Lyra the Healer
- Spawn: Golden Meadow Market, healer, virtuous/aggressive/outgoing/open-minded
- Start: Day 61, 06:50, HP 100, healer skill 33
- ALIVE at end of playtest at Garden Well
- Ticks: 4

## Actions Tested (48 ticks total across 3 lives)
1. scavenge (x2) — cap 5 stacks working; picks first stacks not most valuable
2. eat (x3) — +25 hunger per bread
3. drink (x3) — thirst fully restored at well
4. travel (x8) — to well, shrine, barracks, library, market, forge, cave, farm, mill
5. explore (x3) — random destination, found coin, carved bone, pretty stone
6. fish (x2) — caught 3 fish; caught nothing (+fishing xp)
7. gather_clay (x4) — 1-2 clay per gather
8. gather_thatch (x1) — 1 thatch
9. farm (x2) — 1-2 wheat
10. mill_grain (x1) — 2 wheat → 2 flour
11. bake_bread_adv (x1) — 2 flour → 3 bread (full chain test!)
12. write_journal (x1) — +3 happiness, -3 stress, +writing skill
13. read_book (x1) — +1 literacy, +happiness
14. start_business (x1) — claimed forge for 5g, NO new actions unlocked
15. sleep (x2) — full fatigue reset, 50 ticks each
16. go_home — tested via sleep destination
17. forage (x1) — DIED during action
18. flee_area (x1) — too slow, DIED during flee
19. check_self (x4), observe (x6), list_actions (x5), recall_memories (x2)

## Locations Visited (17)
Golden Meadow Market, Garden Well, Holy Stone Shrine, Dock, Royal Barracks,
The Gilded Inn, Scribe's Library, Saints Rest Shrine, East Side Market,
Iron Anvil Forge, Shadow Cave, Silvermine Outpost, Sunflower Farm, Stone Mill,
Hill Well, Clearwater Well, Rose Garden, Whispering Forest Edge (death zone)

## Confirmed Bugs

### BUG-1 (CRITICAL): NPCs die during multi-tick non-combat actions — no interrupt
- Wolves attack every tick during forage/travel/gather, NPC cannot react (busy-locked)
- Life 1: 7 attacks during forage killed Theron (HP 100→0)
- Life 2: died during flee_area because travel was 25 ticks long
- FIX IMPLEMENTED: Auto-abort non-combat actions when HP < 50 from enemy damage

### BUG-2 (HIGH): flee_area picks random world location — takes 160+ min
- Should pick NEAREST safe location, not random
- FIX IMPLEMENTED: flee_area now picks nearest enemy-free location

### BUG-3 (MEDIUM): Trade always sells first inventory item
- NPC sells bread (cheapest) while hoarding valuable items
- FIX IMPLEMENTED: trade now picks most valuable item the buyer can afford

### BUG-4 (MEDIUM): Scavenge picks first 5 stacks, not most valuable
- Breadx2 picked before iron ore, ceramic, leather stacks
- FIX IMPLEMENTED: sort ground items by value before scavenging

## Design Issues Found

### DESIGN-1 (SEVERE): Entire territory devoid of NPCs
- 17 locations visited across 48 ticks: ZERO NPCs encountered anywhere
- All social/trade/party actions untestable
- Makes the game a solitary survival sim

### DESIGN-2 (HIGH): Forest death trap near spawn
- Whispering Forest Edge: wolves + bear attack EVERY tick
- Forage/explore randomly send NPCs to this death zone
- NPCs die in 7 ticks with no counterplay

### DESIGN-3 (HIGH): Builder profession has zero unique actions
- Builder with skill 30 has no build/repair/construct actions available
- begin_construction requires materials + carpentry nobody can get solo
- FIX IMPLEMENTED: Added `reinforce_structure` action for builders

### DESIGN-4 (HIGH): Shrines have no solo-friendly actions
- `pray` requires stress > 40 or HP < 70 — rarely available
- Shrines sit empty with no meaningful interaction
- FIX IMPLEMENTED: Added `meditate` and `tend_shrine` actions

### DESIGN-5 (HIGH): Business ownership unlocks zero new actions
- Claimed Iron Anvil Forge — still no forge/smelt/smith actions
- Business ownership is cosmetic

### DESIGN-6 (MEDIUM): Sleep takes 50 ticks real-time
- 240 min sleep + 40+ min travel = 56 ticks per sleep cycle
- Dominates gameplay time

### DESIGN-7 (LOW): Barracks has zero guard-specific actions
- No train, patrol, drill, or guard actions

## Improvement Ideas
- Shorter sleep (120-150 min instead of 240)
- Explore should avoid enemy-occupied locations
- Weather should affect gameplay (storm = slower travel, cold = hunger drain)
- Social need should decay slower when world is empty
- Healer profession should have heal_self action

## New Feature Ideas
- Quest board at markets/barracks
- Reputation-gated actions (high rep = discounts, access)
- Day/night cycle affecting enemy behavior
- Seasonal resource regeneration
- NPC migration events (caravan arrivals)

## New Features Implemented

### 1. meditate (wellbeing)
Solo-friendly shrine action available to all NPCs. Reduces stress (-12),
boosts happiness (+4), reduces fatigue (-3), slowly increases spiritual
sensitivity. Not gated on desperation — usable anytime at any shrine.

### 2. tend_shrine (wellbeing)
Shrine maintenance action. Gains reputation (+2), restores shrine building
durability if damaged, optionally consumes a curiosity item as an offering
for bonus spiritual/reputation gains.

### 3. stargaze (wellbeing)
Night-only outdoor action. Free, solo-friendly. Reduces stress (-8),
boosts happiness (+6). 12% chance of seeing a shooting star (+1 wisdom).
Available at outdoor locations (markets, farms, shrines, etc).

### 4. whittle (craft)
Solo-friendly crafting: consumes 1 log, produces carved figurine (new
curiosity item). Gains carpentry skill. Builders with carpentry >= 50
have 40% chance to produce 2 figurines. Available anywhere.

### 5. reinforce_structure (construction)
Builder-specific action: reinforces any building at current location
(not just owned). Gains carpentry skill, +1 reputation for public
buildings. Optionally consumes 1 stone for bonus durability.

### 6. carved figurine (item)
New curiosity item. Produced by whittle action. Can be left as shrine
offering via tend_shrine. Tradeable at market.

### Bug Fixes Implemented
- Auto-interrupt non-combat actions when HP < 50 from enemy attacks
- flee_area picks nearest safe location (not random)
- Trade sells most valuable item buyer can afford (not first slot)
- Scavenge sorts items by value before picking top 5 stacks

## Regression Checks

| Previous Fix | Status |
|-------------|--------|
| Scavenge gold cap at 20 (PR #7) | CONFIRMED working |
| list_actions location limit 10 (PR #7) | CONFIRMED — shows 10 + "N more distant" |
| Sleep at home-inn gives home sleep (PR #6) | CONFIRMED — "Slept at home (-40 fatigue, -5 stress)" |
| check_location uses loc.ID (PR #9 fix) | CONFIRMED FIXED in main |
| Social decay during sleep reduced (PR #10 fix) | CONFIRMED — 25% rate during sleep |
| Sleep travel fatigue exemption (PR #10 fix) | CONFIRMED — sleep/go_home skip travel fatigue |
| Scavenge non-gold cap 5 stacks (PR #12) | CONFIRMED working |
