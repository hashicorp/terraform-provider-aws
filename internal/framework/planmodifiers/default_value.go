package planmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type defaultValue struct {
	value attr.Value
}

// DefaultValue return an AttributePlanModifier that sets the specified value if the planned value is Null.
func DefaultValue(value attr.Value) tfsdk.AttributePlanModifier {
	return defaultValue{
		value: value,
	}
}

func DefaultStringValue(value string) tfsdk.AttributePlanModifier {
	return DefaultValue(types.StringValue(value))
}

func (m defaultValue) Description(context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %s", m.value)
}

func (m defaultValue) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m defaultValue) Modify(ctx context.Context, request tfsdk.ModifyAttributePlanRequest, response *tfsdk.ModifyAttributePlanResponse) {
	if v, err := request.AttributePlan.ToTerraformValue(ctx); err != nil {
		response.Diagnostics.AddAttributeError(request.AttributePath, "getting attribute value", err.Error())

		return
	} else if v.IsNull() {
		response.AttributePlan = m.value
	}
}
