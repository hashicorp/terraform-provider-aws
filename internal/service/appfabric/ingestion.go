// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appfabric"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Ingestion")
// @Tags(identifierAttribute="arn")
func newResourceIngestion(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceIngestion{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameIngestion = "Ingestion"
)

type resourceIngestion struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceIngestion) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_appfabric_ingestion"
}

func (r *resourceIngestion) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"arn": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"app_bundle_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_bundle_arn": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			"ingestion_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"tenant_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceIngestion) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AppFabricClient(ctx)

	var plan resourceIngestionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &appfabric.CreateIngestionInput{
		App:                 aws.String(plan.App.ValueString()),
		AppBundleIdentifier: aws.String(plan.AppBundleIdentifier.ValueString()),
		IngestionType:       awstypes.IngestionType(plan.IngestionType.ValueString()),
		TenantId:            aws.String(plan.TenantId.ValueString()),
		Tags:                getTagsIn(ctx),
	}

	out, err := conn.CreateIngestion(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppFabric, create.ErrActionCreating, ResNameIngestion, plan.App.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Ingestion == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppFabric, create.ErrActionCreating, ResNameIngestion, plan.App.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.App = flex.StringToFramework(ctx, out.Ingestion.App)
	plan.AppBundleArn = flex.StringToFramework(ctx, out.Ingestion.AppBundleArn)
	plan.AppBundleIdentifier = flex.StringToFramework(ctx, out.Ingestion.AppBundleArn)
	plan.ARN = flex.StringToFramework(ctx, out.Ingestion.Arn)
	plan.ID = types.StringValue(createIngestionID(string(*out.Ingestion.AppBundleArn), string(*out.Ingestion.Arn)))
	plan.IngestionType = flex.StringToFramework(ctx, aws.String(string(out.Ingestion.IngestionType)))
	plan.State = flex.StringToFramework(ctx, aws.String(string(out.Ingestion.State)))
	plan.TenantId = flex.StringToFramework(ctx, out.Ingestion.TenantId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceIngestion) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().AppFabricClient(ctx)

	var plan resourceIngestionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &appfabric.CreateIngestionInput{
		App:                 aws.String(plan.App.ValueString()),
		AppBundleIdentifier: aws.String(plan.AppBundleIdentifier.ValueString()),
		IngestionType:       awstypes.IngestionType(plan.IngestionType.ValueString()),
		TenantId:            aws.String(plan.TenantId.ValueString()),
		Tags:                getTagsIn(ctx),
	}

	out, err := conn.CreateIngestion(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppFabric, create.ErrActionCreating, ResNameIngestion, plan.App.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Ingestion == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppFabric, create.ErrActionCreating, ResNameIngestion, plan.App.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.App = flex.StringToFramework(ctx, out.Ingestion.App)
	plan.AppBundleArn = flex.StringToFramework(ctx, out.Ingestion.AppBundleArn)
	plan.AppBundleIdentifier = flex.StringToFramework(ctx, out.Ingestion.AppBundleArn)
	plan.ARN = flex.StringToFramework(ctx, out.Ingestion.Arn)
	plan.ID = types.StringValue(createIngestionID(string(*out.Ingestion.AppBundleArn), string(*out.Ingestion.Arn)))
	plan.IngestionType = flex.StringToFramework(ctx, aws.String(string(out.Ingestion.IngestionType)))
	plan.State = flex.StringToFramework(ctx, aws.String(string(out.Ingestion.State)))
	plan.TenantId = flex.StringToFramework(ctx, out.Ingestion.TenantId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceIngestion) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	conn := r.Meta().AppFabricClient(ctx)

	var state resourceIngestionData

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findIngestionByID(ctx, conn, state.AppBundleIdentifier.ValueString(), state.ARN.ValueString())

	if tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.AppFabric, create.ErrActionReading, ResNameIngestion, state.App.ValueString())
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.Append(create.DiagErrorFramework(names.AppFabric, create.ErrActionReading, ResNameIngestion, state.App.ValueString(), err))
		return
	}

	state.App = flex.StringToFramework(ctx, out.App)
	state.AppBundleArn = flex.StringToFramework(ctx, out.AppBundleArn)
	state.AppBundleIdentifier = flex.StringToFramework(ctx, out.AppBundleArn)
	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.ID = types.StringValue(createIngestionID(string(*out.AppBundleArn), string(*out.Arn)))
	state.IngestionType = flex.StringToFramework(ctx, aws.String(string(out.IngestionType)))
	state.State = flex.StringToFramework(ctx, aws.String(string(out.State)))
	state.TenantId = flex.StringToFramework(ctx, out.TenantId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceIngestion) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	conn := r.Meta().AppFabricClient(ctx)

	var state resourceIngestionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &appfabric.DeleteIngestionInput{
		AppBundleIdentifier: aws.String(state.AppBundleIdentifier.ValueString()),
		IngestionIdentifier: aws.String(state.ARN.ValueString()),
	}

	_, err := conn.DeleteIngestion(ctx, in)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppFabric, create.ErrActionDeleting, ResNameIngestion, state.ARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceIngestion) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: AppBundleIdentifier, IngestionARN. Got: %q", req.ID),
		)
		return
	}

	appBundleId := idParts[0]
	arn := idParts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_bundle_identifier"), aws.String(appBundleId))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("arn"), aws.String(arn))...)
}

func (r *resourceIngestion) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func createIngestionID(appBundleARN string, ingestionARN string) string {
	return strings.Join([]string{appBundleARN, ingestionARN}, ",")
}

func findIngestionByID(ctx context.Context, conn *appfabric.Client, appBundleIdentifier string, arn string) (*awstypes.Ingestion, error) {
	in := &appfabric.GetIngestionInput{
		AppBundleIdentifier: aws.String(appBundleIdentifier),
		IngestionIdentifier: aws.String(arn),
	}
	out, err := conn.GetIngestion(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}
		return nil, err
	}
	if out == nil || out.Ingestion == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}
	return out.Ingestion, nil
}

type resourceIngestionData struct {
	App                 types.String `tfsdk:"app"`
	ARN                 types.String `tfsdk:"arn"`
	AppBundleIdentifier types.String `tfsdk:"app_bundle_identifier"`
	AppBundleArn        types.String `tfsdk:"app_bundle_arn"`
	ID                  types.String `tfsdk:"id"`
	IngestionType       types.String `tfsdk:"ingestion_type"`
	State               types.String `tfsdk:"state"`
	Tags                types.Map    `tfsdk:"tags"`
	TagsAll             types.Map    `tfsdk:"tags_all"`
	TenantId            types.String `tfsdk:"tenant_id"`
}
