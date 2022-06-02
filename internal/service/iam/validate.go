package iam

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var validRolePolicyName = validResourceName(rolePolicyNameMaxLen)

func validResourceName(max int) schema.SchemaValidateFunc {
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

var validOpenIDURL = validation.All(
	validation.IsURLWithHTTPS,
	func(v interface{}, k string) (ws []string, es []error) {
		value := v.(string)
		u, err := url.Parse(value)
		if err != nil {
			// validation.IsURLWithHTTPS will already have returned an error for an invalid URL
			return
		}
		if len(u.Query()) > 0 {
			es = append(es, fmt.Errorf("%q cannot contain query parameters per the OIDC standard", k))
		}
		return
	},
)
