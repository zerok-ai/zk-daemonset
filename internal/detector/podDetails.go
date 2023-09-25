package detector

import (
	v1 "k8s.io/api/core/v1"
	"zk-daemonset/internal/models"
)

func GetPodDetails(pod *v1.Pod) (string, models.PodDetails) {
	var podDetails models.PodDetails
	for _, container := range pod.Spec.Containers {
		var containerDetails models.ContainerDetails
		containerDetails.Name = container.Name
		containerDetails.Image = container.Image
		containerDetails.ProcessExecutablePath = container.Command
		containerDetails.Ports = container.Ports
		containerDetails.ProcessCommandArgs = container.Args
		podDetails.Containers = append(podDetails.Containers, containerDetails)
	}
	podDetails.K8SNodeName = pod.Spec.NodeName
	podDetails.CreateTS = pod.GetCreationTimestamp().String()
	podDetails.K8SNamespaceName = pod.Namespace
	podDetails.K8SPodName = pod.Name
	podDetails.K8SDeploymentName = pod.OwnerReferences[0].Name
	podDetails.ServiceName = pod.OwnerReferences[0].Kind
	return pod.Status.PodIP, podDetails
}
