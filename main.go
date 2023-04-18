package main

import (
	"flag"
	"fmt"
	"strings"

	"zerok.ai/deamonset/detector"
	server "zerok.ai/deamonset/server"
)

func main() {
	fmt.Println("Start lang detection.")
	result := parseArgs()
	fmt.Println("The args are ", result)
	detector.GetContainerResultsForAllPods()
	server.StartServer()
}

func parseArgs() *server.LangDetect {
	result := server.LangDetect{}
	var names string
	flag.StringVar(&result.PodUID, "pod-uid", "", "The UID of the target pod")
	flag.StringVar(&names, "container-names", "", "The container names in the target pod")
	flag.Parse()

	result.Containers = strings.Split(names, ",")

	return &result
}
