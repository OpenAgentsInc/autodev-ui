package action

import (
	"encoding/json"
	"fmt"

	"github.com/openagentsinc/autodev/pkg/observation"
)

// ActionType represents the type of action
type ActionType string

const (
	TypeNull       ActionType = "NULL"
	TypeRun        ActionType = "RUN"
	TypeKill       ActionType = "KILL"
	TypeBrowse     ActionType = "BROWSE"
	TypeRead       ActionType = "READ"
	TypeWrite      ActionType = "WRITE"
	TypeRecall     ActionType = "RECALL"
	TypeThink      ActionType = "THINK"
	TypeEcho       ActionType = "ECHO"
	TypeSummarize  ActionType = "SUMMARIZE"
	TypeFinish     ActionType = "FINISH"
	TypeDelegate   ActionType = "DELEGATE"
	TypeAddTask    ActionType = "ADD_TASK"
	TypeModifyTask ActionType = "MODIFY_TASK"
)

// Action is the interface that all action types must implement
type Action interface {
	Run(controller AgentController) (observation.Observation, error)
	ToMemory() map[string]interface{}
	ToDict() map[string]interface{}
	IsExecutable() bool
	Message() string
	Type() ActionType
}

// BaseAction provides a base implementation of the Action interface
type BaseAction struct {
	ActionType ActionType `json:"action"`
}

func (ba BaseAction) ToMemory() map[string]interface{} {
	return map[string]interface{}{
		"action": ba.ActionType,
		"args":   map[string]interface{}{},
	}
}

func (ba BaseAction) ToDict() map[string]interface{} {
	dict := ba.ToMemory()
	dict["message"] = ba.Message()
	return dict
}

func (ba BaseAction) Message() string {
	return ""
}

func (ba BaseAction) Type() ActionType {
	return ba.ActionType
}

// NullAction represents an action that does nothing
type NullAction struct {
	BaseAction
}

func NewNullAction() *NullAction {
	return &NullAction{BaseAction{ActionType: TypeNull}}
}

func (na NullAction) Run(controller AgentController) (observation.Observation, error) {
	return observation.NewNullObservation(), nil
}

func (na NullAction) IsExecutable() bool {
	return false
}

func (na NullAction) Message() string {
	return "No action"
}

// CmdRunAction represents an action to run a shell command
type CmdRunAction struct {
	BaseAction
	Command    string `json:"command"`
	Background bool   `json:"background"`
}

func NewCmdRunAction(command string, background bool) *CmdRunAction {
	return &CmdRunAction{
		BaseAction: BaseAction{ActionType: TypeRun},
		Command:    command,
		Background: background,
	}
}

func (cra CmdRunAction) Run(controller AgentController) (observation.Observation, error) {
	return controller.ActionManager().RunCommand(cra.Command, cra.Background)
}

func (cra CmdRunAction) IsExecutable() bool {
	return true
}

func (cra CmdRunAction) Message() string {
	return fmt.Sprintf("Running command: %s", cra.Command)
}

// CmdKillAction represents an action to kill a running command
type CmdKillAction struct {
	BaseAction
	ID int `json:"id"`
}

func NewCmdKillAction(id int) *CmdKillAction {
	return &CmdKillAction{
		BaseAction: BaseAction{ActionType: TypeKill},
		ID:         id,
	}
}

func (cka CmdKillAction) Run(controller AgentController) (observation.Observation, error) {
	return controller.ActionManager().KillCommand(cka.ID)
}

func (cka CmdKillAction) IsExecutable() bool {
	return true
}

func (cka CmdKillAction) Message() string {
	return fmt.Sprintf("Killing command: %d", cka.ID)
}

// BrowseURLAction represents an action to browse a URL
type BrowseURLAction struct {
	BaseAction
	URL string `json:"url"`
}

func NewBrowseURLAction(url string) *BrowseURLAction {
	return &BrowseURLAction{
		BaseAction: BaseAction{ActionType: TypeBrowse},
		URL:        url,
	}
}

func (bua BrowseURLAction) Run(controller AgentController) (observation.Observation, error) {
	// Implement browser functionality here
	return nil, fmt.Errorf("BrowseURLAction not implemented")
}

func (bua BrowseURLAction) IsExecutable() bool {
	return true
}

func (bua BrowseURLAction) Message() string {
	return fmt.Sprintf("Browsing URL: %s", bua.URL)
}

// FileReadAction represents an action to read a file
type FileReadAction struct {
	BaseAction
	Path  string `json:"path"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

func NewFileReadAction(path string, start, end int) *FileReadAction {
	return &FileReadAction{
		BaseAction: BaseAction{ActionType: TypeRead},
		Path:       path,
		Start:      start,
		End:        end,
	}
}

func (fra FileReadAction) Run(controller AgentController) (observation.Observation, error) {
	// Implement file read functionality here
	return nil, fmt.Errorf("FileReadAction not implemented")
}

func (fra FileReadAction) IsExecutable() bool {
	return true
}

func (fra FileReadAction) Message() string {
	return fmt.Sprintf("Reading file: %s", fra.Path)
}

// FileWriteAction represents an action to write to a file
type FileWriteAction struct {
	BaseAction
	Path    string `json:"path"`
	Content string `json:"content"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
}

func NewFileWriteAction(path, content string, start, end int) *FileWriteAction {
	return &FileWriteAction{
		BaseAction: BaseAction{ActionType: TypeWrite},
		Path:       path,
		Content:    content,
		Start:      start,
		End:        end,
	}
}

func (fwa FileWriteAction) Run(controller AgentController) (observation.Observation, error) {
	// Implement file write functionality here
	return nil, fmt.Errorf("FileWriteAction not implemented")
}

func (fwa FileWriteAction) IsExecutable() bool {
	return true
}

func (fwa FileWriteAction) Message() string {
	return fmt.Sprintf("Writing file: %s", fwa.Path)
}

// AgentRecallAction represents an action for the agent to recall information
type AgentRecallAction struct {
	BaseAction
	Query string `json:"query"`
}

func NewAgentRecallAction(query string) *AgentRecallAction {
	return &AgentRecallAction{
		BaseAction: BaseAction{ActionType: TypeRecall},
		Query:      query,
	}
}

func (ara AgentRecallAction) Run(controller AgentController) (observation.Observation, error) {
	memories := controller.Agent().SearchMemory(ara.Query)
	return observation.NewAgentRecallObservation("Recalling memories...", memories), nil
}

func (ara AgentRecallAction) IsExecutable() bool {
	return true
}

func (ara AgentRecallAction) Message() string {
	return fmt.Sprintf("Recalling: %s", ara.Query)
}

// AgentThinkAction represents an action for the agent to think
type AgentThinkAction struct {
	BaseAction
	Thought string `json:"thought"`
}

func NewAgentThinkAction(thought string) *AgentThinkAction {
	return &AgentThinkAction{
		BaseAction: BaseAction{ActionType: TypeThink},
		Thought:    thought,
	}
}

func (ata AgentThinkAction) Run(controller AgentController) (observation.Observation, error) {
	return nil, fmt.Errorf("AgentThinkAction is not executable")
}

func (ata AgentThinkAction) IsExecutable() bool {
	return false
}

func (ata AgentThinkAction) Message() string {
	return ata.Thought
}

// Implement other action types (AgentEchoAction, AgentSummarizeAction, AgentFinishAction, AgentDelegateAction, AddTaskAction, ModifyTaskAction) similarly...

// AgentController interface (to be implemented elsewhere)
type AgentController interface {
	ActionManager() ActionManager
	Agent() Agent
}

// ActionManager interface (to be implemented elsewhere)
type ActionManager interface {
	RunCommand(command string, background bool) (observation.Observation, error)
	KillCommand(id int) (observation.Observation, error)
}

// Agent interface (to be implemented elsewhere)
type Agent interface {
	SearchMemory(query string) []string
}

// ActionFromDict creates an Action from a dictionary
func ActionFromDict(actionMap map[string]interface{}) (Action, error) {
	actionType, ok := actionMap["action"].(string)
	if !ok {
		return nil, fmt.Errorf("'action' key is not found or not a string in %v", actionMap)
	}

	args, ok := actionMap["args"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("'args' key is not found or not a map in %v", actionMap)
	}

	switch ActionType(actionType) {
	case TypeNull:
		return NewNullAction(), nil
	case TypeRun:
		command, _ := args["command"].(string)
		background, _ := args["background"].(bool)
		return NewCmdRunAction(command, background), nil
	case TypeKill:
		id, _ := args["id"].(float64)
		return NewCmdKillAction(int(id)), nil
	case TypeBrowse:
		url, _ := args["url"].(string)
		return NewBrowseURLAction(url), nil
	case TypeRead:
		path, _ := args["path"].(string)
		start, _ := args["start"].(float64)
		end, _ := args["end"].(float64)
		return NewFileReadAction(path, int(start), int(end)), nil
	case TypeWrite:
		path, _ := args["path"].(string)
		content, _ := args["content"].(string)
		start, _ := args["start"].(float64)
		end, _ := args["end"].(float64)
		return NewFileWriteAction(path, content, int(start), int(end)), nil
	case TypeRecall:
		query, _ := args["query"].(string)
		return NewAgentRecallAction(query), nil
	case TypeThink:
		thought, _ := args["thought"].(string)
		return NewAgentThinkAction(thought), nil
	// Implement other action types...
	default:
		return nil, fmt.Errorf("unknown action type: %s", actionType)
	}
}
