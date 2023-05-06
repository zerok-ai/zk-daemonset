package k8utils

import (
	"context"
	"fmt"
	"os"
	"zerok-deamonset/internal/models"

	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func LabelPod(pod *corev1.Pod, path string, value string) {
	k8sClient := GetK8sClientSet().CoreV1()
	payload := []patchStringValue{{
		Op:    "replace",
		Path:  path,
		Value: value,
	}}
	payloadBytes, _ := json.Marshal(payload)
	_, updateErr := k8sClient.Pods(pod.GetNamespace()).Patch(context.Background(), pod.GetName(), types.JSONPatchType, payloadBytes, metav1.PatchOptions{})
	if updateErr == nil {
		logMessage := fmt.Sprintf("Pod %s labeled successfully for Path %s and Value %s.", pod.GetName(), path, value)
		fmt.Println(logMessage)
	} else {
		fmt.Println(updateErr)
	}
}

func GetPodsInCurrentNode(allPods bool) *corev1.PodList {
	clientSet := GetK8sClientSet()
	node := GetCurrentNodeName()
	fmt.Println(node)
	var pods *corev1.PodList
	if allPods {
		pods, _ = clientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + node,
		})
	} else {
		pods, _ = clientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + node,
			LabelSelector: models.ZkOrchStatusKey + "!=" + models.ZkOrchScanned,
		})
	}
	return pods
}

func GetCurrentNodeName() string {
	return os.Getenv("MY_NODE_NAME")
}

func GetK8sClientSet() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}
