package inspectors

import (
	"strings"

	"zerok.ai/deamonset/process"
	types "zerok.ai/deamonset/types"
)

type pythonInspector struct{}

var python = &pythonInspector{}

const pythonProcessName = "python"

func (p *pythonInspector) Inspect(process *process.ProcessDetails) (types.ProgrammingLanguage, bool) {
	if strings.Contains(process.ExeName, pythonProcessName) || strings.Contains(process.CmdLine, pythonProcessName) {
		return types.PythonProgrammingLanguage, true
	}

	return "", false
}
