package types

type ContainerLanguage struct {
	ContainerName string              `json:"containerName"`
	Language      ProgrammingLanguage `json:"language"`
	ProcessName   string              `json:"processName,omitempty"`
	Image         string              `json:"image"`
}

type ProgrammingLanguage string

const (
	JavaProgrammingLanguage       ProgrammingLanguage = "java"
	PythonProgrammingLanguage     ProgrammingLanguage = "python"
	GoProgrammingLanguage         ProgrammingLanguage = "go"
	DotNetProgrammingLanguage     ProgrammingLanguage = "dotnet"
	JavascriptProgrammingLanguage ProgrammingLanguage = "javascript"
)
