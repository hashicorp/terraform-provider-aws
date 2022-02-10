package iam

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var validRolePolicyName = validIamResourceName(rolePolicyNameMaxLen)

func validIamResourceName(max int) schema.SchemaValidateFunc {
	return validation.All(
		validation.StringLenBetween(1, max),
		validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]*$`), "must match [\\w+=,.@-]"),
	)
}

var validAccountAlias = validation.All(
	validation.StringLenBetween(3, 63),
	validation.StringMatch(regexp.MustCompile(`^[a-z0-9][a-z0-9-]+$`), "must start with an alphanumeric character and only contain lowercase alphanumeric characters and hyphens"),
	func(v interface{}, k string) (ws []string, es []error) {
		val := v.(string)
		if strings.Contains(val, "--") {
			es = append(es, fmt.Errorf("%q must not contain consecutive hyphens", k))
		}
		if strings.HasSuffix(val, "-") {
			es = append(es, fmt.Errorf("%q must not end in a hyphen", k))
		}
		return
	},
)

func validOpenIDURL(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	u, err := url.Parse(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q has to be a valid URL", k))
		return
	}
	if u.Scheme != "https" {
		errors = append(errors, fmt.Errorf("%q has to use HTTPS scheme (i.e. begin with https://)", k))
	}
	if len(u.Query()) > 0 {
		errors = append(errors, fmt.Errorf("%q cannot contain query parameters per the OIDC standard", k))
	}
	return
}
