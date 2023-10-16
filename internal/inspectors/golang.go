package inspectors

import (
	"fmt"
	zklogger "github.com/zerok-ai/zk-utils-go/logs"
	"os"
	"zk-daemonset/internal/inspectors/goversion"
	types "zk-daemonset/internal/models"
)

var golangInspectorLogTag = "golangInspector"

type golangInspector struct{}

var golang = &golangInspector{}

const golangProcessName = "go"

func (g *golangInspector) Inspect(p *types.ProcessDetails) (types.ProgrammingLanguage, bool) {
	file := fmt.Sprintf("/proc/%d/exe", p.ProcessID)
	_, err := os.Stat(file)
	if err != nil {
		zklogger.Error(golangInspectorLogTag, "could not perform os.stat: %s\n", err)
		return "", false
	}

	x, err := goversion.OpenExe(file)
	if err != nil {
		zklogger.Error(golangInspectorLogTag, "could not perform OpenExe: %s\n", err)
		return "", false
	}

	vers, _ := goversion.FindVersion(x)
	if vers == "" {
		// Not a golang app
		return "", false
	}

	return types.GoProgrammingLanguage, true
}
