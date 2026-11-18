// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdamicrovms

import (
	"context"
	"fmt"

	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdamicrovms/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// Models shared by the image resource, the image version data source, and the
// MicroVM resource.

type hooksModel struct {
	MicroVMHooks      fwtypes.ListNestedObjectValueOf[microVMHooksModel]      `tfsdk:"microvm_hooks"`
	MicroVMImageHooks fwtypes.ListNestedObjectValueOf[microVMImageHooksModel] `tfsdk:"microvm_image_hooks"`
	Port              types.Int32                                             `tfsdk:"port"`
}

type microVMHooksModel struct {
	Resume                    fwtypes.StringEnum[awstypes.HookState] `tfsdk:"resume"`
	ResumeTimeoutInSeconds    types.Int32                            `tfsdk:"resume_timeout_in_seconds"`
	Run                       fwtypes.StringEnum[awstypes.HookState] `tfsdk:"run"`
	RunTimeoutInSeconds       types.Int32                            `tfsdk:"run_timeout_in_seconds"`
	Suspend                   fwtypes.StringEnum[awstypes.HookState] `tfsdk:"suspend"`
	SuspendTimeoutInSeconds   types.Int32                            `tfsdk:"suspend_timeout_in_seconds"`
	Terminate                 fwtypes.StringEnum[awstypes.HookState] `tfsdk:"terminate"`
	TerminateTimeoutInSeconds types.Int32                            `tfsdk:"terminate_timeout_in_seconds"`
}

type microVMImageHooksModel struct {
	Ready                    fwtypes.StringEnum[awstypes.HookState] `tfsdk:"ready"`
	ReadyTimeoutInSeconds    types.Int32                            `tfsdk:"ready_timeout_in_seconds"`
	Validate                 fwtypes.StringEnum[awstypes.HookState] `tfsdk:"validate"`
	ValidateTimeoutInSeconds types.Int32                            `tfsdk:"validate_timeout_in_seconds"`
}

type loggingModel struct {
	CloudWatch fwtypes.ListNestedObjectValueOf[cloudWatchLoggingModel] `tfsdk:"cloudwatch"`
	Disabled   fwtypes.ListNestedObjectValueOf[loggingDisabledModel]   `tfsdk:"disabled"`
}

type cloudWatchLoggingModel struct {
	LogGroup  types.String `tfsdk:"log_group"`
	LogStream types.String `tfsdk:"log_stream"`
}

type loggingDisabledModel struct{}

type resourcesModel struct {
	MinimumMemoryInMiB types.Int32 `tfsdk:"minimum_memory_in_mib"`
}

var (
	_ fwflex.Expander  = loggingModel{}
	_ fwflex.Flattener = (*loggingModel)(nil)
)

func (m loggingModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.CloudWatch.IsNull():
		model, d := m.CloudWatch.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.LoggingMemberCloudWatch
		diags.Append(fwflex.Expand(ctx, model, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags

	case !m.Disabled.IsNull():
		r := awstypes.LoggingMemberDisabled{
			Value: awstypes.LoggingDisabled{},
		}
		return &r, diags
	}

	return nil, diags
}

func (m *loggingModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.LoggingMemberCloudWatch:
		var data cloudWatchLoggingModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.CloudWatch = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
		m.Disabled = fwtypes.NewListNestedObjectValueOfNull[loggingDisabledModel](ctx)

	case awstypes.LoggingMemberDisabled:
		m.CloudWatch = fwtypes.NewListNestedObjectValueOfNull[cloudWatchLoggingModel](ctx)
		m.Disabled = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &loggingDisabledModel{})

	case awstypes.UnknownUnionMember:
		fwflex.HandleFlattenUnknownUnionMember(ctx, t.Tag, &diags)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("loggingModel.Flatten: %T", v),
		)
	}

	return diags
}
