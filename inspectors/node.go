package inspectors

import (
	"strings"

	"zerok.ai/langdetector/process"
	types "zerok.ai/langdetector/types"
)

type nodejsInspector struct{}

var nodeJs = &nodejsInspector{}

const nodeProcessName = "node"

func (n *nodejsInspector) Inspect(process *process.ProcessDetails) (types.ProgrammingLanguage, bool) {
	if strings.Contains(process.ExeName, nodeProcessName) || strings.Contains(process.CmdLine, nodeProcessName) {
		return types.JavascriptProgrammingLanguage, true
	}

	return "", false
}
