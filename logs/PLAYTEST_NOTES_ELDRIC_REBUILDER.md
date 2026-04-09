## Playtest Session: Eldric the Rebuilder

- **Date:** 2026-04-09
- **NPC:** Eldric the Rebuilder (farmer)
- **Personality:** outgoing, mentor, loyal, open-minded
- **Spawn location:** Golden Meadow Market
- **Spawn tick:** 2407
- **Starting stats:** HP 100, hunger 67, thirst 80, fatigue 26
- **Token:** ca4b2ed6-5154-4dba-bcb1-9b6cbb429701
- **NPC ID:** 2ab3cb16-ce35-487d-b172-c476cdf670b8
- **Budget:** 48 ticks / ~4 hours

### Context / PR History
- PR #12 merged: scavenge item cap (5 stacks)
- PR #11 merged: recall_memories category fix
- PR #10 merged: SpawnOnDemand profession equipment, sleep travel fatigue, social decay during sleep, check_location loc.ID
- Priority per new rules: (1) fix bugs, (2) ship a NEW feature (REQUIRED), (3) polish

### Known design issues to investigate as candidates:
- Shrine has no meditate/pray action for most NPCs
- No solo eat action
- Scribe/Guard/Builder/Farmer unique profession actions
- Explore travel time wrong
- Talk cooldown invisible
- Gold accumulation at inns
- Duplicate NPC names severe
- Golden Meadow empty / zero NPCs reported last playtest

---

## Tick-by-Tick Log

### Ticks 1-15 Summary
- **T1** Golden Meadow Market empty — scavenge: 5 bread + 20 gold. **PR #12 CONFIRMED** (5 stack cap).
- **T2** Travel to Holy Stone Shrine (10 min). check_location shows only name+desc+time.
- **T3** Observe shrine. **DESIGN-M1 CONFIRMED: no pray/meditate action** with stress 11, HP 100. Eat bread solo — works (+25 hunger).
- **T4** Farm at Sunflower Farm (safer than Oakhaven, which still has 2 enemies per check_location). +1 wheat.
- **T5** Gather_thatch at Sunflower Farm. +1 thatch.
- **T6** Drink at Garden Well (thirst restored).
- **T7** Gather_clay at Garden Well. +2 clay.
- **T8** Explore → Rose Garden. No NPCs. Fatigue 74.
- **T9** Sleep. Routed to The Gilded Inn (spawn-assigned home). 49 tick duration. Final fatigue 7.
- **T10** Observe Gilded Inn — no NPCs, tons of ground items (incl. bread60 stack). Scavenge: 5 bread + 20 gold. **PR #12 confirmed again.**
- **T11** craft_pottery → 2 clay → 1 ceramic. Works solo at inn.
- **T12** Travel to East Side Market. No NPCs.
- **T13** Scavenge ESM (bread60 stack ignored in favor of 5 bread x2 stacks — **DESIGN-ISSUE: scavenge priority**)
- **T14** Travel to Scribe's Library. Library has NO copy_text/read_book/write_journal actions — confirms scribe/library design issue.
- **T15** Fish at Dock (+35 min travel). Caught 1 fish.

### Regression Status (Ticks 1-15)
- **PR #12** (scavenge item cap): CONFIRMED working at 3+ locations
- **PR #11** (recall_memories category filter): CONFIRMED — routine/economic both return correct entries
- **PR #10** (SpawnOnDemand profession equip): mostly confirmed — farmer has no special items but has farm skill 32
- **PR #10** (sleep travel fatigue fix): Fatigue at sleep commit was 74; no spike observed (stayed ~67 during travel, may still need verification)
- **PR #9** (check_location loc.ID): check_location now works but doesn't show NPCs/enemies at safe locations (only when enemies present)
- Design issues confirmed still present: DESIGN-M1 (no pray), DESIGN-M5 (duplicate names: "Hill Well" and "Dock" in list), empty world across all territories

### Bugs found this session
1. **Explore → travel time shown in list_actions is the random pick's time, not actual** (minor, existing)
2. **Scavenge picks first 5 stacks regardless of value** — misses ceramic/leather to pick 5 bread x2 stacks (DESIGN-M4-like)
3. **Solo world is completely non-viable for economy actions** — trade/gift/talk/work_together all blocked without NPCs (SEVERE for solo play)
4. **DIRE WOLVES near spawn — Sunrise Stable (+5 min from market) has 2 dire wolves!** Extreme death trap at an innocuous-looking stable.
5. **Social need crashed to 2/100 solo** — no recovery mechanism for lone NPC. Mood "lonely".
6. **flee_area doesn't actually flee instantly** — takes 160 min travel. NPC killed before fleeing.
7. **check_location doesn't warn about enemies for stables/farms** when user chooses travel target (check_location worked but not surfaced in travel lists)
8. **Owning library unlocks ZERO new actions** — start_business has no payoff (confirmed from PR #11)

### Ticks 16-24 (Eldric final)
- **T16** gather_clay at Dock (+2 clay)
- **T17** travel back to Scribe's Library (40 min)
- **T18** start_business → claimed Scribe's Library for 5 gold
- **T19** craft_pottery → 1 more ceramic (owning library gives NO new actions)
- **T20** explore → Garden Well (fatigue 64)
- **T21** drink
- **T22** travel to Sunrise Stable → DIRE WOLF AMBUSH. HP 100→63 during transit
- **T23** flee_area attempted — BUT took 160 min and was interrupted. HP 22
- **T24** flee_area again, HP 7→0 — **DIED** to dire wolves at Sunrise Stable

### DEATH
- Eldric the Rebuilder killed by 2 dire wolves at Sunrise Stable (+5 min from market). Tick 2628, game time Day 56 09:20.
- Per new rules: spawning new NPC to continue.

---

## Second Life: Mara the Gatherer (forager, brave)

- **Spawn:** Golden Meadow Market (tick ~2628)
- **Profession:** forager
- **Token:** 049ae9aa-6a6a-43c1-9690-9d77edb24547
- **Budget remaining:** 48 - 24 = 24 ticks

### Ticks 25-32 (Mara)
- **T25** Scavenge at Golden Meadow Market: 3 bread + 1 iron ore + 20 gold. PR #12 cap confirmed again.
- **T26** Drink at Garden Well (thirst fully restored).
- **T27** Eat bread (+25 hunger → 100).
- **T28** Farm at Sunflower Farm (safe): +2 wheat. Fatigue +12.
- **T29** Gather_thatch at Sunflower Farm: +1 thatch.
- **T30** Fish at Dock: caught nothing (+fishing exp). Fish has skill-based success rate.
- **T31** Gather_clay at Dock: +2 clay. Fatigue 74.
- **T32** Sleep (routed to home inn). In progress...

### Mara observations
- Same empty world as Eldric — zero NPCs anywhere
- Social declining: 68 → 46 after ~7 actions
- Stress rising: 35 (up from unknown starting value)
- Forager profession has no unique forage bonus visible (forage skill 47 but no special yield)
- fish action failed (skill-based) — farmer caught fish but forager didn't. Interesting
- Dock is safe (no enemies). Good gathering hub.

### Final Session Stats (before sleep)
- **Mara:** HP 100, hunger 98, thirst 96, fatigue 74, social 46, happiness 63, stress 30
- **Total unique actions tested this session:** scavenge, eat, drink, sleep, farm, gather_thatch, gather_clay, fish, craft_pottery, explore, travel, start_business, flee_area, check_location, recall_memories, observe, check_self, list_actions
- **Total unique locations visited:** Golden Meadow Market, Holy Stone Shrine, Sunflower Farm, Garden Well, Rose Garden, The Gilded Inn, East Side Market, Scribe's Library, Dock, Sunrise Stable (DEATH)

