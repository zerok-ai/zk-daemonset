package main

import (
	"deamonset/internal/detector"
	types "deamonset/internal/models"
	"deamonset/pkg/zkclient/controller"
)

func main() {
	injectorClient := &controller.InjectorClient{
		ContainerResults: []types.ContainerRuntime{},
	}

	detector.ScanExistingPods(injectorClient)
	detector.AddWatcherToPods(injectorClient)
}
