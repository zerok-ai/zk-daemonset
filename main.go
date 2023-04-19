package main

import (
	"fmt"
	"time"

	types "zerok.ai/deamonset/common"
	"zerok.ai/deamonset/detector"
	zkclient "zerok.ai/deamonset/zkclient"
)

func main() {
	injectorClient := &zkclient.InjectorClient{
		ContainerResults: []types.ContainerRuntime{},
	}
	for {
		fmt.Println("Rescaning pods")
		detector.ReScanPods(injectorClient)
		time.Sleep(5 * time.Second)
	}
}
