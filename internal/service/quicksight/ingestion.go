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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_ingestion", name="Ingestion")
func newIngestionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &ingestionResource{}, nil
}

const (
	resNameIngestion = "Ingestion"
)

type ingestionResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithImportByID
}

func (r *ingestionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			names.AttrAWSAccountID: schema.StringAttribute{
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
			names.AttrID: framework.IDAttribute(),
			"ingestion_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ingestion_status": schema.StringAttribute{
				Computed: true,
			},
			"ingestion_type": schema.StringAttribute{
				Required: true,
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

func (r *ingestionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var plan resourceIngestionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.AWSAccountID.IsUnknown() || plan.AWSAccountID.IsNull() {
		plan.AWSAccountID = types.StringValue(r.Meta().AccountID(ctx))
	}
	awsAccountID, dataSetID, ingestionID := flex.StringValueFromFramework(ctx, plan.AWSAccountID), flex.StringValueFromFramework(ctx, plan.DataSetID), flex.StringValueFromFramework(ctx, plan.IngestionID)
	in := quicksight.CreateIngestionInput{
		AwsAccountId:  aws.String(awsAccountID),
		DataSetId:     aws.String(dataSetID),
		IngestionId:   aws.String(ingestionID),
		IngestionType: awstypes.IngestionType(plan.IngestionType.ValueString()),
	}

	out, err := conn.CreateIngestion(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, resNameIngestion, plan.IngestionID.String(), nil),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionCreating, resNameIngestion, plan.IngestionID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringValueToFramework(ctx, ingestionCreateResourceID(awsAccountID, dataSetID, ingestionID))
	plan.ARN = flex.StringToFramework(ctx, out.Arn)
	plan.IngestionStatus = flex.StringValueToFramework(ctx, out.IngestionStatus)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *ingestionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceIngestionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, dataSetID, ingestionID, err := ingestionParseResourceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionReading, resNameIngestion, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	out, err := findIngestionByThreePartKey(ctx, conn, awsAccountID, dataSetID, ingestionID)
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionSetting, resNameIngestion, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.IngestionID = flex.StringToFramework(ctx, out.IngestionId)
	state.IngestionStatus = flex.StringValueToFramework(ctx, out.IngestionStatus)
	state.AWSAccountID = flex.StringValueToFramework(ctx, awsAccountID)
	state.DataSetID = flex.StringValueToFramework(ctx, dataSetID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ingestionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().QuickSightClient(ctx)

	var state resourceIngestionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	awsAccountID, dataSetID, ingestionID, err := ingestionParseResourceID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionReading, resNameIngestion, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	_, err = conn.CancelIngestion(ctx, &quicksight.CancelIngestionInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSetId:    aws.String(dataSetID),
		IngestionId:  aws.String(ingestionID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionDeleting, resNameIngestion, state.ID.String(), nil),
			err.Error(),
		)
	}
}

func findIngestionByThreePartKey(ctx context.Context, conn *quicksight.Client, awsAccountID, dataSetID, ingestionID string) (*awstypes.Ingestion, error) {
	input := &quicksight.DescribeIngestionInput{
		AwsAccountId: aws.String(awsAccountID),
		DataSetId:    aws.String(dataSetID),
		IngestionId:  aws.String(ingestionID),
	}

	return findIngestion(ctx, conn, input)
}

func findIngestion(ctx context.Context, conn *quicksight.Client, input *quicksight.DescribeIngestionInput) (*awstypes.Ingestion, error) {
	output, err := conn.DescribeIngestion(ctx, input)

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

const ingestionResourceIDSeparator = ","

func ingestionCreateResourceID(awsAccountID, dataSetID, ingestionID string) string {
	parts := []string{awsAccountID, dataSetID, ingestionID}
	id := strings.Join(parts, ingestionResourceIDSeparator)

	return id
}

func ingestionParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, ingestionResourceIDSeparator, 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%[1]s), expected AWS_ACCOUNT_ID%[2]sDATA_SET_ID%[2]sINGESTION_ID", id, ingestionResourceIDSeparator)
	}

	return parts[0], parts[1], parts[2], nil
}

type resourceIngestionData struct {
	ARN             types.String `tfsdk:"arn"`
	AWSAccountID    types.String `tfsdk:"aws_account_id"`
	DataSetID       types.String `tfsdk:"data_set_id"`
	ID              types.String `tfsdk:"id"`
	IngestionID     types.String `tfsdk:"ingestion_id"`
	IngestionStatus types.String `tfsdk:"ingestion_status"`
	IngestionType   types.String `tfsdk:"ingestion_type"`
}
