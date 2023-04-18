package main

import (
	"flag"
	"fmt"
	"strings"

	"zerok.ai/langdetector/detector"
	server "zerok.ai/langdetector/server"
	//"zerok.ai/langdetector/utils"
)

func main() {
	// node := utils.GetCurrentNodeName()
	// fmt.Println("Node is ", node)
	fmt.Println("Start lang detection.")
	result := parseArgs()
	fmt.Println("The args are ", result)
	detector.FindLang(result.PodUID, result.Containers, result.Image)
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
