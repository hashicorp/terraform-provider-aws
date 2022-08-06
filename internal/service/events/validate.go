package events

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func validateRuleName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 64 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 64 characters: %q", k, value))
	}

	// http://docs.aws.amazon.com/eventbridge/latest/APIReference/API_PutRule.html
	pattern := `^[\.\-_A-Za-z0-9]+$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q doesn't comply with restrictions (%q): %q",
			k, pattern, value))
	}

	return
}

func validateTargetID(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 64 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 64 characters: %q", k, value))
	}

	// http://docs.aws.amazon.com/eventbridge/latest/APIReference/API_Target.html
	pattern := `^[\.\-_A-Za-z0-9]+$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q doesn't comply with restrictions (%q): %q",
			k, pattern, value))
	}

	return
}

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

var validSourceName = validation.All(
	validation.StringLenBetween(1, 256),
	validation.StringMatch(regexp.MustCompile(`^aws\.partner(/[\.\-_A-Za-z0-9]+){2,}$`), ""),
)

var validCustomEventBusName = validation.All(
	validation.StringLenBetween(1, 256),
	validation.StringMatch(regexp.MustCompile(`^[/\.\-_A-Za-z0-9]+$`), ""),
	validation.StringDoesNotMatch(regexp.MustCompile(`^default$`), "cannot be 'default'"),
)
