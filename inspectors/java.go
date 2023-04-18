package inspectors

import (
	"fmt"
	"os"
	"strings"

	"zerok.ai/deamonset/process"
	types "zerok.ai/deamonset/types"
)

type javaInspector struct{}

var java = &javaInspector{}

const processName = "java"
const hsperfdataDirName = "hsperfdata"

func (j *javaInspector) Inspect(p *process.ProcessDetails) (types.ProgrammingLanguage, bool) {
	if strings.Contains(p.ExeName, processName) || strings.Contains(p.CmdLine, processName) {
		return types.JavaProgrammingLanguage, true
	}

	if j.isHsperfdataPresent(p.ProcessID) {
		return types.JavaProgrammingLanguage, true
	}

	return "", false
}

func (j *javaInspector) isHsperfdataPresent(pid int) bool {
	tempDir := fmt.Sprintf("/proc/%d/root/tmp/", pid)
	files, err := os.ReadDir(tempDir)
	if err != nil {
		return false
	}

	for _, f := range files {
		if f.IsDir() {
			name := f.Name()
			if strings.Contains(name, hsperfdataDirName) {
				return true
			}
		}
	}
	return false
}
