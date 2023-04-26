package main

import (
	types "zerok.ai/deamonset/common"
	"zerok.ai/deamonset/detector"
	zkclient "zerok.ai/deamonset/zkclient"
)

func main() {
	injectorClient := &zkclient.InjectorClient{
		ContainerResults: []types.ContainerRuntime{},
	}

	detector.ScanExistingPods(injectorClient)
	detector.AddWatcherToPods(injectorClient)
}
