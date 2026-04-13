# Playtest Log: Elara the Healer / Brenna the Builder (2026-04-13)

## Life 1: Elara the Healer
- Profession: healer | Personality: virtuous, curious, mentor, generous, introverted
- Starting: Golden Meadow Market | HP 100 | Hunger 68 | Thirst 71 | Fatigue 28

### Tick Log
| Tick | Action | Location | Result | Notes |
|------|--------|----------|--------|-------|
| 1 | scavenge | Golden Meadow Market | 5x bread x2 | Bread over iron ore/logs (bug) |
| 2 | talk | Golden Meadow Market | Social +20, rel +6 with Thalric | Only NPC in territory |
| 3 | explore | -> Sunflower Farm | Arrived safely | Rain weather |
| 4 | farm | Sunflower Farm | 3 wheat harvested | Farm has 30 wheat, 20 thatch |
| 5 | drink | -> Garden Well | Thirst fully restored | 5 min travel |
| 6 | eat | Garden Well | +25 hunger from bread | Hunger 90.5 |
| 7 | forage | -> Whispering Forest Edge | DIED | Wolves: 24+18+19+18+11 = 90 dmg in 5 ticks |

**Cause of death**: Locked in forage action while dire wolves attacked. No auto-interrupt fired because HP went from 100->76->59->25->8->0 and the old threshold was HP<40.

## Life 2: Brenna the Builder
- Profession: builder | Personality: brave
- Starting: Golden Meadow Market | HP 100 | Hunger 68 | Thirst 75 | Fatigue 26

### Tick Log
| Tick | Action | Location | Result | Notes |
|------|--------|----------|--------|-------|
| 8 | scavenge | Golden Meadow Market | 5x bread x2 | Same bug: bread over ore |
| 9 | talk | Golden Meadow Market | Social +20, rel +3 with Thalric | Same sole NPC |
| 10 | drink | -> Garden Well | Thirst 100 | |
| 11 | explore | -> Deep Vein Mine | Took in sights | Mine: iron_ore 12, stone 20 |
| 12 | eat | Deep Vein Mine | +25 hunger | |
| 13 | explore | -> Golden Meadow Market | No find | Long travel |
| 14 | farm | -> Oakhaven Farm | 3 wheat | Wolves at farm! |
| 15 | gather_clay | -> Dock | 1 clay | Wolf damage: 14+8+15+11+21 = 69 dmg |
| 16 | scavenge | Dock | Failed (nothing) | |
| 17 | eat | Garden Well | +25 hunger | |
| 18 | drink | Garden Well | Thirst 100 | |
| 19-22 | various | Garden Well | Some failed (fatigue too high) | |
| 23 | sleep | -> (rough) | Started | Fatigue 73.7, HP 48 |

### Key Observations
- Dire wolf event: wolves appear at farms AND forests, making outdoor activity lethal
- Only 1 NPC (Thalric) found in entire 23-tick playtest across 6+ locations
- Fatigue rises fast: 26->73 in ~15 actions
- Scavenge consistently picks bread over valuable items (iron ore, logs)
- No herbs available without entering dangerous forests
- Builder profession has no unique actions at all
