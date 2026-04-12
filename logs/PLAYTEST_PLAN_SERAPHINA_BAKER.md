# Playtest Plan: Seraphina the Baker / Elian the Builder — 2026-04-12

## Phase 0.3: PR Review
- PRs #13-15 OPEN (not merged). #12-#1 include merged/closed history.
- No comments or reviews on any open PRs.
- PRs #13-15 cover: meditate, reflect, stargaze, tend_shrine, whittle, reinforce_structure, guard_patrol, healer_triage, profession aliases, auto-interrupt, smart flee, trade value fix, scavenge sort, sleep text fix, smith forge access.

## Phase 1: Spawn and Initial State
### Life 1: Elian the Builder (DEAD)
- Spawn: Golden Meadow Market, builder, diligent/generous/open-minded/jealous/timid
- Start tick: 6021 | Builder skill: 47 | Gold: 15
- DIED at tick 6045 at Oakhaven Farm — killed by dire wolves (100 -> 0 HP over 13 ticks)
- Total damage: 18 + 8 + 19 + 8 + 19 + 23 + 24 = ~119 damage

### Life 2: Seraphina the Baker (ALIVE)
- Spawn: Golden Meadow Market, baker, ambitious/neurotic/generous
- Start tick: 6045 | Baker skill: 46 | Gold: 15
- Currently: Sleeping at inn

## Phase 2: 48-Tick Playtest Log

### Ticks 1-6 (Elian): Scavenge, Eat, Travel to Shrine, Farm, DEATH
- Tick 1 (6021-6024): Scavenge at market -> 5x bread stacks (BUG: picks bread over iron ore)
- Tick 2 (6024-6027): Eat bread -> +25 hunger
- Tick 3 (6027-6033): Travel to Holy Stone Shrine -> NO shrine actions available (confirmed bug)
- Tick 4 (6033-6041): Farm at Oakhaven -> 4 wheat, BUT took 48 damage from dire wolves
- Tick 5 (6041-6043): Flee area -> INTERRUPTED, picked location 33 ticks away (BUG: random flee)
- Tick 6 (6043-6045): Travel to market -> DIED during travel from continued wolf attacks

### Ticks 7-16 (Seraphina): Drink, Fish, Clay, Pottery, Farm, Explore, Business, Sleep
- Tick 7 (6045-6049): Drink at Garden Well -> thirst fully restored
- Tick 8 (6049-6058): Fish at Dock -> caught 2 fish
- Tick 9 (6058-6065): Gather clay at Dock -> 2 clay
- Tick 10 (6065-6074): Craft pottery -> 1 ceramic from 2 clay
- Tick 11 (6074-6082): Farm at Sunflower Farm -> 2 wheat (no wolves here!)
- Tick 12 (6082-6085): Scavenge at Sunflower -> bread x27, iron ore x4, bread x33, gold x10, clay x1
- Tick 13 (6085-6091): Explore -> Rose Garden (no finds)
- Tick 14 (6091-6096): Travel to Golden Meadow Market
- Tick 15 (6096-6105): Start business -> Claimed market for 5 gold (NO new actions unlocked)
- Tick 16 (6105-6161): Sleep at inn -> 56 ticks!!

## Bugs Confirmed (on main)
1. **Scavenge picks bread over iron ore** — Known, PRs #13-15 fix it
2. **Shrine has zero actions** — Known, PRs #13-15 add meditate
3. **flee_area random location** — Picked 33-tick distant location instead of nearest safe
4. **No auto-interrupt during farming** — Took 48 HP damage while farming (8 ticks)
5. **Sleep text "Slept at home"** — Known, PRs #13-15 fix it
6. **Business ownership no benefit** — Claimed market, zero new actions
7. **Baker has no profession-specific actions** — No bake action available without flour

## Design Issues Found
1. **SEVERE**: Zero NPCs in 20+ locations across 2 lives
2. **HIGH**: Oakhaven Farm death trap (dire wolves at closest farm to spawn)
3. **HIGH**: Sleep takes 56 ticks (46% of total gameplay time)
4. **HIGH**: Baker profession has zero unique actions without flour
5. **MEDIUM**: Weather has no gameplay effect (rain during farm/fish)
6. **MEDIUM**: No way to cook fish/meat without being inn worker

## Phase 4: New Features Plan (DIFFERENT from PRs #13-15)

### Feature 1: Campfire System
- New `build_campfire` action: consumes 1 log at any outdoor location
- New `campfire_cook` action: cook fish/meat into cooked meal at campfire/fire
- New `warm_by_fire` action: reduce stress, fatigue at campfire location
- Makes solo play more viable, gives non-inn-workers access to cooking

### Feature 2: Herb Gathering
- New `gather_herbs` action at forest/garden locations
- Direct herb gathering instead of relying on 15% random forage chance
- Enables healing potion brewing chain for solo players

### Feature 3: Salvage/Breakdown
- New `salvage` action: break down equipment into raw materials
- Iron sword -> iron ingot, leather armor -> leather, etc.
- Gives resource recycling for damaged equipment

### Feature 4: Weather Effects on Actions
- Rain: +1 bonus wheat from farming, -10% travel speed
- Cloudy: no effect
- Clear: +happiness from outdoor actions

### Feature 5: Baker Profession Actions
- Baker gets `bake_bread_solo`: bake bread from wheat directly (no flour needed)
- Baker gets access to cook at any fire source, not just inn
