package budgets

import (
	"fmt"
	"strconv"
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

func TimePeriodSecondsFromString(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	ts, err := time.Parse(timePeriodLayout, s)

	if err != nil {
		return "", err
	}

	return strconv.FormatInt(aws.Time(ts).Unix(), 10), nil
}

func TimePeriodSecondsToString(s string) (string, error) {
	startTime, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return "", err
	}

	startTime = startTime * 1000

	return aws.SecondsTimeValue(&startTime).UTC().Format(timePeriodLayout), nil
}

func ValidTimePeriodTimestamp(v interface{}, k string) (ws []string, errors []error) {
	_, err := time.Parse(timePeriodLayout, v.(string))

	if err != nil {
		errors = append(errors, fmt.Errorf("%q cannot be parsed as %q: %w", k, timePeriodLayout, err))
	}

	return
}
