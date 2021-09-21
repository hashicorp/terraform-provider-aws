package budgets

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
)

const (
	timePeriodLayout = "2006-01-02_15:04"
)

func TimePeriodTimestampFromString(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}

	ts, err := time.Parse(timePeriodLayout, s)

	if err != nil {
		return nil, err
	}

	return aws.Time(ts), nil
}

func TimePeriodTimestampToString(ts *time.Time) string {
	if ts == nil {
		return ""
	}

	return aws.TimeValue(ts).Format(timePeriodLayout)
}

func ValidateTimePeriodTimestamp(v interface{}, k string) (ws []string, errors []error) {
	_, err := time.Parse(timePeriodLayout, v.(string))

	if err != nil {
		errors = append(errors, fmt.Errorf("%q cannot be parsed as %q: %w", k, timePeriodLayout, err))
	}

	return
}
