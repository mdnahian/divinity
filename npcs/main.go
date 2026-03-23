package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

const tokenFile = ".npc-token"
const tokenDir = ".npc-tokens"

func main() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)

	coreURL := envStr("CORE_URL", "http://localhost:8080")
	llmEndpoint := envStr("LLM_ENDPOINT", "http://localhost:11434/v1/chat/completions")
	llmAPIKey := envStr("LLM_API_KEY", "")
	llmModel := envStr("LLM_MODEL", "qwen3:8b")
	llmMaxTokens := envInt("LLM_MAX_TOKENS", 4096)
	llmTemperature := envFloat("LLM_TEMPERATURE", 0.7)
	llmTimeoutMs := envInt("LLM_TIMEOUT_MS", 120000)
	agentCount := envInt("NPC_AGENT_COUNT", 1)

	log.Printf("[NPC Agent] Core server: %s", coreURL)
	log.Printf("[NPC Agent] LLM endpoint: %s (model: %s)", llmEndpoint, llmModel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Printf("Shutting down all agents...")
		cancel()
	}()

	if agentCount <= 1 {
		// Single-agent mode (backwards compatible)
		runSingleAgent(ctx, coreURL, llmEndpoint, llmAPIKey, llmModel, llmMaxTokens, llmTemperature, llmTimeoutMs)
	} else {
		// Multi-agent mode: spawn and run N agents concurrently
		runMultiAgent(ctx, agentCount, coreURL, llmEndpoint, llmAPIKey, llmModel, llmMaxTokens, llmTemperature, llmTimeoutMs)
	}
}

func runSingleAgent(ctx context.Context, coreURL, llmEndpoint, llmAPIKey, llmModel string, llmMaxTokens int, llmTemperature float64, llmTimeoutMs int) {
	client := NewCoreClient(coreURL)
	llmClient := NewLLMClient(llmEndpoint, llmAPIKey, llmModel, llmMaxTokens, llmTemperature, llmTimeoutMs)

	var name, profession, personality string

	savedToken := loadToken()
	if savedToken != "" {
		log.Printf("[NPC Agent] Found saved token, attempting to resume...")
		client.SetToken(savedToken)
		resp, err := client.Resume()
		if err != nil {
			log.Printf("[NPC Agent] Resume failed: %v — will spawn fresh", err)
			savedToken = ""
		} else {
			name = resp.Name
			profession = resp.Profession
			log.Printf("[NPC Agent] Resumed as %s the %s (alive=%v)", name, profession, resp.Alive)
			if !resp.Alive {
				log.Printf("[NPC Agent] NPC is dead. Exiting.")
				os.Exit(0)
			}
		}
	}

	if savedToken == "" {
		log.Printf("[NPC Agent] Spawning new NPC...")
		resp, err := client.Spawn()
		if err != nil {
			log.Fatalf("[NPC Agent] Spawn failed: %v", err)
		}
		client.SetToken(resp.Token)
		saveToken(resp.Token)
		name = resp.Name
		profession = resp.Profession
		personality = resp.Personality
		log.Printf("[NPC Agent] Spawned as %s the %s at %s (HP: %d)", name, profession, resp.Location, resp.HP)
	}

	log.SetPrefix("[" + name + "] ")
	agent := NewAgent(client, llmClient, name, profession, personality)
	log.Printf("Agent loop starting")
	if err := agent.Run(ctx); err != nil {
		log.Printf("Agent stopped: %v", err)
	}
}

func runMultiAgent(ctx context.Context, count int, coreURL, llmEndpoint, llmAPIKey, llmModel string, llmMaxTokens int, llmTemperature float64, llmTimeoutMs int) {
	log.Printf("[Multi-Agent] Spawning %d NPC agents...", count)

	os.MkdirAll(tokenDir, 0755)

	// Load saved tokens from previous runs
	savedTokens := loadMultiTokens()

	type agentEntry struct {
		token       string
		name        string
		profession  string
		personality string
	}

	var agents []agentEntry

	// Resume previously saved agents
	for _, token := range savedTokens {
		client := NewCoreClient(coreURL)
		client.SetToken(token)
		resp, err := client.Resume()
		if err != nil {
			log.Printf("[Multi-Agent] Resume failed for token %s...: %v — skipping", token[:8], err)
			continue
		}
		if !resp.Alive {
			log.Printf("[Multi-Agent] %s is dead — skipping", resp.Name)
			continue
		}
		log.Printf("[Multi-Agent] Resumed: %s the %s", resp.Name, resp.Profession)
		agents = append(agents, agentEntry{token: token, name: resp.Name, profession: resp.Profession})
	}

	// Spawn more until we reach the target count
	for len(agents) < count {
		client := NewCoreClient(coreURL)
		resp, err := client.Spawn()
		if err != nil {
			log.Printf("[Multi-Agent] Spawn failed (have %d/%d): %v — stopping spawn", len(agents), count, err)
			break
		}
		log.Printf("[Multi-Agent] Spawned: %s the %s at %s", resp.Name, resp.Profession, resp.Location)
		agents = append(agents, agentEntry{
			token:       resp.Token,
			name:        resp.Name,
			profession:  resp.Profession,
			personality: resp.Personality,
		})
	}

	if len(agents) == 0 {
		log.Fatalf("[Multi-Agent] No agents spawned. Exiting.")
	}

	// Save all tokens for next restart
	var allTokens []string
	for _, a := range agents {
		allTokens = append(allTokens, a.token)
	}
	saveMultiTokensList(allTokens)

	log.Printf("[Multi-Agent] Running %d agents concurrently", len(agents))

	var wg sync.WaitGroup
	for _, entry := range agents {
		wg.Add(1)
		go func(e agentEntry) {
			defer wg.Done()
			client := NewCoreClient(coreURL)
			client.SetToken(e.token)
			llmClient := NewLLMClient(llmEndpoint, llmAPIKey, llmModel, llmMaxTokens, llmTemperature, llmTimeoutMs)
			agent := NewAgent(client, llmClient, e.name, e.profession, e.personality)
			log.Printf("[%s] Agent loop starting", e.name)
			if err := agent.Run(ctx); err != nil {
				log.Printf("[%s] Agent stopped: %v", e.name, err)
			}
		}(entry)
	}
	wg.Wait()
	log.Printf("[Multi-Agent] All agents stopped")
}

func loadMultiTokens() []string {
	data, err := os.ReadFile(filepath.Join(tokenDir, "tokens.json"))
	if err != nil {
		return nil
	}
	var tokens []string
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil
	}
	return tokens
}

func saveMultiTokensList(tokens []string) {
	data, err := json.Marshal(tokens)
	if err != nil {
		log.Printf("[Multi-Agent] WARNING: failed to marshal tokens: %v", err)
		return
	}
	if err := os.WriteFile(filepath.Join(tokenDir, "tokens.json"), data, 0600); err != nil {
		log.Printf("[Multi-Agent] WARNING: failed to save tokens: %v", err)
	}
}

func loadToken() string {
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func saveToken(token string) {
	if err := os.WriteFile(tokenFile, []byte(token+"\n"), 0600); err != nil {
		log.Printf("[NPC Agent] WARNING: failed to save token: %v", err)
	}
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
