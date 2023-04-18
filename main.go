package main

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"zerok.ai/deamonset/detector"
	types "zerok.ai/deamonset/types"
	zkclient "zerok.ai/deamonset/zkclient"
)

func main() {
	podUID := "05b1065e-5aa7-4dbd-b851-185a0c8ea073"
	container := v1.Container{Name: "server"}
	containers := []v1.Container{container}
	fmt.Println("Testing temp pod.")
	detector.FindLang(podUID, containers, "")
	injectorClient := &zkclient.InjectorClient{
		ContainerResults: []types.ContainerRuntime{},
	}
	for {
		fmt.Println("Rescaning pods")
		detector.ReScanPods(injectorClient)
		time.Sleep(5 * time.Second)
	}
}
