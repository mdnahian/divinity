# Final Playtest Plan: Elara/Brenna (2026-04-13)

## Character History
- **Life 1**: Elara the Healer — spawned at Golden Meadow Market, killed by wolves at Whispering Forest Edge while foraging (tick ~7, HP 100->0 in 5 ticks)
- **Life 2**: Brenna the Builder — spawned at Golden Meadow Market, survived wolf attacks at Oakhaven Farm (HP dropped to 26), recovered to 48 HP

## Bugs Found (Priority Order)

### CRITICAL
1. **No effective auto-interrupt during combat** — NPCs locked in multi-tick actions (forage, farm) are attacked by wolves every tick but cannot react until the action completes. Elara died in 5 ticks (100->76->59->25->8->0 HP). The existing interrupt threshold (HP < 40 with enemies) is too low and only works for non-moving actions.
   - **Fix**: Raised interrupt threshold to HP < 50 for non-combat actions. Added comprehensive list of non-combat action IDs that should be interrupted when enemies attack.
   - **File**: `server/engine/turns.go` lines 182-216

2. **flee_area picks random destination** — When fleeing, NPCs pick a random safe location which can be 160+ minutes away, causing them to die during the long travel. Should pick the nearest safe location.
   - **Fix**: Changed flee_area Destination to pick nearest safe location by Manhattan distance instead of random.
   - **File**: `server/action/combat.go` lines 68-78

### HIGH
3. **Scavenge picks items in insertion order** — Market has 75+ stacks of "bread x2" mixed with iron ore x4 and logs x2. Scavenge always picks the first 5 bread stacks, missing valuable items.
   - **Fix**: Added value-based sorting to scavenge Execute function. Items are now sorted by estimated value * quantity before picking.
   - **File**: `server/action/gather.go` (scavenge Execute)
   - **New**: `server/action/survival_ext.go` (scavengeItemValue helper)

4. **Wolves at farms** — Dire wolf event spawns wolves at farm locations (Oakhaven Farm), not just forests. Brenna took 86 damage while farming wheat.
   - **Observation only** — GOD AI wolf placement issue, not a code bug

### MEDIUM
5. **Sleep fallback text "Slept at home" for homeless NPCs** — The catch-all fallback at survival.go printed "Slept at home" even for NPCs with no HomeID.
   - **Fix**: Changed fallback text to "Rested and slept" for homeless NPCs.
   - **File**: `server/action/survival.go` line 163

6. **Ground item spam** — Market has 75+ bread stacks making observe output extremely long and obscuring valuable items.
   - **Observation** — Could be addressed by grouping identical items in observe output

## New Features Implemented (all DIFFERENT from PRs #13-16)

### 1. Hunting Trap System — `set_trap`, `check_trap`
- **set_trap**: Place a hunting trap at any forest (costs 1 rope or 1 log). Max 3 traps per NPC. Gains hunter skill.
- **check_trap**: Return to check your traps. 40% daily chance to catch 1-2 raw meat (consumes game resource). Single-use traps.
- **Why**: Safe alternative to hunting — no combat risk. Solo viability for food.
- **Files**: `server/action/survival_ext.go`, `server/world/world.go` (Trap type + TickTraps), `server/engine/daily.go`

### 2. Water Canteen System — `craft_canteen`, `fill_canteen`, `drink_canteen`
- **craft_canteen**: Craft a canteen from 1 leather + 1 clay.
- **fill_canteen**: Fill at well (costs 3 water resource). Provides 3 drinks.
- **drink_canteen**: Drink anywhere (+30 thirst per drink). No well needed.
- **Why**: Eliminates constant trips to well. Major quality-of-life for exploration.
- **Files**: `server/action/survival_ext.go`, `server/item/registry.go`

### 3. Basic Tool Crafting — `craft_basic_tool`
- Craft a walking stick (1 log, weapon bonus +2) or wooden club (1 log + 1 stone, weapon bonus +5) without a forge.
- **Why**: New NPCs start unarmed. This gives a path to basic self-defense without needing blacksmith skill.
- **Files**: `server/action/survival_ext.go`

### 4. Temporary Shelter — `build_lean_to`
- Build a lean-to at any outdoor location (2 logs + 2 thatch). Sets location as temporary home.
- **Why**: Homeless NPCs suffer -10 happiness, +10 stress, -3 HP from rough sleeping. A lean-to prevents this.
- **Files**: `server/action/survival_ext.go`, `server/item/registry.go`

### 5. Snare Crafting — `craft_snare`
- Craft a snare from 1 rope. Used for trap-setting.
- **Files**: `server/action/survival_ext.go`, `server/item/registry.go`

### 6. Foraging Knowledge (Enhanced Forage)
- Herbalism skill now increases herb find chance during foraging (15% base + up to 50% bonus at max skill).
- Forage now grants herbalism skill XP (+0.2 per forage).
- **Why**: Repeated foraging becomes more rewarding, encouraging specialization.
- **Files**: `server/action/gather.go` (forage Execute), `server/action/survival_ext.go` (forageKnowledgeBonus)

## Design Issues (not fixed, documented)

1. **SEVERE**: Solo world — Only 1 NPC (Thalric the Farmer) found in 23+ ticks across 6+ locations
2. **HIGH**: Dire wolf event makes most outdoor locations lethal (farms, forests)
3. **HIGH**: Sleep consumes 48 game ticks (240 game minutes) — 50%+ of active playtime
4. **HIGH**: No social recovery for solo play (social need will crash with no other NPCs)
5. **MEDIUM**: Fatigue rises too fast (from 28 to 73 in ~15 actions)
6. **MEDIUM**: No way to heal without medicine items (herbs) which require dangerous forests

## Regression Checks
- Scavenge still respects 5 stack / 20 gold caps
- Sleep still works at home, inn, and rough
- Eat/drink/explore/farm/fish actions unchanged
- Combat (attack_enemy, flee_area) flow unchanged except flee destination fix
