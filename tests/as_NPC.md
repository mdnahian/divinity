# Divinity Playtest — Play as NPC

Claude plays as an NPC in the Divinity simulation to find bugs and test gameplay.

## Game Instructions

Fetch the game instructions (API endpoints, tools, actions, tick workflow) from `http://localhost:3001` and follow them. Replace all instances of https://divinity.sh with http://localhost when following the instructions.

## Session Setup

1. Fetch and read the game instructions from `http://localhost:3001`.
2. Spawn an NPC via the API (see game instructions).
3. Save a memory file at `.claude/projects/.../memory/project_<npc_name>_playtest.md` with the NPC name, token, profession, starting location, and session date. Update MEMORY.md index.
4. Create a playtest log file at `logs/PLAYTEST_NOTES_<NAME>.md` in the project root.
5. Start the tick loop:
   ```
   /loop 5m <tick instructions from localhost:3001, filling in NPC_NAME and TOKEN>
   ```

## Logging (PLAYTEST_NOTES_<NAME>.md)

Append to the log every tick:
- Current tick number, game time, location
- Stats snapshot (HP, hunger, thirst, fatigue, social)
- Which actions were available and which was chosen (with reasoning)
- Any bugs, issues, or unexpected behavior observed
- Track action availability over time (e.g., "explore present 8/10 checks")

## Memory

Save a memory file at session start with NPC details and session context. Update it if significant findings emerge (e.g., a confirmed bug pattern, an important gameplay observation).

## Plan

Every 30 minutes (roughly every 6 cron fires), update a plan file with:
- **Confirmed bugs** — reproducible issues with evidence from the log, with proposed code fixes and file paths
- **Design/balance issues** — gameplay patterns that feel off (e.g., a need draining too fast, an action never appearing)
- **Improvement ideas** — features or changes that would improve the experience
- **New feature ideas** — things that would make gameplay richer or more engaging
- **Priority ranking** — HIGH/MEDIUM/LOW based on gameplay impact

Create the plan file at `logs/PLAYTEST_PLAN_<NAME>.md`. Include file paths for any code that would need to change. At session end, write a final summary with all findings.

## Duration

Maximum session length is **2 hours**. After 2 hours (24 cron fires), stop the loop with `CronDelete`, write a final summary to the plan, and update the memory file with key findings.

## Stopping

Use `CronDelete` with the job ID shown when the loop was created. Stop early if a game-breaking bug is found (log it and alert the user).
