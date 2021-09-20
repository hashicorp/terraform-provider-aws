package autoscaling

import (
	"fmt"
	"time"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func validScheduleTimestamp(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := time.Parse(awsAutoscalingScheduleTimeLayout, value)
	if err != nil {
		errors = append(errors, fmt.Errorf(
			"%q cannot be parsed as iso8601 Timestamp Format", value))
	}

	return
}
