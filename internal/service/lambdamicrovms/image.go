// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdamicrovms

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdamicrovms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdamicrovms/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_lambdamicrovms_image", name="Image")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/lambdamicrovms;lambdamicrovms.GetMicrovmImageOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
func newImageResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &imageResource{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameImage = "Image"
)

type imageResource struct {
	framework.ResourceWithModel[imageResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *imageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"additional_os_capabilities": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringEnumType[awstypes.Capability](),
				Optional:    true,
				ElementType: types.StringType,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"base_image_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"base_image_version": schema.StringAttribute{
				Optional: true,
			},
			"build_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"egress_network_connectors": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Optional:    true,
				ElementType: types.StringType,
			},
			"environment_variables": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				Optional:    true,
				ElementType: types.StringType,
			},
			"latest_active_image_version": schema.StringAttribute{
				Computed: true,
			},
			"latest_failed_image_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrState: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.MicrovmImageState](),
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *imageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	conn := r.Meta().LambdaMicrovmsClient(ctx)

	var plan imageResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input lambdamicrovms.CreateMicrovmImageInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	input.ClientToken = aws.String(sdkid.UniqueId())
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateMicrovmImage(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)

	outWait, err := waitImageCreated(ctx, conn, plan.ImageArn.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, outWait, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *imageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LambdaMicrovmsClient(ctx)

	var state imageResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findImageByARN(ctx, conn, state.ImageArn.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ImageArn.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *imageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LambdaMicrovmsClient(ctx)

	var plan, state imageResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input lambdamicrovms.UpdateMicrovmImageInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		input.ImageIdentifier = plan.ImageArn.ValueStringPointer()
		input.ClientToken = aws.String(sdkid.UniqueId())
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateMicrovmImage(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ImageArn.String())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.ImageArn.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
		if resp.Diagnostics.HasError() {
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		outWait, err := waitImageUpdated(ctx, conn, plan.ImageArn.ValueString(), updateTimeout)

		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ImageArn.String())
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, outWait, &plan))
		if resp.Diagnostics.HasError() {
			return
		}

	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *imageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	conn := r.Meta().LambdaMicrovmsClient(ctx)

	var state imageResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := lambdamicrovms.DeleteMicrovmImageInput{
		ImageIdentifier: aws.String(state.ImageArn.ValueString()),
	}

	_, err := conn.DeleteMicrovmImage(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ImageArn.ValueString())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitImageDeleted(ctx, conn, state.ImageArn.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ImageArn.ValueString())
		return
	}
}

func waitImageCreated(ctx context.Context, conn *lambdamicrovms.Client, id string, timeout time.Duration) (*lambdamicrovms.GetMicrovmImageOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.MicrovmImageStateCreating),
		Target:                    enum.Slice(awstypes.MicrovmImageStateCreated),
		Refresh:                   statusImage(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdamicrovms.GetMicrovmImageOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitImageUpdated(ctx context.Context, conn *lambdamicrovms.Client, id string, timeout time.Duration) (*lambdamicrovms.GetMicrovmImageOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.MicrovmImageStateUpdating),
		Target:                    enum.Slice(awstypes.MicrovmImageStateUpdated),
		Refresh:                   statusImage(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdamicrovms.GetMicrovmImageOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitImageDeleted(ctx context.Context, conn *lambdamicrovms.Client, id string, timeout time.Duration) (*lambdamicrovms.GetMicrovmImageOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.MicrovmImageStateDeleting),
		Target:  []string{},
		Refresh: statusImage(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*lambdamicrovms.GetMicrovmImageOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusImage(conn *lambdamicrovms.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findImageByARN(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.State), nil
	}
}

func findImageByARN(ctx context.Context, conn *lambdamicrovms.Client, arn string) (*lambdamicrovms.GetMicrovmImageOutput, error) {
	input := lambdamicrovms.GetMicrovmImageInput{
		ImageIdentifier: aws.String(arn),
	}

	out, err := conn.GetMicrovmImage(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type imageResourceModel struct {
	framework.WithRegionModel
	AdditionalOsCapabilities fwtypes.ListOfStringEnum[awstypes.Capability]  `tfsdk:"additional_os_capabilities"`
	BaseImageArn             fwtypes.ARN                                    `tfsdk:"base_image_arn"`
	BaseImageVersion         types.String                                   `tfsdk:"base_image_version"`
	BuildRoleArn             fwtypes.ARN                                    `tfsdk:"build_role_arn"`
	Description              types.String                                   `tfsdk:"description"`
	EgressNetworkConnectors  fwtypes.ListOfString                           `tfsdk:"egress_network_connectors"`
	EnvironmentVariables     fwtypes.MapOfString                            `tfsdk:"environment_variables"`
	ImageArn                 types.String                                   `tfsdk:"arn"`
	LatestActiveImageVersion types.String                                   `tfsdk:"latest_active_image_version"`
	LatestFailedImageVersion types.String                                   `tfsdk:"latest_failed_image_version"`
	Name                     types.String                                   `tfsdk:"name"`
	State                    fwtypes.StringEnum[awstypes.MicrovmImageState] `tfsdk:"state"`
	Timeouts                 timeouts.Value                                 `tfsdk:"timeouts"`
}

func sweepImages(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := lambdamicrovms.ListMicrovmImagesInput{}
	conn := client.LambdaMicrovmsClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := lambdamicrovms.NewListMicrovmImagesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newImageResource, client,
				sweepfw.NewAttribute(names.AttrARN, aws.ToString(v.ImageArn))),
			)
		}
	}

	return sweepResources, nil
}
