package inspectors

import (
	types "deamonset/internal/models"
	"strings"
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
