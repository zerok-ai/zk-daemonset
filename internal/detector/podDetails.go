package detector

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"strings"
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
		podDetails.Spec.Containers = append(podDetails.Spec.Containers, containerDetails)
	}
	// metadata
	podDetails.Metadata.CreateTS = pod.GetCreationTimestamp().String()
	podDetails.Metadata.PodName = fmt.Sprintf("%v/%v", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)
	podDetails.Metadata.PodId = string(pod.ObjectMeta.UID)
	podDetails.Metadata.WorkloadName = pod.ObjectMeta.OwnerReferences[0].Name
	podDetails.Metadata.WorkloadKind = pod.ObjectMeta.OwnerReferences[0].Kind
	podDetails.Metadata.ServiceName = strings.TrimSuffix(pod.ObjectMeta.GenerateName, "-")
	// Spec
	podDetails.Spec.ServiceAccountName = pod.Spec.ServiceAccountName
	podDetails.Spec.NodeName = pod.Spec.NodeName
	// Status
	podDetails.Status.Phase = string(pod.Status.Phase)
	podDetails.Status.PodIP = pod.Status.PodIP

	return pod.Status.PodIP, podDetails
}
