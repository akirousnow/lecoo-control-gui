package lecoo

const fanDutyMax = 220

type Capabilities struct {
	FanDutyMax int `json:"fanDutyMax"`
}

func (s *Service) Capabilities() Capabilities {
	return Capabilities{FanDutyMax: fanDutyMax}
}
