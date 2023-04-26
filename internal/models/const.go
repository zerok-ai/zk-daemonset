package models

const ZkOrchStatusKey = "zk-status"
const ZkOrchStatusPath = "/metadata/labels/" + ZkOrchStatusKey
const ZkOrchScanned = "orchestrated"

const (
	JavaProgrammingLanguage       ProgrammingLanguage = "java"
	PythonProgrammingLanguage     ProgrammingLanguage = "python"
	GoProgrammingLanguage         ProgrammingLanguage = "go"
	DotNetProgrammingLanguage     ProgrammingLanguage = "dotnet"
	JavascriptProgrammingLanguage ProgrammingLanguage = "javascript"
	UnknownLanguage               ProgrammingLanguage = "unknown"
)
