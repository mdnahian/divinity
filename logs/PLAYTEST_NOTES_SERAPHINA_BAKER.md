# Playtest Notes: Seraphina the Baker / Elian the Builder — 2026-04-12

## NPC Lives
### Life 1: Elian the Builder (DIED tick 6045)
- Spawn: Golden Meadow Market, builder, diligent/generous/open-minded/jealous/timid
- Builder skill: 47 | Gold: 15
- DIED at Oakhaven Farm — killed by dire wolves while farming then trying to flee
- Total wolf damage: ~119 HP over 13 ticks
- Cause: No auto-interrupt during farming, flee_area picked distant random location

### Life 2: Seraphina the Baker (ALIVE, HP 26)
- Spawn: Golden Meadow Market, baker, ambitious/neurotic/generous
- Baker skill: 46 | Gold: 20
- Owned business: Golden Meadow Market (5 gold)
- Home: The Gilded Inn
- Status: Alive but critical HP 26, being attacked by wolves at Oakhaven Farm
- Fled from wolves but flee picked a 35-tick-away location and was interrupted

## Actions Committed (48 total)
### Life 1 (6 ticks):
1. scavenge - picked up 5x bread stacks (BUG: ignored iron ore)
2. eat - +25 hunger
3. travel - to Holy Stone Shrine (0 shrine actions, BUG confirmed)
4. farm - at Oakhaven (4 wheat, took 48 damage from wolves)
5. flee_area - INTERRUPTED (picked 33-tick distant location)
6. travel - DIED during travel from wolf attacks

### Life 2 (42 ticks):
7-10. drink, fish (2), gather_clay (2), craft_pottery (ceramic)
11-15. farm (2 wheat), scavenge (iron ore, gold, bread, clay), explore (Rose Garden), travel (market), start_business (claimed market for 5g)
16-19. sleep (56 ticks!), explore (Stone Mill), mill_grain (2 flour), bake_bread_adv (3 bread)
20-24. travel (Saints Rest Shrine), drink, farm (4 wheat), explore (East Side Market), eat
25-29. fish (2), gather_clay (2), explore, scavenge, sleep (50 ticks)
30-35. explore (Crystal Well), drink, farm (1 wheat), gather_thatch, fish, eat
36-40. explore (Sunrise Stable), travel (market), explore (Iron Anvil Forge), gather_clay, craft_pottery
41-45. sleep (48 ticks rough), drink, travel (library), explore, scavenge
46-48. farm (Sunflower), explore (Oakhaven - wolves!), flee_area (INTERRUPTED)

## Locations Visited (25+)
Golden Meadow Market, Holy Stone Shrine, Oakhaven Farm, Garden Well, Dock,
Rose Garden, Sunflower Farm, The Gilded Inn, Stone Mill, Saints Rest Shrine,
Old Well, Crystal Well, South Gate Forge, East Side Market, Sunrise Stable,
Iron Anvil Forge, Scribe's Library, Hill Well

## Bugs Confirmed (on main)
1. **Scavenge picks bread over iron ore** — Both lives, picks first 5 stacks by insertion order
2. **Shrine has zero actions** — Both shrines (Holy Stone, Saints Rest) have only generic actions
3. **flee_area random location** — Life 1: 33 ticks away. Life 2: 35 ticks away. Never picks nearby safe location
4. **No auto-interrupt during farming** — Elian took 48 damage over 8 ticks while farming
5. **Sleep text "Slept at home"** — Shows when NPC has no home (fixed in our code)
6. **Business ownership no benefit** — Claimed market for 5g, zero new actions unlocked
7. **Baker has no profession-specific actions** — No bake without flour; can't cook without being inn worker

## NEW Bugs Found
8. **BUG: Rough sleep still gives "Slept at home" text** — When fatigue >= 80 and NPC rough-sleeps, it still says "Slept at home" because the fallback text is the same
9. **BUG: Explore can send you to wolf-infested locations** — Explore randomly picked Oakhaven Farm (known wolf area), leading to near-death. Explore should avoid locations with known enemies.

## Design Issues
1. **SEVERE**: Zero NPCs found across 25+ locations, 2 lives, 48 ticks
2. **HIGH**: Sleep consumes 46-50% of game ticks (3 sleeps = 154 of 500 ticks total)
3. **HIGH**: Oakhaven Farm death trap (2 deaths/near-deaths at closest farm)
4. **HIGH**: Social need drops from 75 to 0 with zero recovery (no reflect on main)
5. **HIGH**: Weather has zero gameplay effect (rain, clear, storm all identical)
6. **MEDIUM**: Baker profession is just "generic NPC with baker skill"
7. **MEDIUM**: No way to cook fish/meat without being inn worker
8. **MEDIUM**: Campfires don't exist — no solo cooking or warmth source
9. **LOW**: Ground bread accumulation (100+ stacks at markets)
10. **LOW**: Divine dreams mention companions but none appear

## New Features Implemented (Phase 4)
1. **Campfire System** — build_campfire (logs), warm_by_fire, campfire_cook
2. **Herb Gathering** — gather_herbs action at forests/gardens
3. **Salvage** — break down equipment into raw materials
4. **Bake at Campfire** — bake_campfire: wheat to bread at any fire source
5. **Herbal Tea** — brew_herbal_tea: herbs to stress/HP recovery at fire
6. **Weather Effects** — rain boosts farm/fish, storm penalizes, clear gives happiness
7. **New Items** — grilled fish, herb bread, herbal poultice
8. **Sleep Text Fix** — "Rested and slept" instead of "Slept at home" when homeless
9. **Resource Decay** — campfires burn down over time (negative regen)
