// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package xray

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/xray"
	awstypes "github.com/aws/aws-sdk-go-v2/service/xray/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_xray_trace_segment_destination", name="Trace Segment Destination")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/xray;xray.GetTraceSegmentDestinationOutput")
// @Testing(generator=false)
// @Testing(checkDestroyNoop=true)
// @Testing(serialize=true)
func newTraceSegmentDestinationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &traceSegmentDestinationResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)

	return r, nil
}

type traceSegmentDestinationResource struct {
	framework.ResourceWithModel[traceSegmentDestinationResourceModel]
	framework.WithTimeouts
	framework.WithNoOpDelete
	framework.WithImportByIdentity
}

func (r *traceSegmentDestinationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDestination: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.TraceSegmentDestination](),
				Required:   true,
			},
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrRegion)),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *traceSegmentDestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan traceSegmentDestinationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().XRayClient(ctx)

	input := xray.UpdateTraceSegmentDestinationInput{
		Destination: plan.Destination.ValueEnum(),
	}
	_, err := conn.UpdateTraceSegmentDestination(ctx, &input)
	switch {
	case errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "The destination is already set to "):
		// e.g. "InvalidRequestException: The destination is already set to XRay".
	case err != nil:
		resp.Diagnostics.AddError("creating XRay Trace Segment Destination", err.Error())
		return
	}

	plan.ID = fwflex.StringValueToFramework(ctx, r.Meta().Region(ctx))

	if _, err := waitTraceSegmentDestinationActive(ctx, conn, r.CreateTimeout(ctx, plan.Timeouts)); err != nil {
		resp.Diagnostics.AddError("waiting for XRay Trace Segment Destination create", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *traceSegmentDestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state traceSegmentDestinationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().XRayClient(ctx)

	output, err := findTraceSegmentDestination(ctx, conn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("reading XRay Trace Segment Destination", err.Error())
		return
	}

	state.Destination = fwtypes.StringEnumValue(output.Destination)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *traceSegmentDestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan traceSegmentDestinationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().XRayClient(ctx)

	input := xray.UpdateTraceSegmentDestinationInput{
		Destination: plan.Destination.ValueEnum(),
	}
	_, err := conn.UpdateTraceSegmentDestination(ctx, &input)
	switch {
	case errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "The destination is already set to "):
		// e.g. "InvalidRequestException: The destination is already set to XRay".
	case err != nil:
		resp.Diagnostics.AddError("updating XRay Trace Segment Destination", err.Error())
		return
	}

	if _, err := waitTraceSegmentDestinationActive(ctx, conn, r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
		resp.Diagnostics.AddError("waiting for XRay Trace Segment Destination update", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func findTraceSegmentDestination(ctx context.Context, conn *xray.Client) (*xray.GetTraceSegmentDestinationOutput, error) {
	var input xray.GetTraceSegmentDestinationInput
	output, err := conn.GetTraceSegmentDestination(ctx, &input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusTraceSegmentDestination(conn *xray.Client) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTraceSegmentDestination(ctx, conn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitTraceSegmentDestinationActive(ctx context.Context, conn *xray.Client, timeout time.Duration) (*xray.GetTraceSegmentDestinationOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TraceSegmentDestinationStatusPending),
		Target:  enum.Slice(awstypes.TraceSegmentDestinationStatusActive),
		Refresh: statusTraceSegmentDestination(conn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*xray.GetTraceSegmentDestinationOutput); ok {
		return out, err
	}

	return nil, err
}

type traceSegmentDestinationResourceModel struct {
	framework.WithRegionModel
	Destination fwtypes.StringEnum[awstypes.TraceSegmentDestination] `tfsdk:"destination"`
	ID          types.String                                         `tfsdk:"id"`
	Timeouts    timeouts.Value                                       `tfsdk:"timeouts"`
}
