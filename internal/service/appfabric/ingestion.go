// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appfabric"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	uuid "github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Ingestion")
// @Tags(identifierAttribute="arn")
func newIngestionResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &ingestionResource{}

	return r, nil
}

type ingestionResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[ingestionResourceModel]
	framework.WithImportByID
}

func (*ingestionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_appfabric_ingestion"
}

func (r *ingestionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"app": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_bundle_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"ingestion_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.IngestionType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"tenant_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1024),
				},
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

	// Additional fields.
	input.AppBundleIdentifier = fwflex.StringFromFramework(ctx, data.AppBundleARN)
	input.ClientToken = aws.String(errs.Must(uuid.GenerateUUID()))
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateIngestion(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating AppFabric Ingestion", err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set values for unknowns.
	data.ARN = fwflex.StringToFramework(ctx, output.Ingestion.Arn)
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

	ingestion, err := findIngestionByTwoPartKey(ctx, conn, data.AppBundleARN.ValueString(), data.ARN.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading AppFabric Ingestion (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, ingestion, &data)...)
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

	_, err := conn.DeleteIngestion(ctx, &appfabric.DeleteIngestionInput{
		AppBundleIdentifier: aws.String(data.AppBundleARN.ValueString()),
		IngestionIdentifier: aws.String(data.ARN.ValueString()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting AppFabric Ingestion (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *ingestionResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findIngestionByTwoPartKey(ctx context.Context, conn *appfabric.Client, appBundleARN, arn string) (*awstypes.Ingestion, error) {
	input := &appfabric.GetIngestionInput{
		AppBundleIdentifier: aws.String(appBundleARN),
		IngestionIdentifier: aws.String(arn),
	}

	output, err := conn.GetIngestion(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Ingestion == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Ingestion, nil
}

type ingestionResourceModel struct {
	App           types.String                               `tfsdk:"app"`
	AppBundleARN  fwtypes.ARN                                `tfsdk:"app_bundle_arn"`
	ARN           types.String                               `tfsdk:"arn"`
	ID            types.String                               `tfsdk:"id"`
	IngestionType fwtypes.StringEnum[awstypes.IngestionType] `tfsdk:"ingestion_type"`
	Tags          types.Map                                  `tfsdk:"tags"`
	TagsAll       types.Map                                  `tfsdk:"tags_all"`
	TenantId      types.String                               `tfsdk:"tenant_id"`
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

	m.AppBundleARN = fwtypes.ARNValue(parts[0])
	m.ARN = types.StringValue(parts[1])

	return nil
}

func (m *ingestionResourceModel) setID() {
	m.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{m.AppBundleARN.ValueString(), m.ARN.ValueString()}, ingestionResourceIDPartCount, false)))
}
