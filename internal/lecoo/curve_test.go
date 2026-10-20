package lecoo

import "testing"

func TestDefaultCurveInterpolates(t *testing.T) {
	curve, err := DefaultCurve().Normalized()
	if err != nil {
		t.Fatalf("Normalized: %v", err)
	}
	cases := []struct {
		temp int
		duty int
	}{
		{30, 30},
		{40, 30},
		{47, 35},
		{55, 40},
		{60, 48},
		{65, 55},
		{85, 95},
		{90, 95},
		{95, 100},
		{99, 100},
	}
	for _, tc := range cases {
		if got := curve.DutyFor(tc.temp); got != tc.duty {
			t.Errorf("DutyFor(%d) = %d, want %d", tc.temp, got, tc.duty)
		}
	}
}

func TestCurveRespectsMinAndMax(t *testing.T) {
	curve := DefaultCurve()
	curve.MinDuty = 45
	curve.MaxDuty = 80
	normalized, err := curve.Normalized()
	if err != nil {
		t.Fatalf("Normalized: %v", err)
	}
	if got := normalized.DutyFor(20); got != 45 {
		t.Errorf("cold duty = %d, want floor 45", got)
	}
	if got := normalized.DutyFor(84); got != 80 {
		t.Errorf("hot duty = %d, want ceiling 80", got)
	}
	if got := normalized.DutyFor(99); got != 100 {
		t.Errorf("panic duty = %d, want 100", got)
	}
}

func TestCurveSortsAndDedupes(t *testing.T) {
	curve := FanCurve{
		Target:     "both",
		Points:     []CurvePoint{{TempC: 80, Duty: 90}, {TempC: 50, Duty: 40}, {TempC: 80, Duty: 95}},
		MinDuty:    20,
		MaxDuty:    100,
		Hysteresis: 8,
		PanicTemp:  100,
	}
	normalized, err := curve.Normalized()
	if err != nil {
		t.Fatalf("Normalized: %v", err)
	}
	if len(normalized.Points) != 2 {
		t.Fatalf("points = %d, want 2 after dedupe", len(normalized.Points))
	}
	if normalized.Points[0].TempC != 50 || normalized.Points[1].TempC != 80 {
		t.Errorf("points not sorted: %+v", normalized.Points)
	}
	if normalized.Points[1].Duty != 95 {
		t.Errorf("duplicate temp kept %d, want last value 95", normalized.Points[1].Duty)
	}
}

func TestCurveRejectsBadInput(t *testing.T) {
	if _, err := (FanCurve{Target: "both", Points: []CurvePoint{{TempC: 50, Duty: 40}}}).Normalized(); err == nil {
		t.Error("expected an error for a single point")
	}
	if _, err := (FanCurve{Target: "lcd", Points: DefaultCurve().Points}).Normalized(); err == nil {
		t.Error("expected an error for an unknown target")
	}
	if _, err := (FanCurve{Target: "cpu", Points: []CurvePoint{{TempC: 50, Duty: 40}, {TempC: 50, Duty: 60}}}).Normalized(); err == nil {
		t.Error("expected an error when all points share one temperature")
	}
}

func TestPWMConversion(t *testing.T) {
	curve, err := DefaultCurve().Normalized()
	if err != nil {
		t.Fatalf("Normalized: %v", err)
	}
	if got := curve.PWMDuty(65); got != 121 {
		t.Errorf("PWMDuty(65) = %d, want 121", got)
	}
	if got := curve.PWMDuty(95); got != 220 {
		t.Errorf("PWMDuty(95) = %d, want 220", got)
	}
	if got := curve.PWMDuty(20); got != 66 {
		t.Errorf("PWMDuty(20) = %d, want 66", got)
	}
}
