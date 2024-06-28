package agent

import (
	"time"
)

type PlanUpdate struct {
	TaskID string
	Goal   string
	State  string
}

func (a *Agent) GenerateDemoPlan() <-chan PlanUpdate {
	updates := make(chan PlanUpdate)
	go func() {
		defer close(updates)
		tasks := []PlanUpdate{
			{"1", "Analyze project requirements", "open"},
			{"2", "Set up development environment", "open"},
			{"3", "Design system architecture", "open"},
			{"4", "Implement core functionality", "open"},
			{"5", "Write unit tests", "open"},
			{"6", "Perform integration testing", "open"},
			{"7", "Deploy to staging environment", "open"},
			{"8", "Conduct user acceptance testing", "open"},
			{"9", "Prepare documentation", "open"},
			{"10", "Deploy to production", "open"},
		}

		for _, task := range tasks {
			time.Sleep(500 * time.Millisecond) // Simulate processing time
			updates <- task
			time.Sleep(500 * time.Millisecond) // Simulate thinking time
			updates <- PlanUpdate{task.TaskID, task.Goal, "in_progress"}
			time.Sleep(1 * time.Second) // Simulate work time
			updates <- PlanUpdate{task.TaskID, task.Goal, "completed"}
		}
	}()
	return updates
}
