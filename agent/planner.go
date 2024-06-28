package agent

// Task represents a single task or subtask in the plan
type Task struct {
	ID       string
	Goal     string
	Subtasks []*Task
	State    string
}

// Plan represents the overall plan with a main goal and tasks
type Plan struct {
	MainGoal string
	Tasks    []*Task
}

// NewPlan creates a new Plan with a main goal and initial tasks
func NewPlan(mainGoal string, initialTasks []*Task) *Plan {
	return &Plan{
		MainGoal: mainGoal,
		Tasks:    initialTasks,
	}
}

// Agent represents our AI agent with planning capabilities
type Agent struct {
	CurrentPlan *Plan
}

// NewAgent creates a new Agent with an initial plan
func NewAgent(plan *Plan) *Agent {
	return &Agent{
		CurrentPlan: plan,
	}
}

// GetPlan returns the current plan of the agent
func (a *Agent) GetPlan() *Plan {
	return a.CurrentPlan
}

func (a *Agent) ResetPlan() {
	a.CurrentPlan = NewPlan(a.CurrentPlan.MainGoal, []*Task{})
}
