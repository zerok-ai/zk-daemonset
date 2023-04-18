package inspectors

import (
	"zerok.ai/deamonset/process"
	types "zerok.ai/deamonset/types"
)

type inspector interface {
	Inspect(process *process.ProcessDetails) (types.ProgrammingLanguage, bool)
}

var inspectorsList = []inspector{java, nodeJs, python}

func DetectLanguage(processes []process.ProcessDetails) ([]types.ProgrammingLanguage, string) {
	var result []types.ProgrammingLanguage
	processName := ""
	for _, p := range processes {
		for _, i := range inspectorsList {
			inspectionResult, detected := i.Inspect(&p)
			if detected {
				result = append(result, inspectionResult)
				if inspectionResult == types.GoProgrammingLanguage {
					processName = p.ExeName
				}
				break
			}
		}
	}

	return result, processName
}
