package detector

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	types "zk-daemonset/internal/models"

	"github.com/fntlnz/mountinfo"
)

func FindProcessInContainer(podUID string, containerName string) ([]types.ProcessDetails, error) {
	procFile, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}

	var result []types.ProcessDetails
	for {
		dirs, err := procFile.Readdir(15)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, di := range dirs {
			if !di.IsDir() {
				continue
			}

			dname := di.Name()
			if dname[0] < '0' || dname[0] > '9' {
				continue
			}

			pid, err := strconv.Atoi(dname)
			if err != nil {
				return nil, err
			}

			mountInfos, err := mountinfo.GetMountInfo(path.Join("/proc", dname, "mountinfo"))
			if err != nil {
				continue
			}

			for _, mountInfo := range mountInfos {
				root := mountInfo.Root
				if strings.Contains(root, fmt.Sprintf("%s/containers/%s", podUID, containerName)) {
					exeName, err := os.Readlink(path.Join("/proc", dname, "exe"))
					if err != nil {
						exeName = ""
					}

					cmdLine, err := os.ReadFile(path.Join("/proc", dname, "cmdline"))
					var cmd string
					if err != nil {
						cmd = ""
					} else {
						cmdLine = bytes.ReplaceAll(cmdLine, []byte{0}, []byte(" "))
						cmd = string(cmdLine)
					}

					envMap, err := parseProcEnviron(dname)

					result = append(result, types.ProcessDetails{
						ProcessID: pid,
						ExeName:   exeName,
						CmdLine:   cmd,
						EnvMap:    envMap,
					})
				}
			}
		}
	}

	return result, nil
}

func parseProcEnviron(pid string) (map[string]string, error) {
	envMap := make(map[string]string)

	data, err := os.ReadFile(fmt.Sprintf("/proc/%s/environ", pid))
	if err != nil {
		return nil, err
	}

	envEntries := strings.Split(string(data), "\000")

	for _, entry := range envEntries {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			if strings.HasPrefix(key, "OTEL") || strings.HasPrefix(key, "JAVA") {
				envMap[key] = parts[1]
			}
		}
	}

	return envMap, nil
}
