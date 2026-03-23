package llm

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"
)

var thinkTagRe = regexp.MustCompile(`(?s)<think>.*?</think>`)
var rawJSONRe = regexp.MustCompile(`(?s)(\{[^{}]*"action"[^{}]*\})`)

func StripThinkTags(s string) string {
	return strings.TrimSpace(thinkTagRe.ReplaceAllString(s, ""))
}

// extractOuterJSON finds the outermost {...} by brace counting, handling nested objects/arrays.
func extractOuterJSON(s string) string {
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return ""
	}
	depth := 0
	inString := false
	escape := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if escape {
			escape = false
			continue
		}
		if c == '\\' && inString {
			escape = true
			continue
		}
		if c == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if c == '{' || c == '[' {
			depth++
		} else if c == '}' || c == ']' {
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}

func ExtractJSON(raw string) (map[string]interface{}, error) {
	cleaned := StripThinkTags(raw)

	// Try direct unmarshal of the full cleaned text
	trimmed := strings.TrimSpace(cleaned)
	if strings.HasPrefix(trimmed, "{") {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(trimmed), &result); err == nil {
			return result, nil
		}
	}

	// Extract outermost JSON object by brace counting
	candidate := extractOuterJSON(cleaned)
	if candidate != "" {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(candidate), &result); err == nil {
			return result, nil
		}
	}

	// Flat action JSON fallback
	if m := rawJSONRe.FindStringSubmatch(cleaned); len(m) > 1 {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(m[1]), &result); err == nil {
			return result, nil
		}
	}

	log.Printf("[LLM] Failed to extract JSON from response (len=%d): %.500s", len(raw), raw)
	return nil, ErrNoJSON
}

type ParsedAction struct {
	Action   string
	Target   string
	Dialogue string
	Goal     string
	Location string
}

func ParseActionResponse(raw string) (*ParsedAction, error) {
	data, err := ExtractJSON(raw)
	if err != nil {
		return nil, err
	}

	action, _ := data["action"].(string)
	if action == "" {
		return nil, ErrNoAction
	}

	target, _ := data["target"].(string)
	if strings.EqualFold(strings.TrimSpace(target), "null") {
		target = ""
	}
	dialogue, _ := data["dialogue"].(string)
	goal, _ := data["goal"].(string)
	location, _ := data["location"].(string)

	return &ParsedAction{
		Action:   strings.TrimSpace(action),
		Target:   strings.TrimSpace(target),
		Dialogue: strings.TrimSpace(dialogue),
		Goal:     strings.TrimSpace(goal),
		Location: strings.TrimSpace(location),
	}, nil
}

type ParsedGodAction struct {
	Action  string
	Details map[string]interface{}
}

func ParseGodActions(raw string) ([]ParsedGodAction, string, error) {
	data, err := ExtractJSON(raw)
	if err != nil {
		return nil, "", err
	}

	analysis, _ := data["analysis"].(string)
	if analysis != "" {
		log.Printf("[GOD] Analysis: %s", analysis)
	}

	var actions []ParsedGodAction

	if arr, ok := data["actions"].([]interface{}); ok && len(arr) > 0 {
		seen := make(map[string]bool)
		for _, item := range arr {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			a, _ := m["action"].(string)
			if a == "" || seen[a] {
				continue
			}
			seen[a] = true
			actions = append(actions, ParsedGodAction{Action: a, Details: m})
		}
	}

	if len(actions) == 0 {
		action, _ := data["action"].(string)
		if action == "" {
			return nil, analysis, ErrNoAction
		}
		actions = append(actions, ParsedGodAction{Action: action, Details: data})
	}

	return actions, analysis, nil
}

var (
	ErrNoJSON   = &ParseError{Msg: "no valid JSON found in response"}
	ErrNoAction = &ParseError{Msg: "no action field in response"}
)

type ParseError struct {
	Msg string
}

func (e *ParseError) Error() string { return e.Msg }
