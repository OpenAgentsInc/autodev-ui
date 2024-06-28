package state

import (
	"github.com/openagentsinc/autodev/pkg/action"
	"github.com/openagentsinc/autodev/pkg/observation"
	"github.com/openagentsinc/autodev/pkg/plan"
)

// State represents the current state of the agent's execution
type State struct {
	Plan                  *plan.Plan
	Iteration             int
	NumOfChars            int
	BackgroundCommandsObs []observation.CmdOutputObservation
	History               []HistoryEntry
	UpdatedInfo           []HistoryEntry
	Inputs                map[string]interface{}
	Outputs               map[string]interface{}
}

// HistoryEntry represents a single entry in the agent's history
type HistoryEntry struct {
	Action      action.Action
	Observation observation.Observation
}

// NewState creates a new State instance
func NewState(p *plan.Plan) *State {
	return &State{
		Plan:                  p,
		Iteration:             0,
		NumOfChars:            0,
		BackgroundCommandsObs: make([]observation.CmdOutputObservation, 0),
		History:               make([]HistoryEntry, 0),
		UpdatedInfo:           make([]HistoryEntry, 0),
		Inputs:                make(map[string]interface{}),
		Outputs:               make(map[string]interface{}),
	}
}
