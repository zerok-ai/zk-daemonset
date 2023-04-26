package models

type ProcessDetails struct {
	ProcessID   int                 `json:"pid"`
	ExeName     string              `json:"exe"`
	CmdLine     string              `json:"cmd"`
	Runtime     ProgrammingLanguage `json:"runtime"`
	ProcessName string              `json:"pname"`
}

type ProgrammingLanguage string

type ContainerRuntime struct {
	PodUID        string           `json:"uid"`
	ContainerName string           `json:"cont"`
	Image         string           `json:"image"`
	ImageID       string           `json:"imageId"`
	Process       []ProcessDetails `json:"process"`
}

type RuntimeSyncRequest struct {
	RuntimeDetails []ContainerRuntime `json:"details"`
}
