package boolplanmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type defaultValue struct {
	val bool
}

// DefaultValue return a bool plan modifier that sets the specified value if the planned value is Null.
func DefaultValue(b bool) planmodifier.Bool {
	return defaultValue{
		val: b,
	}
}

func (m defaultValue) Description(context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %t", m.val)
}

func (m defaultValue) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m defaultValue) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if !req.ConfigValue.IsNull() {
		return
	}

	// If the attribute plan is "known" and "not null", then a previous plan modifier in the sequence
	// has already been applied, and we don't want to interfere.
	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}

	resp.PlanValue = types.BoolValue(m.val)
}
