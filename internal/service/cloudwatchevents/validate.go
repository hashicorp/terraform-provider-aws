package cloudwatchevents

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func mapKeysDoNotMatch(r *regexp.Regexp, message string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		m, ok := i.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be map", k))
			return warnings, errors
		}

		for key := range m {
			if ok := r.MatchString(key); ok {
				errors = append(errors, fmt.Errorf("%s: %s: %s", k, message, key))
			}
		}

		return warnings, errors
	}
}

func mapMaxItems(max int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		m, ok := i.(map[string]interface{})
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be map", k))
			return warnings, errors
		}

		if len(m) > max {
			errors = append(errors, fmt.Errorf("expected number of items in %s to be less than or equal to %d, got %d", k, max, len(m)))
		}

		return warnings, errors
	}
}

var validArchiveName = validation.All(
	validation.StringLenBetween(1, 48),
	validation.StringMatch(regexp.MustCompile(`^[\.\-_A-Za-z0-9]+`), ""),
)

var validBusNameOrARN = validation.Any(
	verify.ValidARN,
	validation.All(
		validation.StringLenBetween(1, 256),
		validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9._\-/]+$`), ""),
	),
)

var validCloudWatchEventTargetId = validation.All(
	validation.StringLenBetween(1, 64),
	validation.StringMatch(regexp.MustCompile(`^[\.\-_A-Za-z0-9]+$`), "must contain only alphanumeric characters, underscores, periods and hyphens"),
)

var validCloudWatchEventRuleName = validation.All(
	validation.StringLenBetween(1, 64),
	validation.StringMatch(regexp.MustCompile(`^[\.\-_A-Za-z0-9]+$`), "must contain only alphanumeric characters, underscores, periods and hyphens"),
)

var validCustomEventBusEventSourceName = validation.All(
	validation.StringLenBetween(1, 256),
	validation.StringMatch(regexp.MustCompile(`^aws\.partner(/[\.\-_A-Za-z0-9]+){2,}$`), ""),
)

var validCustomEventBusName = validation.All(
	validation.StringLenBetween(1, 256),
	validation.StringMatch(regexp.MustCompile(`^[/\.\-_A-Za-z0-9]+$`), ""),
	validation.StringDoesNotMatch(regexp.MustCompile(`^default$`), "cannot be 'default'"),
)
