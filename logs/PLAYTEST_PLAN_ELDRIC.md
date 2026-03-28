# Playtest Plan — Eldric (Miller)

**Session:** 2026-03-28
**NPC:** Eldric (Mill worker at Whispering Peak Inn)
**Personality:** empathetic, diligent, addictive personality
**Last updated:** Final — ~20 ticks played

---

## Confirmed Bugs

### BUG-E1 (MEDIUM): FindNPCByName resolves wrong NPC with duplicate names
- **Evidence:** Tried to talk to "Elara" and "Eldrin" at The Shivering Hearth. Both returned "X is no longer here" despite observe listing them as present.
- **Root cause:** `server/gametools/npc_tools.go` line 611 calls `w.FindNPCByName(params.Target)` which returns the FIRST alive NPC with that name globally (server/world/world.go line 653). When multiple NPCs share a name, it finds one at a different location.
- **Impact:** Agents/players can't reliably target common-named NPCs. With many NPCs named "Elara", "Eldric", "Eldrin", this breaks social interactions.
- **Fix:** In the commit_action handler (npc_tools.go line 611), find NPC by name AT the current location first. Fall back to global search only if no local match.

### BUG-E2 (HIGH): Sleep at home-inn gives rough sleep instead of free home sleep
- **Evidence:** NPC traveled to Whispering Peak Inn (HomeID), but sleep result was "Slept rough on the ground" with penalties (-3 HP, -10 happiness, +10 stress, -10 hygiene).
- **Root cause:** `server/action/survival.go` line 108: `needsInn := n.HomeID == "" || (atInn && n.HomeID == n.LocationID)`. When home IS an inn, `needsInn` is true, so the code tries to charge an inn fee. If the inn is unowned, `canAffordInn` is false and it falls through to rough sleep.
- **Impact:** Any NPC whose home is an inn (spawned there, or claimed it) will rough-sleep at their own home if the inn is unowned or overcrowded. This causes unnecessary stat penalties.
- **Fix:** Check if NPC is at their home first. If `n.LocationID == n.HomeID`, give free home sleep regardless of inn status.

### BUG-E3 (DESIGN - SEVERE): Massive gold accumulation on ground
- **Evidence:** 1156 gold on ground at Mist-Flow Mill. Previous playtest (PR #4) found 48 gold — this is 24x worse.
- **Likely cause:** Dead NPCs dropping all their gold, with no decay/despawn system for ground items.
- **Impact:** Economy completely breaks when anyone scavenges. Our NPC got 1197 gold from two scavenge actions.
- **Suggestion:** Add gold decay on ground (e.g., despawn after 5 game days), or cap scavenge gold to 10-20 per pickup.

## Design/Balance Issues

### DESIGN-E1: Sleep travel fatigue can cause collapse (MEDIUM)
- **Evidence:** Fatigue went from 74 to 84 during travel to home for sleep. If fatigue was higher (~92+), the NPC would collapse during travel.
- **Impact:** NPCs with distant homes may collapse trying to reach bed.
- **Suggestion:** Reduce or eliminate travel fatigue for sleep actions, since the NPC is "going to bed."

### DESIGN-E2: Forage/action routing may pick far locations despite nearby match (LOW)
- **Evidence:** Forage with location "Whispering Pine Forest" (+5 min travel) resulted in ~47 tick busy time, suggesting routing to a much farther forest.
- **Possible cause:** Location capacity check in destNearestOfType skipping the nearest forest because it's "full."
- **Impact:** NPCs waste time traveling to distant locations when a nearby one should work.

### DESIGN-E3: Excessive duplicate NPC names (MEDIUM)
- **Evidence:** 5+ NPCs named "Elara", 3+ named "Eldric", 2+ named "Eldrin" at same location.
- **Impact:** Combined with BUG-E1, makes social interactions unreliable. Also confusing for observers.
- **Suggestion:** Name generation should avoid duplicates within the same territory, or add distinguishing suffixes.

## Improvement Ideas

1. **Local-first NPC name resolution** — When targeting by name, prefer NPCs at same location (fixes BUG-E1)
2. **Home-aware sleep** — If at HomeID, always give free proper sleep (fixes BUG-E2)
3. **Ground item decay** — Despawn gold/items on ground after N game days
4. **Scavenge gold cap** — Limit gold pickup to 20 per scavenge action
5. **Travel fatigue exemption for sleep** — Reduce fatigue gain while traveling to sleep

## New Feature Ideas

1. **Miller profession bonus** — Millers should get bonus flour yield or reduced fatigue at mills
2. **Inn owner income** — As inn owner, should receive gold when guests sleep/rest (now that we own the inn)
3. **NPC nickname system** — Allow NPCs to earn nicknames to help with disambiguation

## Regression Check

| Previous Bug | Status |
|---|---|
| BUG-M1: Gather wrong location resources | **FIXED** (code verified: LocationByID) |
| BUG-M2: Trade only at first market | **FIXED** (code verified: loc.Type check) |
| BUG-M3: IsWorkerAtType first location | **FIXED** (code verified: iterates all) |
| BUG-M4: Inn fee first inn only | **FIXED** (code verified: loc.Type == "inn") |
| BUG-M5: Force-interrupt clears resume state | **STILL PRESENT** (line 297 in turns.go) |
| BUG-M6: mill_grain first mill only | **FIXED** (gameplay verified at Mist-Flow Mill) |
| BUG-M7: copy_text first library only | **FIXED** (code verified: iterates all libs) |
| DESIGN-M1: go_home only at night | **FIXED** (gameplay verified: available at 09:35 with fatigue 61) |
| DESIGN-M2: Large gold on ground | **WORSE** (1156 gold vs 48 gold previously) |
| BUG-J1: gather_clay no location names | **FIXED** (location names shown) |
| BUG-J3: Session token expiry | **NOT TRIGGERED** (session active ~2 hours) |
| BUG-J4: Duplicate world events | **NOT TRIGGERED** (single event text each time) |

## Working Mechanics (Confirmed)

- **scavenge:** Picks up ALL ground items correctly
- **talk:** +20 social, +6 relationship (charisma bonus), dialogue stored in memory + target memory
- **farm:** +1-4 wheat, +12 fatigue, +0.3 farmer skill
- **mill_grain:** 2 wheat -> 2 flour at any mill (PR #4 fix confirmed)
- **bake_bread_adv:** 2 flour -> 3 bread, +6 fatigue, +0.4 cook skill
- **drink:** Fully restores thirst to 100%, uses 1 well water
- **eat:** Bread +25 hunger, consumes 1 item
- **explore:** Random nearby location within 60 units, 20% chance to find item, +5 fatigue
- **brew_potion:** 2 herbs -> 1 healing potion, +5 fatigue, +0.5 herbalist skill
- **gift:** Gives food to hungry NPC, +12 relationship, +3 reputation
- **comfort:** -15 stress, -3 trauma on target, +10 relationship
- **start_business:** Claims unowned establishment for 5 gold
- **sleep:** Routes to home/inn, -40 fatigue, -5 stress on completion. DecayNeeds reduces fatigue during sleep.
- **go_home:** Available at night OR when fatigue >= 60 (PR #4 fix confirmed)
- **Memory system:** Properly stores actions, marks vivid/defining moments, propagates rumors
- **Relationship tracking:** Labels progress: neutral -> slight positive
- **Weather transitions:** clear -> rain -> storm -> cloudy observed
- **World events:** Spiritual crisis events with atmospheric text
- **Location descriptions:** Correct names, types, travel times

## Priority Ranking

| ID | Issue | Priority | Fix Status |
|----|-------|----------|------------|
| BUG-E2 | Sleep at home-inn rough sleep | **HIGH** | TO FIX |
| BUG-E1 | FindNPCByName wrong NPC | **MEDIUM** | TO FIX |
| BUG-E3 | Massive gold on ground | **DESIGN** | Suggestion only |
| DESIGN-E1 | Sleep travel fatigue | **LOW** | Suggestion only |
| DESIGN-E2 | Forage routing far | **LOW** | Investigation needed |
| DESIGN-E3 | Duplicate NPC names | **MEDIUM** | Suggestion only |
