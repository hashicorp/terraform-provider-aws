// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appfabric"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
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
	framework.WithImportByID
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
			"app_bundle_arn": framework.ARNAttributeComputedOnly(),
			"app_bundle_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"ingestion_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrState: schema.StringAttribute{
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

	input := &appfabric.CreateIngestionInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateIngestion(ctx, input)
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

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set values for unknowns.
	plan.AppBundleArn = fwflex.StringToFramework(ctx, out.Ingestion.AppBundleArn)
	plan.AppBundleIdentifier = fwflex.StringToFramework(ctx, out.Ingestion.AppBundleArn)
	plan.ARN = fwflex.StringToFramework(ctx, out.Ingestion.Arn)
	plan.State = fwflex.StringToFramework(ctx, aws.String(string(out.Ingestion.State)))
	plan.setID()

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceIngestion) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	conn := r.Meta().AppFabricClient(ctx)

	var state resourceIngestionData

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := state.InitFromID(); err != nil {
		resp.Diagnostics.AddError("parsing resource ID", err.Error())
		return
	}

	out, err := findIngestionByTwoPartKey(ctx, conn, state.AppBundleIdentifier.ValueString(), state.ARN.ValueString())

	if tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.AppFabric, create.ErrActionReading, ResNameIngestion, state.App.ValueString())
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.Append(create.DiagErrorFramework(names.AppFabric, create.ErrActionReading, ResNameIngestion, state.App.ValueString(), err))
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// There is no update API, so this method is a no-op
func (r *resourceIngestion) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

func (r *resourceIngestion) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

const (
	ingestionResourceIDPartCount = 2
)

func (m *resourceIngestionData) InitFromID() error {
	id := m.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, ingestionResourceIDPartCount, false)

	if err != nil {
		return err
	}

	m.AppBundleIdentifier = types.StringValue(parts[0])
	m.ARN = types.StringValue(parts[1])

	return nil
}

func (m *resourceIngestionData) setID() {
	m.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{m.AppBundleIdentifier.ValueString(), m.ARN.ValueString()}, ingestionResourceIDPartCount, false)))
}

func findIngestionByTwoPartKey(ctx context.Context, conn *appfabric.Client, appBundleIdentifier string, arn string) (*awstypes.Ingestion, error) {
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
