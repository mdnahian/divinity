package engine

import (
	"sync"

	"github.com/google/uuid"
)

type TokenRegistry struct {
	mu       sync.RWMutex
	tokens   map[string]string // token -> npcID
	npcToken map[string]string // npcID -> token (reverse lookup)
}

func NewTokenRegistry() *TokenRegistry {
	return &TokenRegistry{
		tokens:   make(map[string]string),
		npcToken: make(map[string]string),
	}
}

func (r *TokenRegistry) Generate(npcID string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tok, ok := r.npcToken[npcID]; ok {
		return tok
	}
	token := uuid.New().String()
	r.tokens[token] = npcID
	r.npcToken[npcID] = token
	return token
}

func (r *TokenRegistry) Lookup(token string) (npcID string, ok bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	npcID, ok = r.tokens[token]
	return
}

// All returns a copy of all token->npcID mappings.
func (r *TokenRegistry) All() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]string, len(r.tokens))
	for k, v := range r.tokens {
		out[k] = v
	}
	return out
}

// Restore re-registers an existing token<->NPC mapping (used on startup to reload persisted tokens).
func (r *TokenRegistry) Restore(token, npcID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[token] = npcID
	r.npcToken[npcID] = token
}

func (r *TokenRegistry) Revoke(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if npcID, ok := r.tokens[token]; ok {
		delete(r.npcToken, npcID)
	}
	delete(r.tokens, token)
}
