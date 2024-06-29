package server

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/openagentsinc/autodev/agent"
	"github.com/openagentsinc/autodev/config"
	"github.com/openagentsinc/autodev/llm"
)

func HandleSubmitMessage(cfg *config.Config, myAgent *agent.Agent) echo.HandlerFunc {
	return func(c echo.Context) error {
		message := c.FormValue("message")

		conversationHistory := myAgent.GetConversationHistory()
		conversationHistory = append(conversationHistory, llm.Message{
			Role:    "user",
			Content: message,
		})

		response, err := cfg.LLM.GenerateResponse(conversationHistory, 1024)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		conversationHistory = append(conversationHistory, llm.Message{
			Role:    "assistant",
			Content: response,
		})

		myAgent.SetConversationHistory(conversationHistory)

		newTask := &agent.Task{
			ID:    fmt.Sprintf("%d", len(myAgent.GetPlan().Tasks)+1),
			Goal:  response,
			State: "open",
		}
		myAgent.GetPlan().Tasks = append(myAgent.GetPlan().Tasks, newTask)

		planHTML := generatePlanHTML(myAgent.GetPlan())

		htmlResponse := fmt.Sprintf(`
			<div class="bg-zinc-800 rounded p-3 inline-block">%s</div>
			<div id="plan-display" hx-swap-oob="true">%s</div>
		`, response, planHTML)

		return c.HTML(http.StatusOK, htmlResponse)
	}
}

func generatePlanHTML(plan *agent.Plan) string {
	// Implement this function to generate the HTML for the plan
	// You can move the existing implementation from setup.go if it exists
	return "" // Placeholder return
}
