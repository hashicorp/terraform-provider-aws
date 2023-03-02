package int64planmodifier

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type defaultValue struct {
	val int64
}

// DefaultValue return a bool plan modifier that sets the specified value if the planned value is Null.
func DefaultValue(i int64) planmodifier.Int64 {
	return defaultValue{
		val: i,
	}
}

func (m defaultValue) Description(context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %d", m.val)
}

func (m defaultValue) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m defaultValue) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.ConfigValue.IsNull() {
		return
	}

	// If the attribute plan is "known" and "not null", then a previous plan modifier in the sequence
	// has already been applied, and we don't want to interfere.
	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}

	resp.PlanValue = types.Int64Value(m.val)
}
