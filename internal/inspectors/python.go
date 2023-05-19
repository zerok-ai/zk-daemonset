package inspectors

import (
	"strings"
	types "zk-daemonset/internal/models"
)

type pythonInspector struct{}

var python = &pythonInspector{}

const pythonProcessName = "python"

func (p *pythonInspector) Inspect(process *types.ProcessDetails) (types.ProgrammingLanguage, bool) {
	if strings.Contains(process.ExeName, pythonProcessName) || strings.Contains(process.CmdLine, pythonProcessName) {
		return types.PythonProgrammingLanguage, true
	}

	return "", false
}
