package sdktypes

import (
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	awsdiag "github.com/hashicorp/terraform-provider-aws/internal/diag"
)

const (
	TypeDuration = schema.TypeString
)

type Duration string

func (d Duration) IsNull() bool {
	return d == ""
}

func (d Duration) Value() (time.Duration, bool, error) {
	if d.IsNull() {
		return 0, true, nil
	}

	value, err := time.ParseDuration(string(d))
	if err != nil {
		return 0, false, err
	}
	return value, false, nil
}

func ValidateDuration(i any, path cty.Path) diag.Diagnostics {
	v, ok := i.(string)
	if !ok {
		return diag.Diagnostics{awsdiag.NewIncorrectValueTypeAttributeError(path, "string")}
	}

	duration, err := time.ParseDuration(v)
	if err != nil {
		return diag.Diagnostics{awsdiag.NewInvalidValueAttributeErrorf(path, "Cannot be parsed as duration: %s", err)}
	}

	if duration < 0 {
		return diag.Diagnostics{awsdiag.NewInvalidValueAttributeError(path, "Must be greater than zero")}
	}

	return nil
}
