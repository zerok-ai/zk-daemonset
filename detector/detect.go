package detector

import (
	"fmt"
	"log"

	"strconv"

	inspectors "zerok.ai/langdetector/inspectors"
	"zerok.ai/langdetector/process"
	types "zerok.ai/langdetector/types"
)

func FindLang(targetPodUID string, targetContainers []string, image string) {
	var containerResults []types.ContainerLanguage
	for _, containerName := range targetContainers {
		processes, err := process.FindProcessInContainer(targetPodUID, containerName)
		if err != nil {
			log.Fatalf("could not find processes, error: %s\n", err)
		}

		for i := 0; i < len(processes); i++ {
			fmt.Println(convertProcessDetailsToString(processes[i]))
		}

		processResults, processName := inspectors.DetectLanguage(processes)
		log.Printf("detection result: %s\n", processResults)

		if len(processResults) > 0 {
			containerResults = append(containerResults, types.ContainerLanguage{
				ContainerName: containerName,
				Language:      processResults[0],
				ProcessName:   processName,
				Image:         image,
			})
		}
	}
	fmt.Println(containerResults)
}

func convertProcessDetailsToString(process process.ProcessDetails) string {
	return strconv.Itoa(process.ProcessID) + "," + process.CmdLine + "," + process.ExeName
}
