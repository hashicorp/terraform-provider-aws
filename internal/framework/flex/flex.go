package flex

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// Breaking the cycles without changing all files
var (
	BoolFromFramework        = flex.BoolFromFramework
	ExpandFrameworkStringSet = flex.ExpandFrameworkStringSet
	Int64FromFramework       = flex.Int64FromFramework
	StringFromFramework      = flex.StringFromFramework
)

var (
	BoolToFramework           = flex.BoolToFramework
	FlattenFrameworkStringSet = flex.FlattenFrameworkStringSet
	Int64ToFramework          = flex.Int64ToFramework
	Int64ToFrameworkLegacy    = flex.Int64ToFrameworkLegacy
	StringToFramework         = flex.StringToFramework
	StringToFrameworkLegacy   = flex.StringToFrameworkLegacy
)

func ARNStringFromFramework(_ context.Context, v fwtypes.ARN) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	return aws.String(v.ValueARN().String())
}

func StringToFrameworkARN(ctx context.Context, v *string, diags *diag.Diagnostics) basetypes.StringValuable {
	if v == nil {
		return fwtypes.ARNNull()
	}

	s := StringToFramework(ctx, v)

	x, d := fwtypes.ARNType.ValueFromString(ctx, s)
	diags.Append(d...)

	return x
}
