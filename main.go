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
	fmt.Println("Scanning all podss")
	detector.ScanAllPods(injectorClient)
	time.Sleep(5 * time.Second)
	for {
		fmt.Println("Rescaning pods")
		detector.ReScanPods(injectorClient)
		time.Sleep(5 * time.Second)
	}
}
