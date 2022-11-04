package tags

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/fwplanmodifiers"
)

// Terraform Plugin Framework variants of tags schemas.

func TagsAttribute() tfsdk.Attribute {
	return tfsdk.Attribute{
		Type:     types.MapType{ElemType: types.StringType},
		Optional: true,
		PlanModifiers: []tfsdk.AttributePlanModifier{
			fwplanmodifiers.NormalizeEmptyMap(),
		},
	}
}

func TagsAttributeComputed() tfsdk.Attribute {
	return tfsdk.Attribute{
		Type:     types.MapType{ElemType: types.StringType},
		Optional: true,
		Computed: true,
	}
}
