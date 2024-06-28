package observation

import (
	"encoding/json"
	"fmt"
)

// ObservationType represents the type of observation
type ObservationType string

const (
	TypeNull     ObservationType = "NULL"
	TypeBrowse   ObservationType = "BROWSE"
	TypeMessage  ObservationType = "MESSAGE"
	TypeRecall   ObservationType = "RECALL"
	TypeRun      ObservationType = "RUN"
	TypeRead     ObservationType = "READ"
	TypeWrite    ObservationType = "WRITE"
	TypeDelegate ObservationType = "DELEGATE"
	TypeError    ObservationType = "ERROR"
)

// Observation is the interface that all observation types must implement
type Observation interface {
	GetContent() string
	GetType() ObservationType
	ToDict() map[string]interface{}
	ToMemory() map[string]interface{}
	Message() string
}

// BaseObservation provides a base implementation of the Observation interface
type BaseObservation struct {
	Content string          `json:"content"`
	Type    ObservationType `json:"observation"`
}

func (bo BaseObservation) GetContent() string {
	return bo.Content
}

func (bo BaseObservation) GetType() ObservationType {
	return bo.Type
}

func (bo BaseObservation) ToDict() map[string]interface{} {
	memory := bo.ToMemory()
	memory["message"] = bo.Message()
	return memory
}

func (bo BaseObservation) ToMemory() map[string]interface{} {
	var extras map[string]interface{}
	data, _ := json.Marshal(bo)
	json.Unmarshal(data, &extras)
	delete(extras, "Content")
	delete(extras, "Type")
	return map[string]interface{}{
		"observation": bo.Type,
		"content":     bo.Content,
		"extras":      extras,
	}
}

func (bo BaseObservation) Message() string {
	return ""
}

// NullObservation represents a null observation
type NullObservation struct {
	BaseObservation
}

func NewNullObservation() *NullObservation {
	return &NullObservation{
		BaseObservation: BaseObservation{
			Type: TypeNull,
		},
	}
}

// BrowserOutputObservation represents the output of a browser
type BrowserOutputObservation struct {
	BaseObservation
	URL        string `json:"url"`
	Screenshot string `json:"screenshot"`
	StatusCode int    `json:"status_code"`
	Error      bool   `json:"error"`
}

func NewBrowserOutputObservation(content, url, screenshot string, statusCode int, error bool) *BrowserOutputObservation {
	return &BrowserOutputObservation{
		BaseObservation: BaseObservation{
			Content: content,
			Type:    TypeBrowse,
		},
		URL:        url,
		Screenshot: screenshot,
		StatusCode: statusCode,
		Error:      error,
	}
}

func (bo BrowserOutputObservation) Message() string {
	return "Visited " + bo.URL
}

// UserMessageObservation represents a message sent by the user
type UserMessageObservation struct {
	BaseObservation
	Role string `json:"role"`
}

func NewUserMessageObservation(content string) *UserMessageObservation {
	return &UserMessageObservation{
		BaseObservation: BaseObservation{
			Content: content,
			Type:    TypeMessage,
		},
		Role: "user",
	}
}

// AgentMessageObservation represents a message sent by the agent
type AgentMessageObservation struct {
	BaseObservation
	Role string `json:"role"`
}

func NewAgentMessageObservation(content string) *AgentMessageObservation {
	return &AgentMessageObservation{
		BaseObservation: BaseObservation{
			Content: content,
			Type:    TypeMessage,
		},
		Role: "assistant",
	}
}

// AgentRecallObservation represents a list of memories recalled by the agent
type AgentRecallObservation struct {
	BaseObservation
	Memories []string `json:"memories"`
	Role     string   `json:"role"`
}

func NewAgentRecallObservation(content string, memories []string) *AgentRecallObservation {
	return &AgentRecallObservation{
		BaseObservation: BaseObservation{
			Content: content,
			Type:    TypeRecall,
		},
		Memories: memories,
		Role:     "assistant",
	}
}

func (aro AgentRecallObservation) Message() string {
	return "The agent recalled memories."
}

// CmdOutputObservation represents the output of a command
type CmdOutputObservation struct {
	BaseObservation
	CommandID int    `json:"command_id"`
	Command   string `json:"command"`
	ExitCode  int    `json:"exit_code"`
}

func NewCmdOutputObservation(content string, commandID int, command string, exitCode int) *CmdOutputObservation {
	return &CmdOutputObservation{
		BaseObservation: BaseObservation{
			Content: content,
			Type:    TypeRun,
		},
		CommandID: commandID,
		Command:   command,
		ExitCode:  exitCode,
	}
}

func (co CmdOutputObservation) Error() bool {
	return co.ExitCode != 0
}

func (co CmdOutputObservation) Message() string {
	return fmt.Sprintf("Command `%s` executed with exit code %d.", co.Command, co.ExitCode)
}

// FileReadObservation represents the content of a file
type FileReadObservation struct {
	BaseObservation
	Path string `json:"path"`
}

func NewFileReadObservation(content, path string) *FileReadObservation {
	return &FileReadObservation{
		BaseObservation: BaseObservation{
			Content: content,
			Type:    TypeRead,
		},
		Path: path,
	}
}

func (fro FileReadObservation) Message() string {
	return fmt.Sprintf("I read the file %s.", fro.Path)
}

// FileWriteObservation represents a file write operation
type FileWriteObservation struct {
	BaseObservation
	Path string `json:"path"`
}

func NewFileWriteObservation(content, path string) *FileWriteObservation {
	return &FileWriteObservation{
		BaseObservation: BaseObservation{
			Content: content,
			Type:    TypeWrite,
		},
		Path: path,
	}
}

func (fwo FileWriteObservation) Message() string {
	return fmt.Sprintf("I wrote to the file %s.", fwo.Path)
}

// AgentDelegateObservation represents a delegate observation
type AgentDelegateObservation struct {
	BaseObservation
	Outputs map[string]interface{} `json:"outputs"`
}

func NewAgentDelegateObservation(content string, outputs map[string]interface{}) *AgentDelegateObservation {
	return &AgentDelegateObservation{
		BaseObservation: BaseObservation{
			Content: content,
			Type:    TypeDelegate,
		},
		Outputs: outputs,
	}
}

// AgentErrorObservation represents an error encountered by the agent
type AgentErrorObservation struct {
	BaseObservation
}

func NewAgentErrorObservation(content string) *AgentErrorObservation {
	return &AgentErrorObservation{
		BaseObservation: BaseObservation{
			Content: content,
			Type:    TypeError,
		},
	}
}

func (aeo AgentErrorObservation) Message() string {
	return "Oops. Something went wrong: " + aeo.Content
}

// ObservationFromDict creates an Observation from a dictionary
func ObservationFromDict(observationMap map[string]interface{}) (Observation, error) {
	observationType, ok := observationMap["observation"].(string)
	if !ok {
		return nil, fmt.Errorf("'observation' key is not found or not a string in %v", observationMap)
	}

	content, _ := observationMap["content"].(string)
	extras, _ := observationMap["extras"].(map[string]interface{})

	switch ObservationType(observationType) {
	case TypeNull:
		return NewNullObservation(), nil
	case TypeBrowse:
		url, _ := extras["url"].(string)
		screenshot, _ := extras["screenshot"].(string)
		statusCode, _ := extras["status_code"].(float64)
		error, _ := extras["error"].(bool)
		return NewBrowserOutputObservation(content, url, screenshot, int(statusCode), error), nil
	case TypeMessage:
		role, _ := extras["role"].(string)
		if role == "user" {
			return NewUserMessageObservation(content), nil
		}
		return NewAgentMessageObservation(content), nil
	case TypeRecall:
		memories, _ := extras["memories"].([]string)
		return NewAgentRecallObservation(content, memories), nil
	case TypeRun:
		commandID, _ := extras["command_id"].(float64)
		command, _ := extras["command"].(string)
		exitCode, _ := extras["exit_code"].(float64)
		return NewCmdOutputObservation(content, int(commandID), command, int(exitCode)), nil
	case TypeRead:
		path, _ := extras["path"].(string)
		return NewFileReadObservation(content, path), nil
	case TypeWrite:
		path, _ := extras["path"].(string)
		return NewFileWriteObservation(content, path), nil
	case TypeDelegate:
		outputs, _ := extras["outputs"].(map[string]interface{})
		return NewAgentDelegateObservation(content, outputs), nil
	case TypeError:
		return NewAgentErrorObservation(content), nil
	default:
		return nil, fmt.Errorf("unknown observation type: %s", observationType)
	}
}
