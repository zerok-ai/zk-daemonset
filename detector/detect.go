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
	zkclient "zerok.ai/deamonset/zkclient"
)

func ReScanPods(injectorClient *zkclient.InjectorClient) {
	containerResults := GetContainerResultsForAllPods()
	injectorClient.ContainerResults = append(injectorClient.ContainerResults, containerResults...)
	injectorClient.SyncDataWithInjector()
}

func GetContainerResultsForAllPods() []types.ContainerRuntime {
	podList := utils.GetPodsInCurrentNode()
	fmt.Println("Pods are ", podList)
	containerResults := []types.ContainerRuntime{}
	for _, pod := range podList.Items {
		temp := FindLang(string(pod.UID), pod.Spec.Containers, "")
		containerResults = append(containerResults, temp...)
	}
	return containerResults
}

func FindLang(targetPodUID string, targetContainers []v1.Container, image string) []types.ContainerRuntime {
	var containerResults []types.ContainerRuntime
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

		processes = inspectors.DetectLanguage(processes)

		containerResults = append(containerResults, types.ContainerRuntime{
			ContainerName: containerName,
			Image:         image,
			Process:       processes,
			PodUID:        targetPodUID,
		})

	}
	fmt.Println(containerResults)
	return containerResults
}

func convertProcessDetailsToString(process types.ProcessDetails) string {
	return strconv.Itoa(process.ProcessID) + "," + process.CmdLine + "," + process.ExeName
}
