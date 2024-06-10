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
func newIngestionResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &ingestionResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameIngestion = "Ingestion"
)

type ingestionResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithImportByID
	framework.WithTimeouts
}

func (*ingestionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_appfabric_ingestion"
}

func (r *ingestionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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

func (r *ingestionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data ingestionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	input := &appfabric.CreateIngestionInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateIngestion(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppFabric, create.ErrActionCreating, ResNameIngestion, data.App.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Ingestion == nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppFabric, create.ErrActionCreating, ResNameIngestion, data.App.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set values for unknowns.
	data.AppBundleArn = fwflex.StringToFramework(ctx, out.Ingestion.AppBundleArn)
	data.AppBundleIdentifier = fwflex.StringToFramework(ctx, out.Ingestion.AppBundleArn)
	data.ARN = fwflex.StringToFramework(ctx, out.Ingestion.Arn)
	data.State = fwflex.StringToFramework(ctx, aws.String(string(out.Ingestion.State)))
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *ingestionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data ingestionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())
		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	out, err := findIngestionByTwoPartKey(ctx, conn, data.AppBundleIdentifier.ValueString(), data.ARN.ValueString())

	if tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.AppFabric, create.ErrActionReading, ResNameIngestion, data.App.ValueString())
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.Append(create.DiagErrorFramework(names.AppFabric, create.ErrActionReading, ResNameIngestion, data.App.ValueString(), err))
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *ingestionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data ingestionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppFabricClient(ctx)

	in := &appfabric.DeleteIngestionInput{
		AppBundleIdentifier: aws.String(data.AppBundleIdentifier.ValueString()),
		IngestionIdentifier: aws.String(data.ARN.ValueString()),
	}

	_, err := conn.DeleteIngestion(ctx, in)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppFabric, create.ErrActionDeleting, ResNameIngestion, data.ARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *ingestionResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

const (
	ingestionResourceIDPartCount = 2
)

func (m *ingestionResourceModel) InitFromID() error {
	id := m.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, ingestionResourceIDPartCount, false)

	if err != nil {
		return err
	}

	m.AppBundleIdentifier = types.StringValue(parts[0])
	m.ARN = types.StringValue(parts[1])

	return nil
}

func (m *ingestionResourceModel) setID() {
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

type ingestionResourceModel struct {
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
