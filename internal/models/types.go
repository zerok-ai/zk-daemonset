package models

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
)

type Set map[string]bool

func (s Set) Add(item string) {
	s[item] = true
}

func (s Set) Contains(item string) bool {
	return s[item]
}

type ProcessDetails struct {
	ProcessID   int                 `json:"pid"`
	ExeName     string              `json:"exe"`
	CmdLine     string              `json:"cmd"`
	Runtime     ProgrammingLanguage `json:"runtime"`
	ProcessName string              `json:"pname"`
	EnvMap      map[string]string   `json:"env"`
}

type ContainerDetails struct {
	Name                  string             `json:"container.name"`
	Image                 string             `json:"container.image"`
	ProcessExecutablePath []string           `json:"process.executable.path"`
	ProcessCommandArgs    []string           `json:"process.command_args"`
	Ports                 []v1.ContainerPort `json:"ports"`
}

type PodDetails struct {
	Containers        []ContainerDetails `json:"containers"`
	HostArch          string             `json:"host.arch"`
	HostName          string             `json:"host.name"`
	K8SDeploymentName string             `json:"k8s.deployment.name"`
	K8SNamespaceName  string             `json:"k8s.namespace.name"`
	K8SNodeName       string             `json:"k8s.node.name"`
	K8SPodName        string             `json:"k8s.pod.name"`
	ServiceName       string             `json:"service.name"`
	CreateTS          string             `json:"create_ts"`
}

type ProgrammingLanguage string

//type ContainerRuntime struct {
//	PodUID        string           `json:"uid"`
//	ContainerName string           `json:"cont"`
//	Image         string           `json:"image"`
//	ImageID       string           `json:"imageId"`
//	Process       []ProcessDetails `json:"process"`
//}

type ContainerRuntime struct {
	Image    string            `json:"image"`
	ImageID  string            `json:"imageId"`
	Language []string          `json:"language"`
	Process  string            `json:"process,omitempty"`
	Cmd      []string          `json:"cmd,omitempty"`
	EnvMap   map[string]string `json:"env"`
}

func (cr ContainerRuntime) Equals(newContainerRuntime ContainerRuntime) bool {

	if cr.Image != newContainerRuntime.Image {
		return false
	}

	if cr.ImageID != newContainerRuntime.ImageID {
		return false
	}

	if len(cr.Language) != len(newContainerRuntime.Language) {
		return false
	}

	// collect all the elements for `cr` in a set and the languages may not be in order
	langSet := make(Set)
	for _, lang := range cr.Language {
		langSet.Add(lang)
	}

	// check if all the elements of the new array are present in the old array
	for index, _ := range cr.Language {
		if !langSet.Contains(newContainerRuntime.Language[index]) {
			return false
		}
	}

	return true
}

func (cr ContainerRuntime) String() string {

	stCr := fmt.Sprintf("%s:[", cr.Image)
	for _, lang := range cr.Language {
		stCr += lang + ", "
	}
	stCr += "]"

	return stCr
}

//type ContainerRuntime struct {
//	Image    string `json:"image"`
//	ImageID  string `json:"imageId"`
//	Language string `json:"language"`
//}

type RuntimeSyncRequest struct {
	RuntimeDetails []ContainerRuntime `json:"details"`
}
