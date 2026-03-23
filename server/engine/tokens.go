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

func (r *TokenRegistry) Revoke(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if npcID, ok := r.tokens[token]; ok {
		delete(r.npcToken, npcID)
	}
	delete(r.tokens, token)
}
