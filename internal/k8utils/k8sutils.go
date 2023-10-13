package k8utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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

func LabelPod(pod *corev1.Pod, path string, value string) error {
	k8sClientSet, err := GetK8sClientSet()
	if err != nil {
		return err
	}
	k8sClient := k8sClientSet.CoreV1()
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
	return nil
}

func GetPodsInCurrentNode() (*corev1.PodList, error) {
	clientSet, err := GetK8sClientSet()
	if err != nil {
		return nil, err
	}
	node := GetCurrentNodeName()

	pods, _ := clientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + node,
	})

	return pods, nil
}

func GetServicesInCluster() (*corev1.ServiceList, error) {
	clientSet, err := GetK8sClientSet()
	if err != nil {
		return nil, err
	}
	services, _ := clientSet.CoreV1().Services("").List(context.Background(), metav1.ListOptions{})
	return services, nil
}

func GetCurrentNodeName() string {
	return os.Getenv("MY_NODE_NAME")
}

func GetK8sClientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	return kubernetes.NewForConfig(config)
	//if err != nil {
	//	return nil, err
	//}
	//
	//return clientset, nil
}
