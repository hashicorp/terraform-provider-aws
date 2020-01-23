package validation

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
)

// StringIsNotEmpty is a ValidateFunc that ensures a string is not empty or consisting entirely of whitespace characters
func StringIsNotEmpty(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if v == "" {
		return nil, []error{fmt.Errorf("expected %q to not be an empty string", k)}
	}

	return nil, nil
}

// StringIsNotEmpty is a ValidateFunc that ensures a string is not empty or consisting entirely of whitespace characters
func StringIsNotWhiteSpace(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if strings.TrimSpace(v) == "" {
		return nil, []error{fmt.Errorf("expected %q to not be an empty string or whitespace", k)}
	}

	return nil, nil
}

// StringIsEmpty is a ValidateFunc that ensures a string has no characters
func StringIsEmpty(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if v != "" {
		return nil, []error{fmt.Errorf("expected %q to be an empty string: got %v", k, v)}
	}

	return nil, nil
}

// StringIsEmpty is a ValidateFunc that ensures a string is composed of entirely whitespace
func StringIsWhiteSpace(i interface{}, k string) ([]string, []error) {
	v, ok := i.(string)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be string", k)}
	}

	if strings.TrimSpace(v) != "" {
		return nil, []error{fmt.Errorf("expected %q to be an empty string or whitespace: got %v", k, v)}
	}

	return nil, nil
}

// StringLenBetween returns a SchemaValidateFunc which tests if the provided value
// is of type string and has length between min and max (inclusive)
func StringLenBetween(min, max int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}
		if len(v) < min || len(v) > max {
			es = append(es, fmt.Errorf("expected length of %s to be in the range (%d - %d), got %s", k, min, max, v))
		}
		return
	}
}

// StringMatch returns a SchemaValidateFunc which tests if the provided value
// matches a given regexp. Optionally an error message can be provided to
// return something friendlier than "must match some globby regexp".
func StringMatch(r *regexp.Regexp, message string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) ([]string, []error) {
		v, ok := i.(string)
		if !ok {
			return nil, []error{fmt.Errorf("expected type of %s to be string", k)}
		}

		if ok := r.MatchString(v); !ok {
			if message != "" {
				return nil, []error{fmt.Errorf("invalid value for %s (%s)", k, message)}

			}
			return nil, []error{fmt.Errorf("expected value of %s to match regular expression %q", k, r)}
		}
		return nil, nil
	}
}

// StringDoesNotMatch returns a SchemaValidateFunc which tests if the provided value
// does not match a given regexp. Optionally an error message can be provided to
// return something friendlier than "must not match some globby regexp".
func StringDoesNotMatch(r *regexp.Regexp, message string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) ([]string, []error) {
		v, ok := i.(string)
		if !ok {
			return nil, []error{fmt.Errorf("expected type of %s to be string", k)}
		}

		if ok := r.MatchString(v); ok {
			if message != "" {
				return nil, []error{fmt.Errorf("invalid value for %s (%s)", k, message)}

			}
			return nil, []error{fmt.Errorf("expected value of %s to not match regular expression %q", k, r)}
		}
		return nil, nil
	}
}

// StringInSlice returns a SchemaValidateFunc which tests if the provided value
// is of type string and matches the value of an element in the valid slice
// will test with in lower case if ignoreCase is true
func StringInSlice(valid []string, ignoreCase bool) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		for _, str := range valid {
			if v == str || (ignoreCase && strings.ToLower(v) == strings.ToLower(str)) {
				return
			}
		}

		es = append(es, fmt.Errorf("expected %s to be one of %v, got %s", k, valid, v))
		return
	}
}

// StringDoesNotContainAny returns a SchemaValidateFunc which validates that the
// provided value does not contain any of the specified Unicode code points in chars.
func StringDoesNotContainAny(chars string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if strings.ContainsAny(v, chars) {
			es = append(es, fmt.Errorf("expected value of %s to not contain any of %q", k, chars))
			return
		}

		return
	}
}

// StringIsBase64 is a ValidateFunc that ensures a string can be parsed as Base64
func StringIsBase64(i interface{}, k string) (warnings []string, errors []error) {
	// Empty string is not allowed
	if warnings, errors = StringIsNotEmpty(i, k); len(errors) > 0 {
		return
	}

	// NoEmptyStrings checks it is a string
	v, _ := i.(string)
	if _, err := base64.StdEncoding.DecodeString(v); err != nil {
		errors = append(errors, fmt.Errorf("expected %q to be a base64 string, got %v", k, v))
	}

	return
}

// ValidateListUniqueStrings is a ValidateFunc that ensures a list has no
// duplicate items in it. It's useful for when a list is needed over a set
// because order matters, yet the items still need to be unique.
func ValidateListUniqueStrings(v interface{}, k string) (ws []string, errors []error) {
	for n1, v1 := range v.([]interface{}) {
		for n2, v2 := range v.([]interface{}) {
			if v1.(string) == v2.(string) && n1 != n2 {
				errors = append(errors, fmt.Errorf("%q: duplicate entry - %s", k, v1.(string)))
			}
		}
	}
	return
}

// ValidateJsonString is a SchemaValidateFunc which tests to make sure the
// supplied string is valid JSON.
func ValidateJsonString(v interface{}, k string) (ws []string, errors []error) {
	if _, err := structure.NormalizeJsonString(v); err != nil {
		errors = append(errors, fmt.Errorf("%q contains an invalid JSON: %s", k, err))
	}
	return
}

// ValidateRegexp returns a SchemaValidateFunc which tests to make sure the
// supplied string is a valid regular expression.
func ValidateRegexp(v interface{}, k string) (ws []string, errors []error) {
	if _, err := regexp.Compile(v.(string)); err != nil {
		errors = append(errors, fmt.Errorf("%q: %s", k, err))
	}
	return
}
