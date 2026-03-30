# Playtest Plan — Elara (Guard)

**Session:** 2026-03-30
**NPC:** Elara (Guard at Mistwell Village -> Rat's Nest)
**Personality:** virtuous, generous
**Last updated:** Final — 24 ticks played across ~2 hours

---

## Confirmed Bugs

### BUG-G1 (MEDIUM): check_location always shows "~0 min" travel time
- **Evidence:** Checked Peakwatcher Citadel, Aethelgard Market, Whispering Pines, Valley Manor, Stormhoof Stables, Saints Peak Shrine — all showed "~0 min" travel time despite being 5-20 min away per list_actions.
- **Root cause:** check_location likely calculates travel time from the target location to itself (0) instead of from the NPC's current location.
- **Impact:** Players cannot scout travel times before deciding to visit. Makes check_location unreliable for planning.
- **Fix:** Pass NPC's current location as origin when calculating travel time in check_location handler.

### BUG-G2 (LOW): inspect_person resolves requesting NPC as target with duplicate names
- **Evidence:** inspect_person "Elara" at Rat's Nest (where 6 NPCs are named Elara) returned our own NPC (guard, mood: lonely) instead of another Elara.
- **Root cause:** FindNPCByNameAtLocation (PR #6 fix) finds any NPC with that name at the location, including the requesting NPC itself.
- **Impact:** Cannot inspect other NPCs who share your name. With 6 Elaras at one inn, inspect is useless for 5 of them.
- **Fix:** Filter out the requesting NPC's ID from FindNPCByNameAtLocation results.

### BUG-G3 (LOW): Memory duplicates work_together event text
- **Evidence:** Memory shows both "Worked together with Elara. Improved our skills." and "Worked together with Elara -- improved guard skills and bond." for the same action.
- **Impact:** Minor — wastes memory context tokens with duplicate information.

## Design/Balance Issues

### DESIGN-G1 (HIGH): Guard profession has zero unique actions
- **Evidence:** Guard skill 47. Visited market, castle, barracks (Guard Post), forge, well, shrine, 2 inns. No guard-specific actions ever appeared — no patrol, defend, guard_gate, arrest, or any combat-profession action.
- **Impact:** Guard is a completely non-functional profession. Skill 47 provides zero gameplay benefit. Players who spawn as guards have no profession-specific gameplay.
- **Suggestion:** Add guard actions: patrol (at barracks/castle, reduces crime), defend (protect location from enemies), guard_gate (at castle, earn gold), arrest (social, target hostile NPC).

### DESIGN-G2 (HIGH): Extreme NPC population concentration at inns
- **Evidence:** 10 locations visited. 8 were completely empty (0 NPCs). Only the 2 inns had NPCs (6 and 8 respectively). Markets, castles, farms, forges, wells, shrines, barracks — all deserted.
- **Impact:** Social stat plummets for any NPC not at an inn. Our social went from 66 to 0 in 15 ticks of exploring. Forces all gameplay to happen at inns.
- **Suggestion:** Distribute NPCs more evenly. Add NPC schedules that send them to markets, farms, and other locations during the day.

### DESIGN-G3 (SEVERE): Duplicate NPC names at extreme levels
- **Evidence:** 6 NPCs named "Elara" at Rat's Nest (potter, 2 priests, herbalist, scribe, + our guard). 3 more Elaras at Drowned Lantern Inn. 2 Eldrics at same inn.
- **Impact:** (1) inspect_person returns wrong NPC (BUG-G2). (2) talk/gift/comfort target unpredictable NPC. (3) Relationship tracking merges all same-named NPCs into one entry. (4) Memory text is confusing ("Elara said..." when both parties are Elara).
- **Suggestion:** Name generation MUST check for existing names at same location/territory. Add surnames or titles for disambiguation.

### DESIGN-G4 (MEDIUM): Sleep travel fatigue causes fatigue to spike during sleep
- **Evidence:** Fatigue went from 75 to 87.2 (+12) during travel to home for sleep (240 min travel). If starting fatigue was 88+, NPC would collapse during transit to bed.
- **This is DESIGN-E1 from previous playtests — STILL PRESENT.**
- **Suggestion:** Reduce or eliminate travel fatigue during sleep action. The NPC is "going to bed" — adding fatigue is counterproductive.

### DESIGN-G5 (MEDIUM): Gold accumulation still excessive despite PR #7 decay fix
- **Evidence:** 185 gold at Peakwatcher Citadel, 620 gold at Rat's Nest (564+56), 728 gold at Drowned Lantern Inn (698+15+15). Total 1533 gold observed on ground across 3 locations.
- **Impact:** Scavenge gold cap at 20 helps limit per-action gain, but total gold on ground is still massive. Economy still distorted.
- **Suggestion:** Further increase gold decay rate, or reduce gold dropped on NPC death, or add gold despawn after 1 game day.

### DESIGN-G6 (LOW): Shrines have no unique actions
- **Evidence:** Silent Shrine (shrine) and Saints Peak Shrine — no pray, worship, meditate, or any shrine-specific action.
- **Impact:** Shrines are purely decorative. Missed opportunity for spiritual/stress-relief gameplay.
- **Suggestion:** Add pray action at shrines (-10 stress, +5 happiness, small divine favor chance).

### DESIGN-G7 (LOW): Social drain too fast without NPCs
- **Evidence:** Social went from 66 to 0 in ~15 ticks (no NPCs available for first 15 ticks). Rate is ~4.4/tick.
- **Impact:** Combined with DESIGN-G2 (empty non-inn locations), NPCs exploring or gathering lose all social rapidly.
- **Suggestion:** Reduce social drain rate when NPC is busy with an action (working, foraging, etc.).

## Improvement Ideas

1. **Add guard profession actions** — patrol, defend, guard_gate at barracks/castle locations
2. **Fix check_location travel time** — calculate from NPC position, not from destination to itself
3. **Filter self from name resolution** — exclude requesting NPC from FindNPCByNameAtLocation
4. **Distribute NPC population** — schedules that move NPCs to markets/farms during day
5. **Unique NPC names per territory** — prevent duplicate names at same location
6. **Add shrine actions** — pray, meditate for stress relief
7. **Reduce sleep travel fatigue** — cap or eliminate fatigue during transit to sleep
8. **Faster gold decay** — increase from 15/day to 30/day, or add cap per location

## New Feature Ideas

1. **Guard actions:** patrol (+security at location), defend (protect NPCs from enemies), guard_gate (earn gold at castle)
2. **Shrine actions:** pray (-stress, +happiness), meditate (-fatigue at slow rate), divine blessing (rare buff)
3. **NPC schedules:** NPCs move to workplace during day, inn at night. Markets get vendor NPCs.
4. **Profession title system:** "Elara the Guard" vs "Elara the Potter" for disambiguation
5. **Social activities at shrines/gardens** — group meditation, community gathering events

## Regression Check

| Previous Bug | Status |
|---|---|
| BUG-M1: Gather wrong location resources | Unable to test (didn't gather at location with resource tracking) |
| BUG-M2: Trade only at first market | Unable to test (no trade action appeared) |
| BUG-M3: IsWorkerAtType first location | Unable to test |
| BUG-M4: Inn fee first inn only | Unable to test (slept at home) |
| BUG-M5: Force-interrupt clears resume state | Unable to test (no interrupts) |
| BUG-M6: mill_grain first mill only | Unable to test (didn't mill) |
| BUG-M7: copy_text first library only | Unable to test |
| DESIGN-M1: go_home only at night | **FIXED** (PR #4) — go_home appeared at 18:40 with fatigue 66 |
| BUG-E1: FindNPCByName wrong NPC | **PARTIALLY FIXED** (PR #6) — local resolution works, but self-targeting issue with duplicates |
| BUG-E2: Sleep at home-inn rough sleep | **FIXED** (PR #6) — "Slept at home (-40 fatigue, -5 stress)" at inn-home |
| BUG-E3: Massive gold on ground | **PARTIALLY FIXED** (PR #7) — cap works, decay faster, but accumulation still 500-700 at inns |
| BUG-T1: list_actions too many locations | **FIXED** (PR #7) — shows 10 + "...and N more" |
| BUG-T2: Scavenge unlimited gold | **FIXED** (PR #7) — capped at 20 per scavenge (tested 3 times) |
| BUG-T3: Ground gold slow decay | **PARTIALLY FIXED** (PR #7) — faster decay but still accumulating |
| DESIGN-E1: Sleep travel fatigue | **STILL PRESENT** — fatigue spiked from 75 to 87 during sleep travel |
| DESIGN-E3: Duplicate NPC names | **STILL PRESENT** — 6 Elaras at one inn, 3 at another |
| DESIGN-T1: Forest death trap | Unable to test (didn't visit dangerous forest) |

## Working Mechanics (Confirmed)

- **eat:** +25 hunger from bread. Appears when hunger < ~90%.
- **drink:** Fully restores thirst. Correctly routes to nearest well.
- **farm:** +4 wheat, +14 fatigue (including travel). farmer skill unlocked.
- **gather_clay:** +2 clay from well location. +11 fatigue.
- **craft_pottery:** 2 clay -> 1 ceramic. +9 fatigue.
- **scavenge:** Picks up ground items. Gold capped at 20 per action.
- **sleep:** Routes to home (Rat's Nest inn). Free at home. -40 fatigue, -5 stress. Fatigue decays during sleep.
- **talk:** +20 social, +3 relationship. Dialogue stored in memory and target memory.
- **work_together:** Requires same-profession NPC at location. Improves skills and bond. +20 fatigue.
- **explore:** Random nearby location. +8 fatigue.
- **travel:** Correct routing and fatigue calculation.
- **observe:** Shows NPCs, items, resources, events, weather correctly.
- **check_self:** All stats, inventory, skills, relationships displayed correctly.
- **list_actions:** Location-aware with "(here)" label. Travel times correct. Capped at 10 locations.
- **Memory system:** Recent events, insights at midnight, vivid/defining moments.
- **Relationship tracking:** Labels from neutral -> slight positive. Tracks interaction types.
- **Weather:** clear -> storm -> rain -> cloudy transitions.
- **World events:** Dire wolf sightings, divine omens, marketplace golden light.
- **Memory insights:** Generated reflective thoughts at midnight based on day's activities.

## Priority Ranking

| ID | Issue | Priority | Fix Status |
|----|-------|----------|------------|
| DESIGN-G1 | Guard profession no actions | **HIGH** | Suggestion only |
| DESIGN-G2 | NPC population concentrated at inns | **HIGH** | Suggestion only |
| DESIGN-G3 | Extreme duplicate NPC names | **SEVERE** | Suggestion only |
| DESIGN-G4 | Sleep travel fatigue spike | **MEDIUM** | Suggestion only |
| DESIGN-G5 | Gold accumulation still excessive | **MEDIUM** | Suggestion only |
| BUG-G1 | check_location 0 min travel | **MEDIUM** | TO FIX |
| BUG-G2 | inspect_person returns self | **LOW** | TO FIX |
| BUG-G3 | Memory duplicate work_together | **LOW** | Investigation needed |
| DESIGN-G6 | Shrines no actions | **LOW** | Suggestion only |
| DESIGN-G7 | Social drain too fast | **LOW** | Suggestion only |
