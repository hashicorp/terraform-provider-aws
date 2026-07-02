// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ssoadmin_region", name="Region")
// @IdentityAttribute("instance_arn")
// @IdentityAttribute("region_name")
// @ImportIDHandler("regionImportID")
// @Testing(preCheckWithRegion="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckSSOAdminInstancesWithRegion")
// @Testing(serialize=true)
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(identityRegionOverrideTest=false)
// @Testing(importStateIdAttributes="instance_arn;region_name", importStateIdAttributesSep="flex.ResourceIdSeparator")
func newRegionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &regionResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameRegion = "Region"

	regionIDPartCount = 2
)

type regionResource struct {
	framework.ResourceWithModel[regionResourceModel]
	framework.WithNoUpdate
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *regionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"instance_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RegionStatus](),
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *regionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan regionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSOAdminClient(ctx)

	var input ssoadmin.AddRegionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.AddRegion(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, aws.ToString(input.RegionName))
		return
	}

	output, err := waitRegionActive(ctx, conn, aws.ToString(input.InstanceArn), aws.ToString(input.RegionName), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, aws.ToString(input.RegionName))
		return
	}

	plan.Status = fwtypes.StringEnumValue(output.Status)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *regionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state regionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSOAdminClient(ctx)

	var input ssoadmin.DescribeRegionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, state, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := findRegionByTwoPartKey(ctx, conn, aws.ToString(input.InstanceArn), aws.ToString(input.RegionName))
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, aws.ToString(input.RegionName))
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, output, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *regionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state regionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSOAdminClient(ctx)

	var input ssoadmin.RemoveRegionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, state, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.RemoveRegion(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, aws.ToString(input.RegionName))
		return
	}

	if err := waitRegionDeleted(ctx, conn, aws.ToString(input.InstanceArn), aws.ToString(input.RegionName), r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, aws.ToString(input.RegionName))
		return
	}
}

func findRegionByTwoPartKey(ctx context.Context, conn *ssoadmin.Client, instanceARN, regionName string) (*ssoadmin.DescribeRegionOutput, error) {
	input := &ssoadmin.DescribeRegionInput{
		InstanceArn: aws.String(instanceARN),
		RegionName:  aws.String(regionName),
	}

	output, err := conn.DescribeRegion(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}

func statusRegion(conn *ssoadmin.Client, instanceARN, regionName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findRegionByTwoPartKey(ctx, conn, instanceARN, regionName)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return output, string(output.Status), nil
	}
}

func waitRegionActive(ctx context.Context, conn *ssoadmin.Client, instanceARN, regionName string, timeout time.Duration) (*ssoadmin.DescribeRegionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RegionStatusAdding),
		Target:  enum.Slice(awstypes.RegionStatusActive),
		Refresh: statusRegion(conn, instanceARN, regionName),
		Timeout: timeout,
		Delay:   15 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ssoadmin.DescribeRegionOutput); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitRegionDeleted(ctx context.Context, conn *ssoadmin.Client, instanceARN, regionName string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RegionStatusRemoving),
		Target:  []string{},
		Refresh: statusRegion(conn, instanceARN, regionName),
		Timeout: timeout,
		Delay:   15 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return smarterr.NewError(err)
}

var _ inttypes.ImportIDParser = regionImportID{}

type regionImportID struct{}

func (regionImportID) Parse(id string) (string, map[string]any, error) {
	parts, err := intflex.ExpandResourceId(id, regionIDPartCount, false)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"instance_arn": parts[0],
		"region_name":  parts[1],
	}

	return id, result, nil
}

type regionResourceModel struct {
	framework.WithRegionModel
	InstanceARN fwtypes.ARN                               `tfsdk:"instance_arn"`
	RegionName  types.String                              `tfsdk:"region_name"`
	Status      fwtypes.StringEnum[awstypes.RegionStatus] `tfsdk:"status"`
	Timeouts    timeouts.Value                            `tfsdk:"timeouts"`
}
