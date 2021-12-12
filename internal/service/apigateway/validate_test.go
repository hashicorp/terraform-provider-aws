package apigateway

import (
	"testing"
)

func TestValidUsagePlanQuotaSettings(t *testing.T) {
	cases := []struct {
		Offset   int
		Period   string
		ErrCount int
	}{
		{
			Offset:   0,
			Period:   "DAY",
			ErrCount: 0,
		},
		{
			Offset:   -1,
			Period:   "DAY",
			ErrCount: 1,
		},
		{
			Offset:   1,
			Period:   "DAY",
			ErrCount: 1,
		},
		{
			Offset:   0,
			Period:   "WEEK",
			ErrCount: 0,
		},
		{
			Offset:   6,
			Period:   "WEEK",
			ErrCount: 0,
		},
		{
			Offset:   -1,
			Period:   "WEEK",
			ErrCount: 1,
		},
		{
			Offset:   7,
			Period:   "WEEK",
			ErrCount: 1,
		},
		{
			Offset:   0,
			Period:   "MONTH",
			ErrCount: 0,
		},
		{
			Offset:   27,
			Period:   "MONTH",
			ErrCount: 0,
		},
		{
			Offset:   -1,
			Period:   "MONTH",
			ErrCount: 1,
		},
		{
			Offset:   28,
			Period:   "MONTH",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		m := make(map[string]interface{})
		m["offset"] = tc.Offset
		m["period"] = tc.Period

		errors := validUsagePlanQuotaSettings(m)
		if len(errors) != tc.ErrCount {
			t.Fatalf("API Gateway Usage Plan Quota Settings validation failed: %v", errors)
		}
	}
}
