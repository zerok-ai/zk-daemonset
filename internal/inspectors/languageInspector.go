package inspectors

import (
	types "zk-daemonset/internal/models"
)

type inspector interface {
	Inspect(process *types.ProcessDetails) (types.ProgrammingLanguage, bool)
}

var inspectorsList = []inspector{java, nodeJs, python}

func DetectLanguageOfAllProcesses(processes []types.ProcessDetails) []string {
	results := []string{}
	for _, p := range processes {
		p.Runtime = types.UnknownLanguage
		for _, i := range inspectorsList {
			inspectionResult, detected := i.Inspect(&p)
			if detected {
				p.Runtime = inspectionResult
				break
			}
		}
		results = append(results, string(p.Runtime))
	}
	return results
}
