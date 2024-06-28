package plan

import (
	"fmt"
	"strconv"
	"strings"
)

// Task states
const (
	OpenState       = "open"
	CompletedState  = "completed"
	AbandonedState  = "abandoned"
	InProgressState = "in_progress"
	VerifiedState   = "verified"
)

var validStates = []string{OpenState, CompletedState, AbandonedState, InProgressState, VerifiedState}

// Task represents a single task in the plan
type Task struct {
	ID       string
	Goal     string
	Parent   *Task
	Subtasks []*Task
	State    string
}

// NewTask creates a new Task
func NewTask(parent *Task, goal string, state string, subtasks []*Task) *Task {
	t := &Task{
		Goal:     goal,
		Parent:   parent,
		State:    OpenState,
		Subtasks: make([]*Task, 0),
	}

	if parent == nil {
		t.ID = "0"
	} else {
		t.ID = parent.ID + "." + strconv.Itoa(len(parent.Subtasks))
	}

	if state != "" {
		t.State = state
	}

	for _, subtask := range subtasks {
		t.Subtasks = append(t.Subtasks, subtask)
	}

	return t
}

// String returns a string representation of the task and its subtasks
func (t *Task) String() string {
	return t.toString("")
}

func (t *Task) toString(indent string) string {
	emoji := ""
	switch t.State {
	case VerifiedState:
		emoji = "✅"
	case CompletedState:
		emoji = "🟢"
	case AbandonedState:
		emoji = "❌"
	case InProgressState:
		emoji = "💪"
	case OpenState:
		emoji = "🔵"
	}

	result := fmt.Sprintf("%s%s %s %s\n", indent, emoji, t.ID, t.Goal)
	for _, subtask := range t.Subtasks {
		result += subtask.toString(indent + "    ")
	}
	return result
}

// ToDict returns a dictionary representation of the task
func (t *Task) ToDict() map[string]interface{} {
	subtasks := make([]map[string]interface{}, len(t.Subtasks))
	for i, subtask := range t.Subtasks {
		subtasks[i] = subtask.ToDict()
	}

	return map[string]interface{}{
		"id":       t.ID,
		"goal":     t.Goal,
		"state":    t.State,
		"subtasks": subtasks,
	}
}

// SetState sets the state of the task and its subtasks
func (t *Task) SetState(state string) error {
	if !isValidState(state) {
		return fmt.Errorf("invalid state: %s", state)
	}

	t.State = state

	if state == CompletedState || state == AbandonedState || state == VerifiedState {
		for _, subtask := range t.Subtasks {
			if subtask.State != AbandonedState {
				if err := subtask.SetState(state); err != nil {
					return err
				}
			}
		}
	} else if state == InProgressState {
		if t.Parent != nil {
			if err := t.Parent.SetState(state); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetCurrentTask retrieves the current task in progress
func (t *Task) GetCurrentTask() *Task {
	for _, subtask := range t.Subtasks {
		if subtask.State == InProgressState {
			return subtask.GetCurrentTask()
		}
	}
	if t.State == InProgressState {
		return t
	}
	return nil
}

// Plan represents a plan consisting of tasks
type Plan struct {
	MainGoal string
	Task     *Task
}

// NewPlan creates a new Plan
func NewPlan(task string) *Plan {
	return &Plan{
		MainGoal: task,
		Task:     NewTask(nil, task, "", nil),
	}
}

// String returns a string representation of the plan
func (p *Plan) String() string {
	return p.Task.String()
}

// GetTaskByID retrieves a task by its ID
func (p *Plan) GetTaskByID(id string) (*Task, error) {
	parts := strings.Split(id, ".")
	if parts[0] != "0" {
		return nil, fmt.Errorf("invalid task id, must start with 0: %s", id)
	}

	task := p.Task
	for _, part := range parts[1:] {
		index, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid task id, non-integer: %s", id)
		}
		if index >= len(task.Subtasks) {
			return nil, fmt.Errorf("task does not exist: %s", id)
		}
		task = task.Subtasks[index]
	}
	return task, nil
}

// AddSubtask adds a subtask to a parent task
func (p *Plan) AddSubtask(parentID, goal string, subtasks []*Task) error {
	parent, err := p.GetTaskByID(parentID)
	if err != nil {
		return err
	}

	child := NewTask(parent, goal, "", subtasks)
	parent.Subtasks = append(parent.Subtasks, child)
	return nil
}

// SetSubtaskState sets the state of a subtask
func (p *Plan) SetSubtaskState(id, state string) error {
	task, err := p.GetTaskByID(id)
	if err != nil {
		return err
	}
	return task.SetState(state)
}

// GetCurrentTask retrieves the current task in progress
func (p *Plan) GetCurrentTask() *Task {
	return p.Task.GetCurrentTask()
}

func isValidState(state string) bool {
	for _, s := range validStates {
		if s == state {
			return true
		}
	}
	return false
}
