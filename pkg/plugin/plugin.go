package plugin

import (
	"fmt"
	"path/filepath"
	"strings"
)

// PluginRequirement represents a requirement for a plugin
type PluginRequirement struct {
	Name           string
	HostSrc        string
	SandboxDest    string
	BashScriptPath string
}

// SandboxProtocol defines the interface for sandbox operations
type SandboxProtocol interface {
	Execute(cmd string) (int, string)
	CopyTo(hostSrc, sandboxDest string, recursive bool)
}

// PluginMixin provides plugin support for Sandbox
type PluginMixin struct {
	Sandbox SandboxProtocol
}

// InitPlugins initializes plugins in the sandbox
func (pm *PluginMixin) InitPlugins(requirements []PluginRequirement) error {
	for _, req := range requirements {
		// Simulate copying files
		pm.Sandbox.CopyTo(req.HostSrc, req.SandboxDest, true)
		// logger.Info(fmt.Sprintf("Copied files from [%s] to [%s] inside sandbox.", req.HostSrc, req.SandboxDest))

		// Execute the bash script
		absPathToBashScript := filepath.Join(req.SandboxDest, req.BashScriptPath)
		// logger.Info(fmt.Sprintf("Initializing plugin [%s] by executing [%s] in the sandbox.", req.Name, absPathToBashScript))

		exitCode, output := pm.Sandbox.Execute(absPathToBashScript)
		if exitCode != 0 {
			return fmt.Errorf("failed to initialize plugin %s with exit code %d and output %s", req.Name, exitCode, output)
		}
		// logger.Info(fmt.Sprintf("Plugin %s initialized successfully:\n%s", req.Name, output))
	}

	if len(requirements) > 0 {
		exitCode, output := pm.Sandbox.Execute("source ~/.bashrc")
		if exitCode != 0 {
			return fmt.Errorf("failed to source ~/.bashrc with exit code %d and output %s", exitCode, output)
		}
		// logger.Info("Sourced ~/.bashrc successfully")
	}

	return nil
}

// MockSandbox is a mock implementation of SandboxProtocol for testing
type MockSandbox struct{}

// Execute simulates command execution in the sandbox
func (ms *MockSandbox) Execute(cmd string) (int, string) {
	// Simulate successful execution
	return 0, fmt.Sprintf("Executed command: %s", cmd)
}

// CopyTo simulates file copying in the sandbox
func (ms *MockSandbox) CopyTo(hostSrc, sandboxDest string, recursive bool) {
	// Simulate successful copy
	// logger.Info(fmt.Sprintf("Copied %s to %s (recursive: %v)", hostSrc, sandboxDest, recursive))
}

// NewPluginMixin creates a new PluginMixin with a MockSandbox
func NewPluginMixin() *PluginMixin {
	return &PluginMixin{
		Sandbox: &MockSandbox{},
	}
}

// JupyterRequirement is a placeholder for Jupyter plugin requirement
var JupyterRequirement = PluginRequirement{
	Name:           "jupyter",
	HostSrc:        "/path/to/jupyter/files",
	SandboxDest:    "/sandbox/jupyter",
	BashScriptPath: "setup.sh",
}

// GetPluginRequirements returns a list of plugin requirements
func GetPluginRequirements(names []string) []PluginRequirement {
	requirements := make([]PluginRequirement, 0)
	for _, name := range names {
		switch strings.ToLower(name) {
		case "jupyter":
			requirements = append(requirements, JupyterRequirement)
		// Add more cases for other plugins as needed
		default:
			// logger.Warn(fmt.Sprintf("Unknown plugin: %s", name))
		}
	}
	return requirements
}
