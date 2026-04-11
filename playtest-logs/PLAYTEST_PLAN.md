# Playtest Plan — Kaelen the Smith
**Session:** 2026-04-11
**NPC:** Kaelen the Smith (smith, diligent)
**Territory:** Golden Meadow
**Budget:** 48 ticks across 1 life (no deaths)

---

## Confirmed Bugs

### BUG-K1 (CRITICAL): LLM spawns unregistered "smith" profession
- **Evidence:** SpawnOnDemand returned profession="smith", skill smith:35. Registry only has "blacksmith". HasProfessionOrSkill("blacksmith", ...) fails. Even after claiming Iron Anvil Forge (owning it), zero forge actions appeared.
- **Root cause:** spawn_on_demand.go prompt lists "blacksmith" but LLM freely returns "smith". No validation/normalization of the returned profession.
- **Impact:** Any NPC spawned as "smith" has zero profession gameplay. Smelt, forge_weapon, forge_tool all unreachable.
- **Fix:** Added profession alias normalization map (smith->blacksmith, priest->healer, etc.) + allowed smith profession at forges.

### BUG-K2 (HIGH): Scavenge picks items in insertion order, not by value
- **Evidence:** Inn had 80+ bread x2 stacks plus iron ore x4, leather x2. Three consecutive scavenges each returned only bread. Iron ore never picked up.
- **Root cause:** gather.go iterates ground items array in order, takes first 5 non-gold stacks.
- **Impact:** NPCs accumulate worthless bread while ignoring iron ore, leather, and other materials needed for crafting.
- **Fix:** Sort ground items by total value (price * qty) before scavenging.

### BUG-K3 (HIGH): flee_area picks random world location (not nearest safe)
- **Evidence:** From previous playtests (PR #13, #14 logs). Code at combat.go uses rand.Intn(len(safe)).
- **Impact:** Fleeing NPC travels 160+ min while being attacked every tick. Flee is functionally a death sentence.
- **Fix:** Changed to pick nearest safe location by Manhattan distance.

### BUG-K4 (HIGH): No auto-interrupt when NPC HP drops during busy actions
- **Evidence:** runEnemyAttackPhase in tick.go attacks busy NPCs without checking HP threshold.
- **Impact:** NPCs die helplessly during forage, travel, gather, sleep in enemy zones.
- **Fix:** Added auto-interrupt: abort non-combat action when HP < 50 from enemy damage.

### BUG-K5 (MEDIUM): Trade sells first inventory item, not most valuable
- **Evidence:** From code review (economy.go lines 65-71). Confirmed by PR #13/#14 logs.
- **Impact:** NPCs sell cheapest items while hoarding valuable ones.
- **Fix:** Trade now picks most valuable item the buyer can afford, with cheapest fallback.

### BUG-K6 (LOW): Sleep fallback text says "home" when NPC has no home
- **Evidence:** survival.go line 163 always says "Slept at home" as catch-all.
- **Fix:** Changed to "Rested and slept" for homeless NPCs.

## Design/Balance Issues

### DESIGN-K1 (CRITICAL): Zero NPCs in entire world
- 12 locations visited, 22 actions committed: ZERO NPC encounters
- Social crashed 59 -> 0 with no recovery
- All social/trade/economy actions permanently blocked
- Game is solitary survival only

### DESIGN-K2 (CRITICAL): Shrines have zero actions for most NPCs
- Saints Rest Shrine visited: zero actions available
- pray requires SpiritualSensitivity > 65 or extreme desperation
- ~90% of NPCs never qualify

### DESIGN-K3 (HIGH): Sleep takes 50 real ticks (46% of game time)
- 240 game minutes = 48 ticks + travel per sleep
- 2 sleep cycles consumed ~100 of 267 game ticks

### DESIGN-K4 (HIGH): Business ownership unlocks zero new actions
- Claimed Iron Anvil Forge for 5 gold
- list_actions before/after: identical
- 5 gold wasted with zero benefit

### DESIGN-K5 (HIGH): Social need has zero solo recovery mechanism
- Social crashed to 0 in ~20 ticks with no way to recover
- No reflect/meditate/journal actions address social need

### DESIGN-K6 (MEDIUM): Weather is purely cosmetic
- Experienced clear, rain, storm: zero gameplay effects

## New Features Implemented

### 1. meditate (wellbeing)
Solo-friendly shrine action. Available to ALL NPCs without stat gates.
-12 stress, +4 happiness, -3 fatigue, +1 spiritual sensitivity.
Routes to nearest shrine. Cannot repeat consecutively.

### 2. reflect (wellbeing)
Solo social recovery. Available when social need > 40 and alone.
-15 social need, -3 stress, +2 happiness. Prevents social death spiral
in empty worlds.

### 3. stargaze (wellbeing)
Night-only outdoor action. -8 stress, +6 happiness. 12% chance of
shooting star (+1 wisdom). Available at outdoor locations at night.

### 4. tend_shrine (wellbeing)
Shrine maintenance. +2 reputation, +2 happiness. Can leave curiosity
items as offerings for bonus spiritual/rep. Encourages shrine use.

### 5. guard_patrol (combat)
Guard/knight-specific action. +combat skill, +1 reputation. Spots
enemies if present, calms nearby NPCs (-3 stress) if clear. Works
anywhere, 30 min duration.

### 6. healer_triage (wellbeing)
Healer self-heal. Uses herbs or medicine to restore own HP when
injured (HP < 90). Gains healer skill. Unique healer solo advantage.

### 7. Profession alias normalization (spawn)
LLM-returned professions normalized: smith->blacksmith, priest->healer,
soldier->guard, etc. Prevents broken profession gameplay.

### 8. Smith forge access (craft)
Smiths/blacksmiths can use forge equipment at ANY forge by profession,
not just owned/employed forges. Removes the barrier where new spawns
couldn't use their profession tools.

## Bug Fixes Implemented
- Auto-interrupt non-combat actions when HP < 50 from enemy attacks
- flee_area picks nearest safe location (not random)
- Trade sells most valuable item buyer can afford (not first slot)
- Scavenge sorts items by value before picking top 5 stacks
- Sleep fallback text corrected for homeless NPCs
- Profession validation for LLM spawn responses

## Regression Check
| Previous Fix | Status |
|-------------|--------|
| Scavenge gold cap 20 (PR #7) | CONFIRMED |
| list_actions limit 10 (PR #7) | CONFIRMED |
| Sleep at home-inn (PR #6) | CONFIRMED |
| Social decay during sleep (PR #10) | STILL SEVERE (41 pt drop) |
| Scavenge 5 stack cap (PR #12) | CONFIRMED |

## Priority Ranking
| ID | Issue | Priority | Status |
|----|-------|----------|--------|
| BUG-K1 | smith profession unrecognized | CRITICAL | FIXED |
| BUG-K2 | scavenge priority | HIGH | FIXED |
| BUG-K3 | flee_area random | HIGH | FIXED |
| BUG-K4 | no auto-interrupt | HIGH | FIXED |
| BUG-K5 | trade first item | MEDIUM | FIXED |
| BUG-K6 | sleep text | LOW | FIXED |
| DESIGN-K1 | empty world | CRITICAL | Documented |
| DESIGN-K2 | shrine no actions | CRITICAL | FIXED (meditate/tend) |
| DESIGN-K3 | sleep 50 ticks | HIGH | Documented |
| DESIGN-K4 | business no benefit | HIGH | Documented |
| DESIGN-K5 | solo social death | HIGH | FIXED (reflect) |
