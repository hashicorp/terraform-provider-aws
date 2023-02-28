package flex

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

func StringToFrameworkARN(ctx context.Context, v *string, diags *diag.Diagnostics) fwtypes.ARN {
	if v == nil {
		return fwtypes.ARNNull()
	}

	a, err := arn.Parse(aws.ToString(v))
	if err != nil {
		diags.AddError(
			"Parsing Error",
			fmt.Sprintf("String %s cannot be parsed as an ARN.", aws.ToString(v)),
		)
	}

	return fwtypes.ARNValue(a)
}
