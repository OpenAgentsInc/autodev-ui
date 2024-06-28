package agent

import (
	"github.com/openagentsinc/autodev/pkg/action"
	"github.com/openagentsinc/autodev/pkg/llm"
	"github.com/openagentsinc/autodev/pkg/plugin"
	"github.com/openagentsinc/autodev/pkg/state"
)

// Agent defines the interface for all agent implementations
type Agent interface {
	// Step performs one step of the agent's execution
	Step(state *state.State) action.Action

	// SearchMemory searches the agent's memory for relevant information
	SearchMemory(query string) []string

	// Reset resets the agent's state
	Reset()

	// IsComplete returns whether the agent has completed its task
	IsComplete() bool
}

// BaseAgent provides a basic implementation of the Agent interface
type BaseAgent struct {
	llm        llm.LLM
	complete   bool
	sandboxReq []plugin.PluginRequirement
}

// NewBaseAgent creates a new BaseAgent
func NewBaseAgent(l llm.LLM, req []plugin.PluginRequirement) *BaseAgent {
	return &BaseAgent{
		llm:        l,
		complete:   false,
		sandboxReq: req,
	}
}

// IsComplete returns whether the agent has completed its task
func (ba *BaseAgent) IsComplete() bool {
	return ba.complete
}

// Reset resets the agent's state
func (ba *BaseAgent) Reset() {
	ba.complete = false
}

// SearchMemory is a placeholder method that should be implemented by specific agents
func (ba *BaseAgent) SearchMemory(query string) []string {
	// This should be implemented by specific agent types
	return nil
}

// Step is a placeholder method that should be implemented by specific agents
func (ba *BaseAgent) Step(state *state.State) action.Action {
	// This should be implemented by specific agent types
	return nil
}
