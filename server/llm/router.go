package llm

import (
	"context"
	"log"
	"time"

	"github.com/divinity/core/config"
)

type Response struct {
	Content          string
	PromptTokens     int
	CompletionTokens int
	Cost             float64
	Provider         string
}

type Router struct {
	openrouter *OpenRouterClient
	breaker    *CircuitBreaker
	usage      *config.TokenUsage
	sem        chan struct{} // general concurrency semaphore
	godSem     chan struct{} // GOD-specific semaphore (smaller, prevents starvation)
}

func NewRouter(cfg *config.Config) *Router {
	maxConcurrent := cfg.Game.MaxConcurrentCalls
	if maxConcurrent <= 0 {
		maxConcurrent = 50
	}
	godMax := 3
	if godMax > maxConcurrent {
		godMax = maxConcurrent
	}

	return &Router{
		openrouter: NewOpenRouterClient(cfg.API.OpenRouterURL, cfg.API.OpenRouterKey, cfg.API.OpenRouterModel, cfg.API.TimeoutMs),
		breaker:    NewCircuitBreaker(3, 30*time.Second),
		usage:      &config.TokenUsage{},
		sem:        make(chan struct{}, maxConcurrent),
		godSem:     make(chan struct{}, godMax),
	}
}

func (r *Router) acquireSem(ctx context.Context) error {
	select {
	case r.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Router) releaseSem() {
	<-r.sem
}

func (r *Router) acquireGodSem(ctx context.Context) error {
	select {
	case r.godSem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Router) releaseGodSem() {
	<-r.godSem
}

func (r *Router) logUsage(tag string, resp *Response) {
	log.Printf("[LLM] %s via %s | prompt: %d, completion: %d | call $%.6f | total $%.6f (%d reqs, %d tokens)",
		tag, resp.Provider, resp.PromptTokens, resp.CompletionTokens,
		resp.Cost, r.usage.ActualCost, r.usage.Requests, r.usage.TotalTokens())
	const maxLog = 2000
	content := resp.Content
	if len(content) > maxLog {
		content = content[:maxLog] + "..."
	}
	if content != "" {
		log.Printf("[LLM] %s response (%d chars): %s", tag, len(resp.Content), content)
	}
}

func (r *Router) Call(ctx context.Context, systemPrompt, prompt string, model string, maxTokens int, temperature float64) (*Response, error) {
	if err := r.acquireSem(ctx); err != nil {
		return nil, err
	}
	defer r.releaseSem()

	resp, err := r.openrouter.Complete(ctx, systemPrompt, prompt, model, maxTokens, temperature)
	if err != nil {
		return nil, err
	}
	r.usage.Record(resp.PromptTokens, resp.CompletionTokens, resp.Cost)
	r.logUsage("CORE", resp)
	return resp, nil
}

func (r *Router) CallGod(ctx context.Context, systemPrompt, prompt string, model string, maxTokens int, temperature float64) (*Response, error) {
	if err := r.acquireGodSem(ctx); err != nil {
		return nil, err
	}
	defer r.releaseGodSem()

	resp, err := r.openrouter.Complete(ctx, systemPrompt, prompt, model, maxTokens, temperature)
	if err != nil {
		return nil, err
	}
	r.usage.Record(resp.PromptTokens, resp.CompletionTokens, resp.Cost)
	r.logUsage("GOD", resp)
	if resp.Content == "" {
		log.Printf("[GOD] WARNING: OpenRouter returned empty content (model=%s, prompt_tok=%d, comp_tok=%d)", model, resp.PromptTokens, resp.CompletionTokens)
	}
	return resp, nil
}

func (r *Router) CallGodWithTools(ctx context.Context, messages []Message, tools []ToolSpec, model string, maxTokens int, temperature float64) (*ToolResponse, error) {
	if err := r.acquireGodSem(ctx); err != nil {
		return nil, err
	}
	defer r.releaseGodSem()

	resp, err := r.openrouter.CompleteWithTools(ctx, messages, tools, model, maxTokens, temperature)
	if err != nil {
		return nil, err
	}
	r.usage.Record(resp.PromptTokens, resp.CompletionTokens, resp.Cost)
	r.logUsage("GOD", &resp.Response)
	if len(resp.ToolCalls) > 0 {
		for _, tc := range resp.ToolCalls {
			log.Printf("[LLM] GOD tool_call: %s(%s)", tc.Function.Name, truncateLog(tc.Function.Arguments, 300))
		}
	}
	return resp, nil
}

func truncateLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (r *Router) Usage() *config.TokenUsage {
	return r.usage
}
