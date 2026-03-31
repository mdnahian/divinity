# Playtest Plan — Elara (Miner)

**Session:** 2026-03-31
**NPC:** Elara (Miner at East Side Market -> various -> Rat's Nest)
**Personality:** greedy, empathetic, neurotic, unscrupulous, introverted, addictive personality
**Last updated:** Final — 21 actions played across ~2 hours

---

## Confirmed Bugs

### BUG-M1 (HIGH): SpawnOnDemand missing profession-specific equipment
- **Evidence:** Spawned as miner (skill 37) but no pickaxe in inventory. mine_ore and mine_stone require pickaxe equipped (gather.go lines 278, 315). Visited forge, market, well, shrine, 2 inns — no mining actions ever appeared despite miner profession.
- **Root cause:** `server/god/spawn_on_demand.go` line 111 only gives `bread x2, gold x15` as StartItems. Compare with `server/god/genesis.go` lines 214-231 and 1318-1342 which give profession-specific items:
  - Miner: pickaxe (equipped)
  - Carpenter: iron axe (equipped), logs x2, rope x2
  - Blacksmith: iron ore x4, iron ingot x2
  - Tailor: leather x2, cloth x2
  - Guard/Knight: iron sword (equipped), leather armor (equipped)
  - Stablehand: hay x5
- **Impact:** ALL on-demand spawned profession NPCs cannot perform their profession's core actions. Miners can't mine. Guards have no weapon. Carpenters can't chop. This effectively breaks profession gameplay for all player-spawned NPCs.
- **Fix:** Add profession-specific equipment to SpawnOnDemand after NPC creation, matching genesis.go logic.

### BUG-M2 (MEDIUM): check_location always shows "~0 min" travel time and no NPCs/enemies
- **Evidence:** Checked The Gilded Inn (+5 min per list_actions), Saints Rest Shrine (+5 min), Iron Anvil Forge (+10 min) — all showed "~0 min" travel. None showed NPCs despite some having 14+ residents.
- **Root cause:** `server/gametools/npc_tools.go` lines 452, 461, 466-467 use `params.LocationID` (the name string user typed) instead of `loc.ID` (the resolved UUID). `NPCsAtLocation`, `EnemiesAtLocation`, and `TravelTicks` all receive a name string instead of an ID, producing wrong results.
- **Impact:** check_location is completely broken for scouting. Travel time always shows 0, NPCs and enemies are never shown.
- **Note:** PR #9 fixes this (changing params.LocationID to loc.ID in these 4 places) but is not yet merged.

### BUG-M3 (LOW): Trade action doesn't store memory for the seller
- **Evidence:** Sold bread to Eldrin for 3 gold. Then `recall_memories` with category "economic" returned "No economic memories." The trade is stored in recent memory text but without CatEconomic category tag.
- **Root cause:** `server/action/economy.go` line 96 only calls `mem.Add(target.ID, ...)` for the buyer. No memory entry is created for the selling NPC.
- **Impact:** Traders can't filter trade history by category. Minor impact on AI agent recall.
- **Fix:** Add `mem.Add(n.ID, memory.Entry{Text: fmt.Sprintf("Sold %s to %s for %d gold.", itemName, target.Name, price), Category: memory.CatEconomic, ...})` for the seller.

### BUG-M4 (MEDIUM): Sleep travel fatigue spike (DESIGN-E1, STILL PRESENT)
- **Evidence:** Fatigue was 70 before sleep commit. Instantly jumped to 79.7 (+9.7) from 160 min travel to home. If fatigue had been 90, it would reach 99.7 — near collapse.
- **Root cause:** `server/engine/turns.go` lines 342-343 apply all travel fatigue immediately when action is submitted, even for sleep/go_home actions.
- **Impact:** NPCs with high fatigue risk collapse during transit to bed. Counterproductive — trying to sleep makes you more tired.
- **Fix:** Skip or reduce travel fatigue when actionID is "sleep" or "go_home".

## Design/Balance Issues

### DESIGN-M1 (HIGH): Shrine pray action has too-restrictive conditions
- **Evidence:** Visited Saints Rest Shrine — no pray action despite being AT the shrine. `server/action/wellbeing.go` lines 63-68 require: SpiritualSensitivity > 65 (believer path) OR stress > 85 / HP < 30 / happiness < 15 (desperate path). Our NPC with stress 6, HP 100, happiness 56 met neither.
- **Impact:** Shrines are functionally useless for ~80% of NPCs. Only deeply spiritual or near-death NPCs can pray.
- **Suggestion:** Lower thresholds: believer path stress > 20 or happiness < 60; desperate path stress > 50 or happiness < 30. Or add a "meditate" action with no stat requirements.

### DESIGN-M2 (MEDIUM): work_together doesn't recover social need
- **Evidence:** Social was 56 before work_together, continued declining after. Code at `social.go` lines 196-217 modifies fatigue, skill, and relationship but NOT SocialNeed.
- **Impact:** Working alongside someone doesn't help loneliness. This is counter-intuitive — collaboration should be social.
- **Suggestion:** Add `n.Needs.SocialNeed = clampF(n.Needs.SocialNeed-10, 0, 100)` for both parties.

### DESIGN-M3 (MEDIUM): Social need decays during sleep
- **Evidence:** Social was 72 when sleep started (tick 1649). After 80 ticks of sleep, social was 31. Decay: 0.5/tick * 80 = 40 social need increase during sleep.
- **Root cause:** `npc.go` line 667 `n.Needs.SocialNeed = clampF(n.Needs.SocialNeed+d.SocialNeed, 0, 100)` runs every tick including during sleep. Unlike hunger/thirst which have halved decay during sleep (line 659-660), social need has no sleep exception.
- **Impact:** NPCs wake up lonely after long sleep. Combined with long travel-to-home times, social crashes.
- **Suggestion:** Reduce or zero social need decay during sleep, matching hunger/thirst treatment.

### DESIGN-M4 (MEDIUM): Trade always sells first inventory item — no choice
- **Evidence:** Had bread and ceramic in inventory. Trade sold bread (lower value) despite ceramic being worth more. Code at `economy.go` lines 65-70 picks first non-gold item with `break` on first match.
- **Impact:** Players/AI can't strategize about what to sell. Higher-value items later in inventory are never sold.
- **Suggestion:** Sell highest-value item first, or allow specifying item via action parameter.

### DESIGN-M5 (SEVERE): Extreme duplicate NPC names
- **Evidence:** 10 Elaras at East Side Market, 7 at The Gilded Inn, 7 at Rat's Nest. Multiple Eldrics, Eldrins, Eamons everywhere. Names generated by LLM in SpawnOnDemand without checking for existing names.
- **Impact:** inspect_person may return wrong NPC. Relationships tracked incorrectly. Memory text is confusing.

### DESIGN-M6 (MEDIUM): Gold accumulation still excessive at inns
- **Evidence:** 334 gold at The Gilded Inn, 165 gold at Rat's Nest. Despite PR #7 decay fix.
- **Impact:** Economy distortion. Scavenging at populated locations yields abundant gold.

### DESIGN-M7 (HIGH): Miners (and other tool-dependent professions) have no path to obtain tools
- **Evidence:** Miner spawned without pickaxe. forge_tool requires blacksmith skill >= 25 AND worker at forge AND iron ingot. No "buy" action exists. Only way for miner to get pickaxe: find a blacksmith who has forged one AND can trade/gift it. Extremely unlikely in practice.
- **Impact:** Tool-dependent professions are permanently locked out of their core gameplay after on-demand spawn.
- **Suggestion:** Either fix BUG-M1 (give tools at spawn) AND/OR add a "buy_tool" action at forges/markets.

## Improvement Ideas

1. **Give profession equipment at spawn** — match genesis.go logic in SpawnOnDemand (fixes BUG-M1)
2. **Fix check_location** — use loc.ID instead of params.LocationID (fixes BUG-M2, PR #9)
3. **Add seller memory for trades** — CatEconomic memory for trading NPC (fixes BUG-M3)
4. **Skip sleep travel fatigue** — exempt sleep/go_home from travel fatigue (fixes BUG-M4)
5. **Lower pray conditions** — make shrines accessible to more NPCs (fixes DESIGN-M1)
6. **Social recovery from work_together** — add SocialNeed reduction (fixes DESIGN-M2)
7. **Reduce social decay during sleep** — halve or zero social drain while sleeping (fixes DESIGN-M3)
8. **Smart trade item selection** — sell highest-value item or allow choice (fixes DESIGN-M4)
9. **Merge PR #9** — contains check_location fix and self-exclusion for name resolution

## New Feature Ideas

1. **Buy action at markets** — purchase items from market at price, enabling tool acquisition
2. **Miner profession bonus** — reduced fatigue at mines, bonus ore yield
3. **NPC name deduplication** — SpawnOnDemand should avoid names already in use at same territory
4. **Tool lending/borrowing** — NPCs could lend tools to same-profession colleagues
5. **Meditation at shrines** — low-requirement action for stress relief without spiritual stat check

## Regression Check

| Previous Bug | Status |
|---|---|
| BUG-M1: Gather wrong location resources (PR #1) | Unable to test (didn't gather at specific resource location) |
| BUG-M2: Trade only at first market (PR #4) | **FIXED** — trade worked at East Side Market (not first market) |
| BUG-M3: IsWorkerAtType first location (PR #4) | Unable to test |
| BUG-M4: Inn fee first inn only (PR #4) | Unable to test |
| BUG-M5: Force-interrupt clears resume state | Unable to test (no interrupts) |
| BUG-M6: mill_grain first mill only (PR #4) | Unable to test |
| BUG-M7: copy_text first library only (PR #4) | Unable to test |
| DESIGN-M1: go_home only at night (PR #4) | Not directly tested (sleep appeared at fatigue 60 during day) |
| BUG-E1: FindNPCByName wrong NPC (PR #6) | **FIXED** — inspect_person found local Elara at location |
| BUG-E2: Sleep at home-inn rough sleep (PR #6) | **FIXED** — "Slept at home (-40 fatigue, -5 stress)" at Rat's Nest inn |
| BUG-E3: Massive gold on ground (PR #7) | **PARTIALLY FIXED** — gold cap works, but 334 gold at one inn |
| BUG-T1: list_actions too many locations (PR #7) | **FIXED** — shows 10 + "...and N more" |
| BUG-T2: Scavenge unlimited gold (PR #7) | **FIXED** — capped at 20 per scavenge (tested 3 times) |
| BUG-T3: Ground gold slow decay (PR #7) | **PARTIALLY FIXED** — faster decay but still accumulating |
| DESIGN-E1: Sleep travel fatigue | **STILL PRESENT** — fatigue spiked from 70 to 79.7 during sleep travel |
| DESIGN-E3: Duplicate NPC names | **STILL PRESENT** — 10 Elaras at one market |
| BUG-G1: check_location 0 min travel (PR #9 unmerged) | **STILL PRESENT** — confirmed at 3 locations |
| BUG-G2: inspect_person returns self (PR #9 unmerged) | **NOT TRIGGERED** — other Elara found first |

## Working Mechanics (Confirmed)

- **eat:** +25 hunger from bread, appears when hunger < ~90%. Consumes 1 item.
- **drink:** Fully restores thirst to ~100%. Routes to nearest well.
- **eat_together:** +15 hunger + +15 social + +8 relationship. Consumes 1 food. Requires hunger < 70.
- **talk:** +20 social, +3 relationship. Dialogue stored in memory and target memory.
- **comfort:** -15 stress, -3 trauma, +10 relationship. Requires empathy > 45 on actor.
- **gift:** Gives food to hungry NPC. +3 rep, +12 relationship.
- **steal:** Steals gold from NPC. Rep loss, relationship negative. Can be caught.
- **work_together:** +skill, +5 relationship, +12 fatigue. Same profession required.
- **trade:** Sells first non-gold item at market price. Dynamic pricing works. +5 relationship.
- **scavenge:** Picks up ground items. Gold capped at 20 per action.
- **craft_pottery:** 2 clay -> 1 ceramic. +9 fatigue.
- **gather_clay:** +2 clay at well locations. +11 fatigue.
- **sleep:** Routes to home. Free at home inn. -40 fatigue, -5 stress. Fatigue decays 0.5/tick during sleep.
- **explore:** Random nearby location. Travel fatigue varies.
- **travel:** Correct routing, fatigue based on distance (3.0 per 10 ticks walking).
- **observe:** Shows NPCs, items, resources, events, weather. Correct at all locations.
- **check_self:** All stats, inventory, skills, relationships correct.
- **list_actions:** Location-aware with "(here)" label. Travel times correct. Capped at 10 locations (PR #7).
- **recall_memories:** Topic/category filtering works. Vivid/defining moments tracked.
- **Dynamic economy:** Bread price fluctuates based on supply/demand (sold for 3 vs base 2).
- **Weather:** clear -> cloudy -> rain transitions observed.
- **World events:** "Strange omens" atmospheric events. NPC death notifications during sleep.
- **Relationship system:** Labels progress with interaction types tracked.

## Priority Ranking

| ID | Issue | Priority | Fix Status |
|----|-------|----------|------------|
| BUG-M1 | SpawnOnDemand no profession equipment | **HIGH** | TO FIX |
| BUG-M2 | check_location 0 min travel / no NPCs | **MEDIUM** | PR #9 (unmerged) |
| BUG-M3 | Trade no seller memory | **LOW** | TO FIX |
| BUG-M4 | Sleep travel fatigue spike | **MEDIUM** | TO FIX |
| DESIGN-M1 | Shrine pray too restrictive | **HIGH** | Suggestion only |
| DESIGN-M2 | work_together no social | **MEDIUM** | Suggestion only |
| DESIGN-M3 | Social decays during sleep | **MEDIUM** | TO FIX |
| DESIGN-M4 | Trade sells first item only | **LOW** | Suggestion only |
| DESIGN-M5 | Duplicate NPC names | **SEVERE** | Suggestion only |
| DESIGN-M6 | Gold accumulation at inns | **MEDIUM** | Suggestion only |
| DESIGN-M7 | No tool acquisition path | **HIGH** | Suggestion only |
