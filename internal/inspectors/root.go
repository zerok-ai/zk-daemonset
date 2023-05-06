package inspectors

import (
	types "zerok-deamonset/internal/models"
)

type inspector interface {
	Inspect(process *types.ProcessDetails) (types.ProgrammingLanguage, bool)
}

var inspectorsList = []inspector{java, nodeJs, python}

func DetectLanguageOfAllProcesses(processes []types.ProcessDetails) []string {
	processName := ""
	results := []string{}
	for _, p := range processes {
		p.ProcessName = processName
		p.Runtime = types.UnknownLanguage
		for _, i := range inspectorsList {
			inspectionResult, detected := i.Inspect(&p)
			if detected {
				if inspectionResult == types.GoProgrammingLanguage {
					processName = p.ExeName
				}
				p.ProcessName = processName
				p.Runtime = inspectionResult
				break
			}
		}
		results = append(results, string(p.Runtime))
	}
	return results
}
