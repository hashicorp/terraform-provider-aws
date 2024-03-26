// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Maximum amount of time to wait for Thing operation eventual consistency
	ReadPolicyTimeOut = 1 * time.Minute
)

// @FrameworkResource(name="Resource Policy")
func newResourceResourcePolicy(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceResourcePolicy{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameResourcePolicy = "Resource Policy"
)

type resourceResourcePolicy struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceResourcePolicy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_dynamodb_resource_policy"
}

func (r *resourceResourcePolicy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": framework.IDAttribute(),
			"policy": schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
			},
			"confirm_remove_self_resource_access": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"revision_id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *resourceResourcePolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DynamoDBClient(ctx)

	var plan resourceResourcePolicyData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &dynamodb.PutResourcePolicyInput{
		Policy:                          aws.String(plan.Policy.ValueString()),
		ResourceArn:                     aws.String(plan.ARN.ValueString()),
		ConfirmRemoveSelfResourceAccess: plan.ConfirmRemoveSelfResourceAccess.ValueBool(),
	}

	out, err := conn.PutResourcePolicy(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DynamoDB, create.ErrActionCreating, ResNameResourcePolicy, plan.ARN.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DynamoDB, create.ErrActionCreating, ResNameResourcePolicy, plan.ARN.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.RevisionId = flex.StringToFramework(ctx, out.RevisionId)
	plan.ID = flex.StringToFramework(ctx, plan.ARN.ValueStringPointer())
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceResourcePolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DynamoDBClient(ctx)

	var state resourceResourcePolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := retry.RetryContext(ctx, ReadPolicyTimeOut, func() *retry.RetryError {
		out, err := conn.GetResourcePolicy(ctx, &dynamodb.GetResourcePolicyInput{
			ResourceArn: aws.String(state.ID.ValueString()),
		})

		// If a policy is initially created and then immediately read, it may not be available.
		if errs.IsA[*awstypes.PolicyNotFoundException](err) ||
			errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		state.Policy = fwtypes.IAMPolicyValue(aws.ToString(out.Policy))
		state.RevisionId = flex.StringToFramework(ctx, out.RevisionId)
		return nil
	})

	if tfresource.TimedOut(err) ||
		errs.IsA[*awstypes.PolicyNotFoundException](err) ||
		errs.IsA[*awstypes.ResourceNotFoundException](err) {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DynamoDB, create.ErrActionReading, ResNameResourcePolicy, state.ID.String(), err),
			err.Error(),
		)
	}

	arn, d := fwtypes.ARNValue(state.ID.ValueString())
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ARN = arn

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceResourcePolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().DynamoDBClient(ctx)

	var plan, state resourceResourcePolicyData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
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
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DynamoDB, create.ErrActionUpdating, ResNameResourcePolicy, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.DynamoDB, create.ErrActionUpdating, ResNameResourcePolicy, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		plan.RevisionId = flex.StringToFramework(ctx, out.RevisionId)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceResourcePolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DynamoDBClient(ctx)

	var state resourceResourcePolicyData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
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
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DynamoDB, create.ErrActionDeleting, ResNameResourcePolicy, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceResourcePolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type resourceResourcePolicyData struct {
	ARN                             fwtypes.ARN       `tfsdk:"resource_arn"`
	Policy                          fwtypes.IAMPolicy `tfsdk:"policy"`
	ID                              types.String      `tfsdk:"id"`
	RevisionId                      types.String      `tfsdk:"revision_id"`
	ConfirmRemoveSelfResourceAccess types.Bool        `tfsdk:"confirm_remove_self_resource_access"`
}
