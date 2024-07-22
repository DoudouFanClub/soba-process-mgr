package processmanager

type ExecDetails struct {
	LaunchCommand string `json:"launch-command"`
	KeepAlive     bool   `json:"keep-alive"`
	ExtraDelay    int    `json:"extra-delay"`
}

type PackageExecDetails struct {
	ParentProcess    ExecDetails          `json:"parent-process"`
	ChildProcesseses []PackageExecDetails `json:"child-processes"`
}