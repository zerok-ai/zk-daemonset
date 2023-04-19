package utils

import (
	"context"
	"fmt"
	"os"

	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"zerok.ai/deamonset/common"
)

type patchStringValue struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func LabelPod(pod *corev1.Pod, path string, value string) {
	k8sClient := getK8sClient().CoreV1()
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
	clientset := getK8sClient()
	node := getCurrentNodeName()
	fmt.Println(node)
	var pods *corev1.PodList
	if allPods {
		pods, _ = clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + node,
		})
	} else {
		pods, _ = clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + node,
			LabelSelector: common.ZkOrchStatusKey + "!=" + common.ZkOrchScanned,
		})
	}
	return pods
}

func getCurrentNodeName() string {
	return os.Getenv("MY_NODE_NAME")
}

func getK8sClient() *kubernetes.Clientset {
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
