package main

import (
	"fmt"
	"time"

	"zerok.ai/deamonset/detector"
	types "zerok.ai/deamonset/types"
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
