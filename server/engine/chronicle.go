package engine

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/divinity/core/world"
)

func (e *Engine) generateDailyChronicle(ctx context.Context) {
	w := e.World
	cfg := e.Config

	w.Mu.RLock()
	todayEvents := buildChronicleSource(w)
	gameDay := w.GameDay
	w.Mu.RUnlock()

	if todayEvents == "" {
		return
	}

	systemPrompt := `You are the divine chronicler of a medieval realm. You will be given a list of ACTUAL EVENTS that happened today. Your job is to retell the most noteworthy event(s) as a short literary passage (3-5 paragraphs, under 250 words) in a medieval chronicle style.

CRITICAL RULES:
- You MUST ONLY write about events that appear in the provided list. Do NOT invent, embellish, or add events that did not happen.
- Use the exact NPC names, locations, and actions from the events.
- You may add literary flavor (descriptions of weather, atmosphere) but the FACTS must come from the events list.
- Pick the 1-2 most compelling events to focus on. Not everything needs to be chronicled.
- Write in past tense, as a historian recording what happened.
- Start with a short title on its own line, then the story.
- If the events are mundane (just farming, eating, sleeping), write a brief atmospheric passage about daily life instead of forcing drama.`

	resp, err := e.Router.Call(ctx, systemPrompt, todayEvents, cfg.GodAgent.Model, 400, 0.8)
	if err != nil {
		log.Printf("[Chronicle] LLM call failed: %v", err)
		return
	}

	text := strings.TrimSpace(resp.Content)
	title, body := splitChronicleTitle(text)

	w.Mu.Lock()
	w.Chronicles = append(w.Chronicles, world.ChronicleEntry{
		Day:   gameDay,
		Title: title,
		Text:  body,
	})
	if len(w.Chronicles) > 100 {
		w.Chronicles = w.Chronicles[len(w.Chronicles)-100:]
	}
	w.Mu.Unlock()
	log.Printf("[Chronicle] Day %d chronicle generated: %s", gameDay, title)
}

func splitChronicleTitle(text string) (string, string) {
	lines := strings.SplitN(text, "\n", 2)
	if len(lines) < 2 {
		return "The Chronicle", text
	}
	title := strings.TrimSpace(lines[0])
	// Strip markdown heading markers
	title = strings.TrimLeft(title, "# ")
	if title == "" {
		title = "The Chronicle"
	}
	return title, strings.TrimSpace(lines[1])
}

func buildChronicleSource(w *world.World) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("EVENTS THAT HAPPENED TODAY (Day %d):\n", w.GameDay))

	eventCount := 0

	// Today's events from EventLog
	for _, e := range w.EventLog {
		// Filter to recent events (current day's events are at the end of the log)
		if e.Type == "npc" || e.Type == "combat" || e.Type == "god" || e.Type == "death" || e.Type == "birth" || e.Type == "coming_of_age" || e.Type == "faction" || e.Type == "world" {
			sb.WriteString(fmt.Sprintf("- [%s] %s\n", e.Time, e.Text))
			eventCount++
		}
	}

	// Cap to most recent 50 events to avoid overwhelming the LLM
	if eventCount == 0 {
		return ""
	}

	// Deaths today
	for _, n := range w.NPCs {
		if !n.Alive && n.Claimed && n.CauseOfDeath != "" {
			sb.WriteString(fmt.Sprintf("- DEATH: %s (%s) died: %s\n", n.Name, n.Profession, n.CauseOfDeath))
		}
	}

	// Notable NPC states
	sb.WriteString("\nNOTABLE NPC STATES:\n")
	for _, n := range w.AliveNPCs() {
		notable := false
		flags := ""
		if n.HP < 30 {
			flags += fmt.Sprintf(" HP:%d", n.HP)
			notable = true
		}
		if n.Needs.Hunger < 15 {
			flags += " STARVING"
			notable = true
		}
		if n.Stress > 80 {
			flags += fmt.Sprintf(" stress:%d", n.Stress)
			notable = true
		}
		if n.IsNoble() {
			flags += fmt.Sprintf(" [%s]", n.NobleRank)
			notable = true
		}
		if notable {
			sb.WriteString(fmt.Sprintf("- %s (%s, gold:%d)%s\n", n.Name, n.Profession, n.GoldCount(), flags))
		}
	}

	return sb.String()
}
