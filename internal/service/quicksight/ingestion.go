// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Ingestion")
func newResourceIngestion(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceIngestion{}, nil
}

const (
	ResNameIngestion = "Ingestion"
)

type resourceIngestion struct {
	framework.ResourceWithConfigure
}

func (r *resourceIngestion) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_quicksight_ingestion"
}

func (r *resourceIngestion) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	ingestionStatusType := fwtypes.StringEnumType[awstypes.IngestionStatus]()
	ingestionTypeType := fwtypes.StringEnumType[awstypes.IngestionType]()

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed: true,
			},
			"aws_account_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data_set_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": framework.IDAttribute(),
			"ingestion_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ingestion_status": schema.StringAttribute{
				CustomType: ingestionStatusType,
				Computed:   true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.IngestionStatus](),
				},
			},
			"ingestion_type": schema.StringAttribute{
				CustomType: ingestionTypeType,
				Required:   true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.IngestionType](),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceIngestion) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var plan resourceIngestionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.AWSAccountID.IsUnknown() || plan.AWSAccountID.IsNull() {
		plan.AWSAccountID = types.StringValue(r.Meta().AccountID)
	}
	plan.ID = types.StringValue(createIngestionID(plan.AWSAccountID.ValueString(), plan.DataSetID.ValueString(), plan.IngestionID.ValueString()))

	in := quicksight.CreateIngestionInput{
		AwsAccountId:  aws.String(plan.AWSAccountID.ValueString()),
		DataSetId:     aws.String(plan.DataSetID.ValueString()),
		IngestionId:   aws.String(plan.IngestionID.ValueString()),
		IngestionType: plan.IngestionType.ValueEnum(),
	}

	out, err := conn.CreateIngestion(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameIngestion, plan.IngestionID.String(), nil),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, ResNameIngestion, plan.IngestionID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}
	plan.ARN = flex.StringToFramework(ctx, out.Arn)
	plan.IngestionStatus = fwtypes.StringEnumValue[awstypes.IngestionStatus](out.IngestionStatus)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceIngestion) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceIngestionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindIngestionByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameIngestion, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.IngestionID = flex.StringToFramework(ctx, out.IngestionId)
	state.IngestionStatus = fwtypes.StringEnumValue[awstypes.IngestionStatus](out.IngestionStatus)

	// To support import, parse the ID for the component keys and set
	// individual values in state
	awsAccountID, dataSetID, _, err := ParseIngestionID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, ResNameIngestion, state.ID.String(), nil),
			err.Error(),
		)
		return
	}
	state.AWSAccountID = flex.StringValueToFramework(ctx, awsAccountID)
	state.DataSetID = flex.StringValueToFramework(ctx, dataSetID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// There is no update API, so this method is a no-op
func (r *resourceIngestion) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceIngestion) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceIngestionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.CancelIngestion(ctx, &quicksight.CancelIngestionInput{
		AwsAccountId: aws.String(state.AWSAccountID.ValueString()),
		DataSetId:    aws.String(state.DataSetID.ValueString()),
		IngestionId:  aws.String(state.IngestionID.ValueString()),
	})
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, ResNameIngestion, state.ID.String(), nil),
			err.Error(),
		)
	}
}

func (r *resourceIngestion) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func FindIngestionByID(ctx context.Context, conn *quicksight.Client, id string) (*awstypes.Ingestion, error) {
	awsAccountID, dataSetID, ingestionID, err := ParseIngestionID(id)
	if err != nil {
		return nil, err
	}

	in := &quicksight.DescribeIngestionInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSetId:    aws.String(dataSetID),
		IngestionId:  aws.String(ingestionID),
	}

	out, err := conn.DescribeIngestion(ctx, in)
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

func ParseIngestionID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, ",", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected AWS_ACCOUNT_ID,DATA_SET_ID,INGESTION_ID", id)
	}
	return parts[0], parts[1], parts[2], nil
}

func createIngestionID(awsAccountID, dataSetID, ingestionID string) string {
	return fmt.Sprintf("%s,%s,%s", awsAccountID, dataSetID, ingestionID)
}

type resourceIngestionData struct {
	ARN             types.String                                 `tfsdk:"arn"`
	AWSAccountID    types.String                                 `tfsdk:"aws_account_id"`
	DataSetID       types.String                                 `tfsdk:"data_set_id"`
	ID              types.String                                 `tfsdk:"id"`
	IngestionID     types.String                                 `tfsdk:"ingestion_id"`
	IngestionStatus fwtypes.StringEnum[awstypes.IngestionStatus] `tfsdk:"ingestion_status"`
	IngestionType   fwtypes.StringEnum[awstypes.IngestionType]   `tfsdk:"ingestion_type"`
}
