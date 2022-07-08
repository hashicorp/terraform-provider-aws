package budgets

import "testing"

func TestTimePeriodSecondsFromString(t *testing.T) {
	seconds, err := TimePeriodSecondsFromString("2020-03-01_00:00")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	want := "1583020800"
	if seconds != want {
		t.Errorf("got %s, expected %s", seconds, want)
	}
}

func TestTimePeriodSecondsToString(t *testing.T) {
	ts, err := TimePeriodSecondsToString("1583020800")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	want := "2020-03-01_00:00"
	if ts != want {
		t.Errorf("got %s, expected %s", ts, want)
	}
}
