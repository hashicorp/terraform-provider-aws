package stringplanmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type stringDefaultValue struct {
	DefaultValue types.String
}

// StringDefaultValue return a string plan modifier that sets the specified value if the planned value is Null.
func StringDefaultValue(s types.String) planmodifier.String {
	return stringDefaultValue{
		DefaultValue: s,
	}
}

func (m stringDefaultValue) Description(context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %s", m.DefaultValue)
}

func (m stringDefaultValue) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m stringDefaultValue) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the attribute configuration is not null, we are done here
	if !req.ConfigValue.IsNull() {
		return
	}

	// If the attribute plan is "known" and "not null", then a previous plan modifier in the sequence
	// has already been applied, and we don't want to interfere.
	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}

	resp.PlanValue = m.DefaultValue
}
