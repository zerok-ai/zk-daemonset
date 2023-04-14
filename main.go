package main

import (
	server "zerok.ai/langdetector/server"
)

var targetPodUID = ""
var targetContainers = []string{}

func main() {
	//fmt.Println("abc")
	server.StartServer()

}
