package models

type ProcessDetails struct {
	ProcessID   int                 `json:"pid"`
	ExeName     string              `json:"exe"`
	CmdLine     string              `json:"cmd"`
	Runtime     ProgrammingLanguage `json:"runtime"`
	ProcessName string              `json:"pname"`
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
	Image    string   `json:"image"`
	ImageID  string   `json:"imageId"`
	Language []string `json:"language"`
}

func (cr ContainerRuntime) Equals(newContainerRuntime ContainerRuntime) bool {

	if cr.Image != newContainerRuntime.Image {
		return false
	}

	if cr.ImageID != newContainerRuntime.ImageID {
		return false
	}

	for index, _ := range cr.Language {
		if cr.Language[index] != newContainerRuntime.Language[index] {
			return false
		}
	}

	return true
}

//type ContainerRuntime struct {
//	Image    string `json:"image"`
//	ImageID  string `json:"imageId"`
//	Language string `json:"language"`
//}

type RuntimeSyncRequest struct {
	RuntimeDetails []ContainerRuntime `json:"details"`
}
