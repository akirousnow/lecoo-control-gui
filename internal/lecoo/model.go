package lecoo

type FanMode struct {
	Kind string `json:"kind"`
	PWM  int    `json:"pwm"`
}

type PowerProfile struct {
	Kind string `json:"kind"`
	Raw  string `json:"raw"`
}

type ChargeTarget struct {
	Kind   string `json:"kind"`
	Preset string `json:"preset"`
	Min    *int   `json:"min"`
	Max    *int   `json:"max"`
	Raw    string `json:"raw"`
}

type Settings struct {
	TelemetryEnabled  bool         `json:"telemetryEnabled"`
	TelemetryClientID string       `json:"telemetryClientId"`
	Power             PowerProfile `json:"power"`
	Charge            ChargeTarget `json:"charge"`
	FanCPU            FanMode      `json:"fanCpu"`
	FanGPU            FanMode      `json:"fanGpu"`
	ReadAt            int64        `json:"readAt"`
}

type Info struct {
	Controller        string `json:"controller"`
	ControllerVersion int    `json:"controllerVersion"`
	HramOffset        string `json:"hramOffset"`
	DaemonVersion     string `json:"daemonVersion"`
	CliVersion        string `json:"cliVersion"`
}

type Sample struct {
	At       int64 `json:"at"`
	CPUTempC *int  `json:"cpuTempC"`
	SysTempC *int  `json:"sysTempC"`
	CPURpm   *int  `json:"cpuRpm"`
	GPURpm   *int  `json:"gpuRpm"`
}

type Battery struct {
	Target    string `json:"target"`
	Effective string `json:"effective"`
	Percent   *int   `json:"percent"`
	At        int64  `json:"at"`
}

type CliError struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

func (e *CliError) Error() string {
	return e.Message
}

type MonitorState struct {
	Running bool   `json:"running"`
	RateMs  int    `json:"rateMs"`
	Error   string `json:"error"`
}

type Health struct {
	CliPath  string `json:"cliPath"`
	CliFound bool   `json:"cliFound"`
	DaemonUp bool   `json:"daemonUp"`
	Version  string `json:"version"`
	Message  string `json:"message"`
}
