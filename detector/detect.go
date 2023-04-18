package detector

import (
	"fmt"
	"log"

	"strconv"

	v1 "k8s.io/api/core/v1"
	inspectors "zerok.ai/deamonset/inspectors"
	"zerok.ai/deamonset/process"
	types "zerok.ai/deamonset/types"
	utils "zerok.ai/deamonset/utils"
)

func GetContainerResultsForAllPods() {
	podList := utils.GetPodsInCurrentNode()
	fmt.Println("Pods are ", podList)
	for _, pod := range podList.Items {
		FindLang(string(pod.UID), pod.Spec.Containers, "")
	}
}

func FindLang(targetPodUID string, targetContainers []v1.Container, image string) {
	var containerResults []types.ContainerLanguage
	fmt.Println("Container Names is ", targetContainers)
	for _, container := range targetContainers {
		containerName := container.Name
		fmt.Println("Container Name is ", containerName)
		processes, err := process.FindProcessInContainer(targetPodUID, containerName)
		fmt.Println("processes are ", processes)
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
