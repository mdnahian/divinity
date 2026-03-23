package nakama

import (
	"github.com/divinity/core/engine"
)

type Module struct {
	Engine *engine.Engine
}

func NewModule(eng *engine.Engine) *Module {
	return &Module{Engine: eng}
}
