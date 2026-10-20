package lecoo

import (
	"errors"
	"math"
	"sort"
)

type CurvePoint struct {
	TempC int `json:"tempC"`
	Duty  int `json:"duty"`
}

type FanCurve struct {
	Enabled    bool         `json:"enabled"`
	Target     string       `json:"target"`
	Points     []CurvePoint `json:"points"`
	MinDuty    int          `json:"minDuty"`
	MaxDuty    int          `json:"maxDuty"`
	Hysteresis int          `json:"hysteresis"`
	PanicTemp  int          `json:"panicTemp"`
}

type CurveState struct {
	Curve       FanCurve `json:"curve"`
	TempC       *int     `json:"tempC"`
	Duty        int      `json:"duty"`
	AppliedPWM  int      `json:"appliedPwm"`
	LastWriteAt int64    `json:"lastWriteAt"`
	LastError   string   `json:"lastError"`
}

func DefaultCurve() FanCurve {
	return FanCurve{
		Enabled:    false,
		Target:     "both",
		Points:     []CurvePoint{{TempC: 40, Duty: 30}, {TempC: 55, Duty: 40}, {TempC: 65, Duty: 55}, {TempC: 75, Duty: 75}, {TempC: 85, Duty: 95}},
		MinDuty:    25,
		MaxDuty:    100,
		Hysteresis: 8,
		PanicTemp:  95,
	}
}

func clampInt(value int, low int, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func (c FanCurve) Normalized() (FanCurve, error) {
	out := c
	if out.Target != "cpu" && out.Target != "gpu" && out.Target != "both" {
		return out, errors.New("未知的风扇目标")
	}
	if len(out.Points) < 2 {
		return out, errors.New("温度曲线至少需要两个节点")
	}
	points := make([]CurvePoint, 0, len(out.Points))
	for _, point := range out.Points {
		points = append(points, CurvePoint{
			TempC: clampInt(point.TempC, 20, 110),
			Duty:  clampInt(point.Duty, 0, 100),
		})
	}
	sort.SliceStable(points, func(i int, j int) bool {
		return points[i].TempC < points[j].TempC
	})
	deduped := points[:1]
	for _, point := range points[1:] {
		if point.TempC == deduped[len(deduped)-1].TempC {
			deduped[len(deduped)-1] = point
			continue
		}
		deduped = append(deduped, point)
	}
	if len(deduped) < 2 {
		return out, errors.New("温度曲线至少需要两个不同温度的节点")
	}
	out.Points = deduped
	out.MinDuty = clampInt(out.MinDuty, 0, 100)
	out.MaxDuty = clampInt(out.MaxDuty, 0, 100)
	if out.MaxDuty < out.MinDuty {
		out.MaxDuty = out.MinDuty
	}
	out.Hysteresis = clampInt(out.Hysteresis, 1, 64)
	out.PanicTemp = clampInt(out.PanicTemp, 60, 110)
	for index, point := range out.Points {
		out.Points[index].Duty = clampInt(point.Duty, out.MinDuty, out.MaxDuty)
	}
	return out, nil
}

func (c FanCurve) DutyFor(tempC int) int {
	if tempC >= c.PanicTemp {
		return 100
	}
	if len(c.Points) == 0 {
		return c.MinDuty
	}
	first := c.Points[0]
	if tempC <= first.TempC {
		return clampInt(first.Duty, c.MinDuty, c.MaxDuty)
	}
	last := c.Points[len(c.Points)-1]
	if tempC >= last.TempC {
		return clampInt(last.Duty, c.MinDuty, c.MaxDuty)
	}
	for index := 1; index < len(c.Points); index++ {
		high := c.Points[index]
		if tempC > high.TempC {
			continue
		}
		low := c.Points[index-1]
		span := high.TempC - low.TempC
		if span <= 0 {
			return clampInt(high.Duty, c.MinDuty, c.MaxDuty)
		}
		ratio := float64(tempC-low.TempC) / float64(span)
		value := float64(low.Duty) + ratio*float64(high.Duty-low.Duty)
		return clampInt(int(math.Round(value)), c.MinDuty, c.MaxDuty)
	}
	return clampInt(last.Duty, c.MinDuty, c.MaxDuty)
}

func (c FanCurve) PWMDuty(tempC int) int {
	percent := c.DutyFor(tempC)
	return int(math.Round(float64(percent) * fanDutyMax / 100))
}
