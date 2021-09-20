package wafregional

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func validMetricName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[0-9A-Za-z]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"Only alphanumeric characters allowed in %q: %q",
			k, value))
	}
	return
}

func validPredicatesType() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		waf.PredicateTypeByteMatch,
		waf.PredicateTypeGeoMatch,
		waf.PredicateTypeIpmatch,
		waf.PredicateTypeRegexMatch,
		waf.PredicateTypeSizeConstraint,
		waf.PredicateTypeSqlInjectionMatch,
		waf.PredicateTypeXssMatch,
	}, false)
}

func sliceContainsMap(l []interface{}, m map[string]interface{}) (int, bool) {
	for i, t := range l {
		if reflect.DeepEqual(m, t.(map[string]interface{})) {
			return i, true
		}
	}

	return -1, false
}
