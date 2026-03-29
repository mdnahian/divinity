# Playtest Plan — Margaret the Kind / Thalric the Builder

**Session:** 2026-03-29
**NPCs:** Margaret the Kind (Innkeeper, died Tick 5) -> Thalric the Builder (Innkeeper)
**Last updated:** Final — ~25 ticks played across 2 NPCs

---

## Confirmed Bugs

### BUG-T1 (MEDIUM): list_actions shows ALL candidate locations globally
- **Evidence:** drink action shows 50+ wells across entire world, including ones 300+ min travel away. gather_clay shows 60+ locations. Travel time output is enormous.
- **Root cause:** `server/gametools/npc_tools.go` line 508: `candidates := a.Candidates(n, w)` returns every location of matching type with no distance limit. The display loop prints all of them.
- **Impact:** Massively clutters the action list output, wastes tokens for AI agents, makes decision-making harder.
- **FIX APPLIED:** Sort candidates by travel distance, limit display to 10 nearest locations, add "... and N more distant locations" suffix when truncated.
- **File:** `server/gametools/npc_tools.go`

### BUG-T2 (SEVERE): Scavenge picks up unlimited gold, breaking economy
- **Evidence:** A single scavenge at The Gilded Inn picked up 631 gold. At Whispering Forest Edge, 500+ gold available. NPC went from 13 gold to 644 gold in one action.
- **Root cause:** `server/action/gather.go` scavenge Execute iterates all ground items and picks up everything via `w.PickUpGroundItem`. No cap on gold or any other item.
- **Impact:** Economy completely broken. NPCs spawn with 15 gold; one scavenge gives 40x that.
- **FIX APPLIED:** Cap gold pickup to 20 per scavenge via `maxGoldPickup` constant. Partial stacks left on ground using new `PartialPickUpGroundItem` method.
- **Files:** `server/action/gather.go`, `server/world/world.go`

### BUG-T3 (SEVERE): Ground gold accumulates faster than it decays
- **Evidence:** Previous playtests found 48 gold (PR #4), then 1156 gold (PR #6), now 631 gold at one inn and 500+ at one forest. Items have durability 80 and decay at 3/day = 27 days to disappear. Meanwhile NPCs die and drop gold daily.
- **Root cause:** `server/world/world.go` `DropItemsOnDeath` gives all items durability 80. `DecayGroundItems` reduces durability by only 3/day. Gold accumulates because it's dropped faster than it decays.
- **FIX APPLIED:** (1) Gold dropped on death now starts at durability 30 instead of 80. (2) Gold decays at 15/day instead of 3/day. Gold now disappears in ~2 days instead of ~27 days.
- **File:** `server/world/world.go`

## Design/Balance Issues

### DESIGN-T1 (CRITICAL): Whispering Forest Edge is a death trap
- **Evidence:** 7 enemies (4 wolves, 2 dire wolves, 1 bear) at the nearest forest from spawn (5 min travel). Margaret (HP 100, innkeeper) died in 2 ticks from wolf/bear attacks while foraging.
- **Impact:** New NPCs that try to forage/hunt at the nearest forest die quickly. Creates a death cycle where NPCs die, drop loot, more NPCs come, die too. 5 NPCs confirmed killed at this location.
- **Suggestion:** Reduce enemy density at forests near spawn. Consider a "safe" forest zone near market areas. Or reduce enemy attack rate (currently 50% chance per enemy per tick).

### DESIGN-T2 (MEDIUM): Duplicate NPC names at same location
- **Evidence:** 3 NPCs named "Elara" at The Gilded Inn (hunter, shepherd, barmaid).
- **Impact:** Confusing for players/agents. inspect_person and talk target first match, which may not be intended.
- **Note:** PR #6 FindNPCByNameAtLocation fix helps with cross-location resolution, but same-location duplicates remain problematic.

### DESIGN-T3 (MEDIUM): Innkeeper must leave inn to serve customers
- **Evidence:** serve_customer requires food in owner's inventory. Getting food requires: farm (travel to farm) -> mill_grain (travel to mill) -> bake_bread_adv -> travel back to inn -> serve. This is 4-5 actions and 100+ game minutes for one customer serving.
- **Impact:** Innkeepers spend most of their time farming/traveling instead of innkeeping.
- **Suggestion:** Allow buying food at market, or have employees deliver supplies.

### DESIGN-T4 (LOW): Enemy attacks during non-combat actions
- **Evidence:** Wolves/bears attacked Margaret continuously while she was foraging. No warning, no chance to flee. 4 hits totaling 61 damage in 2 ticks.
- **Suggestion:** Give NPCs a chance to flee before being attacked, or reduce attack frequency during non-combat actions.

## Improvement Ideas

1. **Limit candidate locations to nearest 10** — Reduces output verbosity for AI agents (FIX APPLIED)
2. **Cap scavenge gold at 20** — Prevents economy-breaking gold pickups (FIX APPLIED)
3. **Faster gold decay** — Gold on ground disappears in 2 days instead of 27 (FIX APPLIED)
4. **Market food supply** — Allow innkeepers to buy food at market for serving customers
5. **Enemy spawn balancing** — Reduce enemy density near spawn locations
6. **Safe zones** — Markets, inns, and shrines should be enemy-free
7. **Auto-flee at low HP** — NPCs below 20% HP should automatically flee combat

## New Feature Ideas

1. **Innkeeper auto-supply** — Hired employees could gather food for the inn automatically
2. **Business income** — Inn owners should earn passive gold when NPCs sleep/eat at their inn
3. **Profession bonus** — Innkeepers should get reduced cost for buying food supplies
4. **NPC disambiguation** — Show profession in parentheses for duplicate names in action targeting

## Regression Check

| Previous Bug | Status |
|---|---|
| BUG-M1: Gather wrong location resources | **FIXED** (foraging worked at correct forest) |
| BUG-M2: Trade only at first market | Unable to test (no tradeable items at market) |
| BUG-M3: IsWorkerAtType first location | **FIXED** (verified: innkeeper actions appeared after claiming inn) |
| BUG-M4: Inn fee first inn only | Unable to test directly (we owned the inn) |
| BUG-M5: Force-interrupt clears resume state | **STILL PRESENT** — Margaret had resume_action="forage" but chose scavenge instead; resume state was lost on death |
| BUG-M6: mill_grain first mill only | **FIXED** — mill_grain worked at Stone Mill (not first mill) |
| BUG-M7: copy_text first library only | Unable to test |
| DESIGN-M1: go_home only at night | **FIXED** — go_home appeared during day at fatigue 37% |
| BUG-E1: FindNPCByName wrong NPC | **FIXED (PR #6)** — inspect_person "Elara" found local Elara (hunter) |
| BUG-E2: Sleep at home-inn rough sleep | Unable to test (didn't sleep this session) |
| BUG-E3: Massive gold on ground | **WORSE** but FIX APPLIED (scavenge cap + faster decay) |
| DESIGN-E1: Sleep travel fatigue | Partially observed (fatigue increased during travel) |
| DESIGN-E3: Duplicate NPC names | **STILL PRESENT** — 3 Elaras at one inn |

## Working Mechanics (Confirmed)

- **eat:** +25 hunger from bread. Only appears when food in inventory and hunger < ~90%.
- **drink:** Fully restores thirst to 100%.
- **farm:** +2 wheat, +13 fatigue, farmer skill unlocked.
- **mill_grain:** 2 wheat -> 2 flour at any mill. PR #4 fix confirmed.
- **bake_bread_adv:** 2 flour -> 3 bread, cook skill unlocked.
- **talk:** +20 social, +3 relationship, dialogue stored in memory.
- **comfort:** -15 stress, -3 trauma, +10 relationship. Requires empathy.
- **gift:** +3 rep, +12 relationship, consumes food item from inventory.
- **scavenge:** Picks up all ground items. Now capped at 20 gold.
- **start_business:** Costs 5 gold, claims establishment. Unlocks serve_customer, hire_employee.
- **serve_customer:** Earns 3 gold, consumes bread from inventory, +2 relationship.
- **hire_employee:** Costs 2 gold/day, employs NPC at establishment.
- **travel:** Correct routing, fatigue increases during transit.
- **observe:** Shows NPCs, enemies, items, resources, events correctly.
- **Weather:** cloudy -> rain transitions. Atmospheric descriptions.
- **World events:** "Strange omens in the sky" with dramatic NPC death visions.
- **Memory system:** Correctly stores actions, marks vivid/defining moments.
- **Relationships:** Labels progress: neutral -> slight positive. Tracks multiple NPCs.
- **Skills:** New skills unlock on use: farmer, cook, barmaid.

## Priority Ranking

| ID | Issue | Priority | Fix Status |
|----|-------|----------|------------|
| BUG-T1 | list_actions too many locations | **MEDIUM** | FIXED |
| BUG-T2 | Scavenge unlimited gold | **SEVERE** | FIXED |
| BUG-T3 | Ground gold slow decay | **SEVERE** | FIXED |
| DESIGN-T1 | Forest death trap | **CRITICAL** | Suggestion only |
| DESIGN-T2 | Duplicate NPC names | **MEDIUM** | Suggestion only |
| DESIGN-T3 | Innkeeper supply chain | **MEDIUM** | Suggestion only |
| DESIGN-T4 | Enemy attacks during actions | **LOW** | Suggestion only |
