package stringplanmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type defaultValue struct {
	val string
}

// DefaultValue return a string plan modifier that sets the specified value if the planned value is Null.
func DefaultValue(s string) planmodifier.String {
	return defaultValue{
		val: s,
	}
}

func (m defaultValue) Description(context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %s", m.val)
}

func (m defaultValue) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m defaultValue) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the attribute configuration is not null, we are done here
	if !req.ConfigValue.IsNull() {
		return
	}

	// If the attribute plan is "known" and "not null", then a previous plan modifier in the sequence
	// has already been applied, and we don't want to interfere.
	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}

	resp.PlanValue = types.StringValue(m.val)
}
