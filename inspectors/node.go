package inspectors

import (
	"strings"

	types "zerok.ai/deamonset/common"
)

type nodejsInspector struct{}

var nodeJs = &nodejsInspector{}

const nodeProcessName = "node"

func (n *nodejsInspector) Inspect(process *types.ProcessDetails) (types.ProgrammingLanguage, bool) {
	if strings.Contains(process.ExeName, nodeProcessName) || strings.Contains(process.CmdLine, nodeProcessName) {
		return types.JavascriptProgrammingLanguage, true
	}

	return "", false
}
