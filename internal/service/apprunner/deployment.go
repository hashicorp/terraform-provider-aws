// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apprunner/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Deployment")
func newDeploymentResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &deploymentResource{}

	r.SetDefaultCreateTimeout(20 * time.Minute)

	return r, nil
}

type deploymentResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithNoOpDelete
	framework.WithTimeouts
}

func (r *deploymentResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_apprunner_deployment"
}

func (r *deploymentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"operation_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *deploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data deploymentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppRunnerClient(ctx)

	serviceARN := data.ServiceARN.ValueString()
	input := &apprunner.StartDeploymentInput{
		ServiceArn: aws.String(serviceARN),
	}

	output, err := conn.StartDeployment(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("starting App Runner Deployment (%s)", serviceARN), err.Error())

		return
	}

	// Set values for unknowns.
	operationID := aws.ToString(output.OperationId)
	data.OperationID = types.StringValue(operationID)
	data.setID()

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)

	op, err := waitDeploymentSucceeded(ctx, conn, serviceARN, operationID, createTimeout)

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("waiting for App Runner Deployment (%s/%s)", serviceARN, operationID), err.Error())

		return
	}

	// Set values for unknowns.
	data.Status = fwflex.StringValueToFramework(ctx, op.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *deploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data deploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppRunnerClient(ctx)

	serviceARN, operationID := data.ServiceARN.ValueString(), data.OperationID.ValueString()
	output, err := findOperationByTwoPartKey(ctx, conn, serviceARN, operationID)

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading App Runner Deployment (%s/%s)", serviceARN, operationID), err.Error())

		return
	}

	data.Status = fwflex.StringValueToFramework(ctx, output.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findOperationByTwoPartKey(ctx context.Context, conn *apprunner.Client, serviceARN, operationID string) (*awstypes.OperationSummary, error) {
	input := &apprunner.ListOperationsInput{
		ServiceArn: aws.String(serviceARN),
	}

	return findOperation(ctx, conn, input, func(v *awstypes.OperationSummary) bool {
		return aws.ToString(v.Id) == operationID
	})
}

func findOperation(ctx context.Context, conn *apprunner.Client, input *apprunner.ListOperationsInput, filter tfslices.Predicate[*awstypes.OperationSummary]) (*awstypes.OperationSummary, error) {
	output, err := findOperations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findOperations(ctx context.Context, conn *apprunner.Client, input *apprunner.ListOperationsInput, filter tfslices.Predicate[*awstypes.OperationSummary]) ([]awstypes.OperationSummary, error) {
	var output []awstypes.OperationSummary

	pages := apprunner.NewListOperationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.OperationSummaryList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusOperation(ctx context.Context, conn *apprunner.Client, serviceARN, operationID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findOperationByTwoPartKey(ctx, conn, serviceARN, operationID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitDeploymentSucceeded(ctx context.Context, conn *apprunner.Client, serviceARN, operationID string, timeout time.Duration) (*awstypes.OperationSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.OperationStatusPending, awstypes.OperationStatusInProgress),
		Target:         enum.Slice(awstypes.OperationStatusSucceeded),
		Refresh:        statusOperation(ctx, conn, serviceARN, operationID),
		Timeout:        timeout,
		PollInterval:   30 * time.Second,
		NotFoundChecks: 30,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.OperationSummary); ok {
		return output, err
	}

	return nil, err
}

type deploymentResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	OperationID types.String   `tfsdk:"operation_id"`
	ServiceARN  fwtypes.ARN    `tfsdk:"service_arn"`
	Status      types.String   `tfsdk:"status"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

func (data *deploymentResourceModel) setID() {
	data.ID = data.OperationID
}
