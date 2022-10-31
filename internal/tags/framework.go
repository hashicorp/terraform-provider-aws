package tags

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Terraform Plugin Framework variants of tags schemas.

func TagsAttributeComputed() tfsdk.Attribute {
	return tfsdk.Attribute{
		Type:     types.MapType{ElemType: types.StringType},
		Optional: true,
		Computed: true,
	}
}
