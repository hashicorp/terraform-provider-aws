package flex

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Terraform Plugin Framework variants of standard flatteners and expanders.

func ExpandFrameworkStringSet(ctx context.Context, set types.Set) []*string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}

	var vs []*string

	if set.ElementsAs(ctx, &vs, false).HasError() {
		return nil
	}

	return vs
}

func ExpandFrameworkStringValueSet(ctx context.Context, set types.Set) []string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}

	var vs []string

	if set.ElementsAs(ctx, &vs, false).HasError() {
		return nil
	}

	return vs
}

func FlattenFrameworkStringList(_ context.Context, vs []*string) types.List {
	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.String{Value: aws.ToString(v)}
	}

	return types.List{ElemType: types.StringType, Elems: elems}
}

func FlattenFrameworkStringValueList(_ context.Context, vs []string) types.List {
	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.String{Value: v}
	}

	return types.List{ElemType: types.StringType, Elems: elems}
}

func FlattenFrameworkStringValueSet(_ context.Context, vs []string) types.Set {
	elems := make([]attr.Value, len(vs))

	for i, v := range vs {
		elems[i] = types.String{Value: v}
	}

	return types.Set{ElemType: types.StringType, Elems: elems}
}

func FlattenFrameworkStringValueMap(_ context.Context, m map[string]string) types.Map {
	elems := make(map[string]attr.Value, len(m))

	for k, v := range m {
		elems[k] = types.String{Value: v}
	}

	return types.Map{ElemType: types.StringType, Elems: elems}
}
