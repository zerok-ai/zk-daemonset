package inspectors

import (
	types "zk-daemonset/internal/models"
)

type inspector interface {
	Inspect(process *types.ProcessDetails) (types.ProgrammingLanguage, bool)
}

var inspectorsList = []inspector{java, nodeJs, python, golang}

func DetectLanguageOfAllProcesses(processes []types.ProcessDetails) ([]string, string) {
	results := []string{}
	processName := ""
	for _, p := range processes {
		p.Runtime = types.UnknownLanguage
		for _, i := range inspectorsList {
			inspectionResult, detected := i.Inspect(&p)
			if detected {
				p.Runtime = inspectionResult
				results = append(results, string(p.Runtime))
				if inspectionResult == types.GoProgrammingLanguage {
					processName = p.ExeName
				}
				break
			}
		}
	}
	return results, processName
}
