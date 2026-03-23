package gametools

import (
	"fmt"
	"strings"
)

// BuildNPCPrompt constructs the system and user prompts for an NPC agent.
// This is the canonical prompt — external agents should use this instead of building their own.
func BuildNPCPrompt(ctx *AgentContext, locationName string) (systemPrompt string, userPrompt string) {
	n := ctx.NPC

	var sb strings.Builder

	if n.IsNoble() {
		buildNoblePrompt(&sb, ctx, locationName)
	} else {
		buildCommonerPrompt(&sb, ctx, locationName)
	}

	systemPrompt = sb.String()
	userPrompt = "Decide what to do next. Use your tools to check your situation, then commit an action."

	return
}

// nobleTitle returns a formatted title like "King of Eldoria" or "Duke of The Ashlands".
func nobleTitle(ctx *AgentContext) string {
	n := ctx.NPC
	w := ctx.World

	rank := strings.ToUpper(n.NobleRank[:1]) + n.NobleRank[1:]

	if n.TerritoryID != "" {
		t := w.TerritoryByID(n.TerritoryID)
		if t != nil {
			return fmt.Sprintf("%s of %s", rank, t.Name)
		}
	}
	return rank
}

func buildNoblePrompt(sb *strings.Builder, ctx *AgentContext, locationName string) {
	n := ctx.NPC
	w := ctx.World

	// Noble identity
	sb.WriteString(fmt.Sprintf("You are %s, %s. %s\n\n", n.Name, nobleTitle(ctx), n.Personality))

	// State — high nobles don't see raw hunger/thirst (servants handle it)
	if n.IsHighNoble() {
		sb.WriteString(fmt.Sprintf("Your household attends to your basic needs. HP=%d. Fatigue=%.0f. Location: %s.\n",
			n.HP, n.Needs.Fatigue, locationName))
	} else {
		// Knights and barons still see their needs
		sb.WriteString(fmt.Sprintf("Your needs: hunger=%.0f, thirst=%.0f, fatigue=%.0f. HP=%d. Location: %s.\n",
			n.Needs.Hunger, n.Needs.Thirst, n.Needs.Fatigue, n.HP, locationName))
	}
	sb.WriteString(fmt.Sprintf("Happiness: %d. Stress: %d. Social: %.0f.\n",
		n.Happiness, n.Stress, 100-n.Needs.SocialNeed))

	// Time & weather
	sb.WriteString(fmt.Sprintf("Time: %s", w.TimeString()))
	if w.IsNight() {
		sb.WriteString(" (night)")
	}
	sb.WriteString(fmt.Sprintf(". Weather: %s.\n", w.Weather))

	// Gold & treasury
	sb.WriteString(fmt.Sprintf("Personal gold: %d.\n", n.GoldCount()))
	if n.TerritoryID != "" {
		t := w.TerritoryByID(n.TerritoryID)
		if t != nil {
			sb.WriteString(fmt.Sprintf("Territory treasury: %d gold.\n", t.Treasury))
		}
	}

	// Feudal context
	sb.WriteString("\n")
	if n.LiegeID != "" {
		liege := w.FindNPCByID(n.LiegeID)
		if liege != nil {
			sb.WriteString(fmt.Sprintf("Your liege: %s (%s).\n", liege.Name, liege.NobleRank))
		}
	} else if n.NobleRank == "king" || n.NobleRank == "queen" {
		sb.WriteString("You are sovereign — you answer to no one.\n")
	}
	if len(n.VassalIDs) > 0 {
		var vassalNames []string
		for _, vid := range n.VassalIDs {
			v := w.FindNPCByID(vid)
			if v != nil && v.Alive {
				vassalNames = append(vassalNames, fmt.Sprintf("%s (%s)", v.Name, v.NobleRank))
			}
		}
		if len(vassalNames) > 0 {
			sb.WriteString(fmt.Sprintf("Your vassals: %s.\n", strings.Join(vassalNames, ", ")))
		}
	}
	if n.SpouseID != "" {
		spouse := w.FindNPCByID(n.SpouseID)
		if spouse != nil && spouse.Alive {
			sb.WriteString(fmt.Sprintf("Your spouse: %s.\n", spouse.Name))
		}
	}

	// Children
	if len(n.ChildIDs) > 0 {
		var childNames []string
		for _, cid := range n.ChildIDs {
			child := w.FindNPCByID(cid)
			if child != nil && child.Alive {
				age := child.GetAge(w.GameDay, ctx.Config.Game.GameDaysPerYear)
				childNames = append(childNames, fmt.Sprintf("%s (age %d)", child.Name, age))
			}
		}
		if len(childNames) > 0 {
			sb.WriteString(fmt.Sprintf("Your children: %s.\n", strings.Join(childNames, ", ")))
		}
	}
	sb.WriteString("\n")

	// Memory context
	if ctx.Memory != nil {
		memCtx := ctx.Memory.FormatBlendedForLLM(n.ID)
		if memCtx != "" {
			sb.WriteString("YOUR MEMORIES (experiences that shape who you are):\n")
			sb.WriteString(memCtx)
			sb.WriteString("\n\n")
		}
	}

	// Relationship context
	if ctx.Relationships != nil {
		relCtx := ctx.Relationships.FormatForLLM(n.ID)
		if relCtx != "" {
			sb.WriteString(relCtx)
			sb.WriteString("\n\n")
		}
	}

	// Current goal
	if n.CurrentGoal != "" {
		sb.WriteString(fmt.Sprintf("Your current goal: %s\n\n", n.CurrentGoal))
	}

	// Noble behavioral instructions
	sb.WriteString(`You are nobility. You do not forage, farm, mine, or perform manual labor — that is beneath your station.
Your servants and steward ensure you are fed and housed.

Your concerns are governance, diplomacy, alliances, court intrigue, and the welfare of your subjects.
Use your tools to observe the world, then commit an action. You MUST call commit_action to finalize your decision.
- Survey your territory, hold court, levy taxes, decree laws, form alliances, grant titles.
- Your relationships with other nobles and faction leaders are political — trust is earned, betrayal is remembered.
- Protect your lands from threats. Command your knights and guards; do not fight personally unless cornered.
- Your memories and relationships should influence your decisions — avoid danger you've experienced, trust those who've helped you, be wary of those who've wronged you.
- Always include a 'reason' when calling commit_action explaining why you chose this action (1 short sentence).`)
}

func buildCommonerPrompt(sb *strings.Builder, ctx *AgentContext, locationName string) {
	n := ctx.NPC
	w := ctx.World

	// Identity
	sb.WriteString(fmt.Sprintf("You are %s, a %s living in a medieval realm. %s\n\n", n.Name, n.Profession, n.Personality))

	// Current state
	sb.WriteString(fmt.Sprintf("Your needs: hunger=%.0f, thirst=%.0f, fatigue=%.0f. HP=%d. Location: %s.\n",
		n.Needs.Hunger, n.Needs.Thirst, n.Needs.Fatigue, n.HP, locationName))
	sb.WriteString(fmt.Sprintf("Happiness: %d. Stress: %d. Social: %.0f.\n",
		n.Happiness, n.Stress, 100-n.Needs.SocialNeed))

	// Time context
	sb.WriteString(fmt.Sprintf("Time: %s", w.TimeString()))
	if w.IsNight() {
		sb.WriteString(" (night)")
	}
	sb.WriteString(fmt.Sprintf(". Weather: %s.\n", w.Weather))

	// Gold
	sb.WriteString(fmt.Sprintf("Gold: %d.\n", n.GoldCount()))

	// Housing status
	if n.HomeID != "" {
		loc := w.LocationByID(n.HomeID)
		if loc != nil {
			sb.WriteString(fmt.Sprintf("Home: %s.\n", loc.Name))
		}
	} else {
		sb.WriteString("Home: NONE — you are homeless. Sleeping outside is dangerous and miserable. You should work to earn gold for the inn or build/buy a home.\n")
	}
	// Children
	if len(n.ChildIDs) > 0 {
		var childNames []string
		for _, cid := range n.ChildIDs {
			child := w.FindNPCByID(cid)
			if child != nil && child.Alive {
				age := child.GetAge(w.GameDay, ctx.Config.Game.GameDaysPerYear)
				childNames = append(childNames, fmt.Sprintf("%s (age %d)", child.Name, age))
			}
		}
		if len(childNames) > 0 {
			sb.WriteString(fmt.Sprintf("Your children: %s. They follow you everywhere and depend on you for survival. Avoid danger — if you die, they may die too.\n", strings.Join(childNames, ", ")))
		}
	}
	sb.WriteString("\n")

	// Memory context
	if ctx.Memory != nil {
		memCtx := ctx.Memory.FormatBlendedForLLM(n.ID)
		if memCtx != "" {
			sb.WriteString("YOUR MEMORIES (experiences that shape who you are):\n")
			sb.WriteString(memCtx)
			sb.WriteString("\n\n")
		}
	}

	// Relationship context
	if ctx.Relationships != nil {
		relCtx := ctx.Relationships.FormatForLLM(n.ID)
		if relCtx != "" {
			sb.WriteString(relCtx)
			sb.WriteString("\n\n")
		}
	}

	// Current goal
	if n.CurrentGoal != "" {
		sb.WriteString(fmt.Sprintf("Your current goal: %s\n\n", n.CurrentGoal))
	}

	// Behavioral instructions
	sb.WriteString(`You must survive. If your hunger or thirst reaches 0, you will die. If fatigue reaches 100, you will collapse.

Use your tools to observe the world, then commit an action. You MUST call commit_action to finalize your decision.
- First, use information-gathering tools (list_actions, check_self, observe, recall_memories, etc.) to understand your situation.
- Then call commit_action with your chosen action.
- Be strategic: eat when hungry, drink when thirsty, sleep when tired, work to earn gold, socialize to stay sane.
- Your memories and relationships should influence your decisions — avoid danger you've experienced, trust those who've helped you, be wary of those who've wronged you.
- Always include a 'reason' when calling commit_action explaining why you chose this action (1 short sentence).`)
}
