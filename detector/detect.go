package detector

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	common "zerok.ai/deamonset/common"
	inspectors "zerok.ai/deamonset/inspectors"
	"zerok.ai/deamonset/process"
	utils "zerok.ai/deamonset/utils"
	zkclient "zerok.ai/deamonset/zkclient"
)

func ReScanPods(injectorClient *zkclient.InjectorClient) {
	containerResults := GetContainerResultsForAllPods(false)
	injectorClient.ContainerResults = append(injectorClient.ContainerResults, containerResults...)
	injectorClient.SyncDataWithInjector()
}

func GetContainerResultsForAllPods(allPods bool) []common.ContainerRuntime {
	podList := utils.GetPodsInCurrentNode(allPods)
	fmt.Println("Pods are ", podList)
	containerResults := []common.ContainerRuntime{}
	for _, pod := range podList.Items {
		temp := FindLang(&pod, pod.Spec.Containers)
		containerResults = append(containerResults, temp...)
	}
	return containerResults
}

func FindLang(pod *v1.Pod, targetContainers []v1.Container) []common.ContainerRuntime {
	targetPodUID := string(pod.UID)
	var containerResults []common.ContainerRuntime
	for _, container := range targetContainers {
		containerName := container.Name
		processes, err := process.FindProcessInContainer(targetPodUID, containerName)
		if err != nil {
			fmt.Println("caught error while getting processes ", processes)
			continue
		}
		processes = inspectors.DetectLanguage(processes)
		containerResults = append(containerResults, common.ContainerRuntime{
			ContainerName: containerName,
			Image:         container.Image,
			Process:       processes,
			PodUID:        targetPodUID,
		})

		utils.LabelPod(pod, common.ZkOrchStatusPath, common.ZkOrchScanned)

	}
	fmt.Println(containerResults)
	return containerResults
}
