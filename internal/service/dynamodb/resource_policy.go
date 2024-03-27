// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Maximum amount of time to wait for Thing operation eventual consistency
	ReadPolicyTimeOut = 1 * time.Minute
)

// @FrameworkResource(name="Resource Policy")
func newResourcePolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePolicyResource{}

	return r, nil
}

type resourcePolicyResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (*resourcePolicyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_dynamodb_resource_policy"
}

func (r *resourcePolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"confirm_remove_self_resource_access": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			names.AttrID: framework.IDAttribute(),
			"policy": schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
			},
			"resource_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"revision_id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *resourcePolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourcePolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DynamoDBClient(ctx)

	input := &dynamodb.PutResourcePolicyInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.PutResourcePolicy(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating DynamoDB Resource Policy (%s)", data.ResourceARN.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.RevisionID = fwflex.StringToFramework(ctx, output.RevisionId)
	data.ID = flex.StringToFramework(ctx, data.ARN.ValueStringPointer())
	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *resourcePolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().DynamoDBClient(ctx)

	var state resourcePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	err := retry.RetryContext(ctx, ReadPolicyTimeOut, func() *retry.RetryError {
		out, err := conn.GetResourcePolicy(ctx, &dynamodb.GetResourcePolicyInput{
			ResourceArn: aws.String(state.ID.ValueString()),
		})

		// If a policy is initially created and then immediately read, it may not be available.
		if errs.IsA[*awstypes.PolicyNotFoundException](err) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		state.Policy = fwtypes.IAMPolicyValue(aws.ToString(out.Policy))
		state.RevisionID = flex.StringToFramework(ctx, out.RevisionId)
		return nil
	})

	// if the dynamodb table gets removed, a Resource not found will be thrown.
	if errs.IsA[*awstypes.PolicyNotFoundException](err) ||
		errs.IsA[*awstypes.ResourceNotFoundException](err) {
		response.State.RemoveResource(ctx)
		return
	}

	if tfresource.TimedOut(err) {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DynamoDB, create.ErrActionReading, ResNameResourcePolicy, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DynamoDB, create.ErrActionReading, ResNameResourcePolicy, state.ID.String(), err),
			err.Error(),
		)
	}

	arn, d := fwtypes.ARNValue(state.ID.ValueString())
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}
	state.ARN = arn

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourcePolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().DynamoDBClient(ctx)

	var plan, state resourcePolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Policy.Equal(state.Policy) || !plan.ConfirmRemoveSelfResourceAccess.Equal(state.ConfirmRemoveSelfResourceAccess) {
		in := dynamodb.PutResourcePolicyInput{
			Policy:                          aws.String(plan.Policy.ValueString()),
			ResourceArn:                     aws.String(plan.ARN.ValueString()),
			ConfirmRemoveSelfResourceAccess: plan.ConfirmRemoveSelfResourceAccess.ValueBool(),
		}
		out, err := conn.PutResourcePolicy(ctx, &in)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DynamoDB, create.ErrActionUpdating, ResNameResourcePolicy, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DynamoDB, create.ErrActionUpdating, ResNameResourcePolicy, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		plan.RevisionID = flex.StringToFramework(ctx, out.RevisionId)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourcePolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().DynamoDBClient(ctx)

	var state resourcePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	in := &dynamodb.DeleteResourcePolicyInput{
		ResourceArn: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteResourcePolicy(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.PolicyNotFoundException](err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DynamoDB, create.ErrActionDeleting, ResNameResourcePolicy, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findResourcePolicyByARN(ctx context.Context, conn *dynamodb.Client, arn string) (*dynamodb.GetResourcePolicyOutput, error) {
	input := &dynamodb.GetResourcePolicyInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetResourcePolicy(ctx, input)

	if errs.IsA[*awstypes.PolicyNotFoundException](err) || errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type resourcePolicyResourceModel struct {
	ConfirmRemoveSelfResourceAccess types.Bool        `tfsdk:"confirm_remove_self_resource_access"`
	ID                              types.String      `tfsdk:"id"`
	Policy                          fwtypes.IAMPolicy `tfsdk:"policy"`
	ResourceARN                     fwtypes.ARN       `tfsdk:"resource_arn"`
	RevisionID                      types.String      `tfsdk:"revision_id"`
}

func (data *resourcePolicyResourceModel) InitFromID() error {
	data.ResourceARN = fwtypes.ARNValue
}
