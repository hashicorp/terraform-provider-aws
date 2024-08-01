// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_rds_integration", name="Integration")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newResourceIntegration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceIntegration{}

	r.SetDefaultCreateTimeout(60 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameIntegration = "Integration"

	IntegrationStatusCreating       = "creating"
	IntegrationStatusActive         = "active"
	IntegrationStatusModifying      = "modifying"
	IntegrationStatusFailed         = "failed"
	IntegrationStatusDeleting       = "deleting"
	IntegrationStatusSyncing        = "syncing"
	IntegrationStatusNeedsAttention = "needs_attention"
)

type resourceIntegration struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceIntegration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_rds_integration"
}

func (r *resourceIntegration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"additional_encryption_context": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"id": framework.IDAttribute(),
			"integration_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kms_key_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"target_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceIntegration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RDSClient(ctx)

	var plan resourceIntegrationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &rds.CreateIntegrationInput{
		IntegrationName: aws.String(plan.IntegrationName.ValueString()),
		SourceArn:       aws.String(plan.SourceArn.ValueString()),
		Tags:            getTagsInV2(ctx),
		TargetArn:       aws.String(plan.TargetArn.ValueString()),
	}
	if !plan.KMSKeyId.IsNull() {
		in.KMSKeyId = aws.String(plan.KMSKeyId.ValueString())
	}
	if !plan.AdditionalEncryptionContext.IsNull() {
		in.AdditionalEncryptionContext = flex.ExpandFrameworkStringValueMap(ctx, plan.AdditionalEncryptionContext)
	}

	out, err := conn.CreateIntegration(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionCreating, ResNameIntegration, plan.IntegrationName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.IntegrationName == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionCreating, ResNameIntegration, plan.IntegrationName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.IntegrationArn)
	plan.ID = types.StringValue(aws.ToString(out.IntegrationArn))

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitIntegrationCreated(ctx, conn, plan.ARN.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionWaitingForCreation, ResNameIntegration, plan.IntegrationName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceIntegration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().RDSClient(ctx)

	var state resourceIntegrationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindIntegrationByARN(ctx, conn, state.ARN.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionSetting, ResNameIntegration, state.ARN.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(errors.New("not found")))
		resp.State.RemoveResource(ctx)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.IntegrationArn)

	state.refreshFromOutput(ctx, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceIntegration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceIntegrationData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceIntegration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().RDSClient(ctx)

	var state resourceIntegrationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &rds.DeleteIntegrationInput{
		IntegrationIdentifier: aws.String(state.ARN.ValueString()),
	}

	_, err := conn.DeleteIntegration(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundFault
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionDeleting, ResNameIntegration, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitIntegrationDeleted(ctx, conn, state.ARN.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionWaitingForDeletion, ResNameIntegration, state.ARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceIntegration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("arn"), req, resp)
}

func (r *resourceIntegration) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func waitIntegrationCreated(ctx context.Context, conn *rds.Client, arn string, timeout time.Duration) (*awstypes.Integration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{IntegrationStatusCreating, IntegrationStatusModifying},
		Target:                    []string{IntegrationStatusActive},
		Refresh:                   statusIntegration(ctx, conn, arn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Integration); ok {
		return out, err
	}

	return nil, err
}

func waitIntegrationDeleted(ctx context.Context, conn *rds.Client, arn string, timeout time.Duration) (*awstypes.Integration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{IntegrationStatusDeleting, IntegrationStatusActive},
		Target:  []string{},
		Refresh: statusIntegration(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Integration); ok {
		return out, err
	}

	return nil, err
}

func statusIntegration(ctx context.Context, conn *rds.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindIntegrationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func FindIntegrationByARN(ctx context.Context, conn *rds.Client, arn string) (*awstypes.Integration, error) {
	in := &rds.DescribeIntegrationsInput{
		IntegrationIdentifier: aws.String(arn),
	}

	out, err := conn.DescribeIntegrations(ctx, in)

	if errs.IsA[*awstypes.IntegrationNotFoundFault](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.Integrations) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &out.Integrations[0], nil
}

type resourceIntegrationData struct {
	ARN                         types.String   `tfsdk:"arn"`
	AdditionalEncryptionContext types.Map      `tfsdk:"additional_encryption_context"`
	ID                          types.String   `tfsdk:"id"`
	IntegrationName             types.String   `tfsdk:"integration_name"`
	KMSKeyId                    types.String   `tfsdk:"kms_key_id"`
	SourceArn                   fwtypes.ARN    `tfsdk:"source_arn"`
	Tags                        types.Map      `tfsdk:"tags"`
	TagsAll                     types.Map      `tfsdk:"tags_all"`
	TargetArn                   fwtypes.ARN    `tfsdk:"target_arn"`
	Timeouts                    timeouts.Value `tfsdk:"timeouts"`
}

// refreshFromOutput writes state data from an AWS response object
func (r *resourceIntegrationData) refreshFromOutput(ctx context.Context, out *awstypes.Integration) {
	if out == nil {
		return
	}

	r.ARN = flex.StringToFramework(ctx, out.IntegrationArn)
	r.AdditionalEncryptionContext = flex.FlattenFrameworkStringValueMap(ctx, out.AdditionalEncryptionContext)
	r.ID = types.StringValue(aws.ToString(out.IntegrationArn))
	r.IntegrationName = flex.StringToFramework(ctx, out.IntegrationName)
	r.KMSKeyId = flex.StringToFramework(ctx, out.KMSKeyId)
	r.SourceArn = flex.StringToFrameworkARN(ctx, out.SourceArn)
	r.TargetArn = flex.StringToFrameworkARN(ctx, out.TargetArn)

	setTagsOutV2(ctx, out.Tags)
}
