// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatch_log_index_policy", name="Index Policy")
func newResourceIndexPolicy(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceIndexPolicy{}
	return r, nil
}

const (
	ResNameIndexPolicy = "Index Policy"
)

type resourceIndexPolicy struct {
	framework.ResourceWithConfigure
}

func (r *resourceIndexPolicy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_cloudwatch_log_index_policy"
}

func (r *resourceIndexPolicy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrLogGroupName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_document": schema.StringAttribute{
				Required:    true,
				Description: "Field index filter policy, in JSON",
				Validators: []validator.String{
					validators.JSON(),
				},
			},
		},
	}
}

func (r *resourceIndexPolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LogsClient(ctx)

	var plan resourceIndexPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cloudwatchlogs.PutIndexPolicyInput{
		LogGroupIdentifier: plan.LogGroupName.ValueStringPointer(),
		PolicyDocument:     plan.PolicyDocument.ValueStringPointer(),
	}

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("IndexPolicy"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.PutIndexPolicy(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionCreating, ResNameIndexPolicy, plan.LogGroupName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.IndexPolicy == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionCreating, ResNameIndexPolicy, plan.LogGroupName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	// Set resource ID
	//id := fmt.Sprintf("%s:%s", *out.IndexPolicy.LogGroupIdentifier, "index-policy")
	//plan.ID = flex.StringToFramework(ctx, &id)

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceIndexPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LogsClient(ctx)

	var state resourceIndexPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findIndexPolicyByLogGroupName(ctx, conn, state.LogGroupName.ValueString())

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionSetting, ResNameIndexPolicy, "", err),
			err.Error(),
		)
		return
	}

	//state.ID = flex.StringToFramework(ctx, state.ID.ValueStringPointer())

	logGroupName, err := logGroupArnToName(*out.LogGroupIdentifier)
	if err != nil {
		resp.Diagnostics.AddError("failed to parse log group name", err.Error())
	}
	state.LogGroupName = flex.StringToFramework(ctx, &logGroupName)
	state.PolicyDocument = flex.StringToFramework(ctx, out.PolicyDocument)

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceIndexPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LogsClient(ctx)

	var plan, state resourceIndexPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.PolicyDocument.Equal(state.PolicyDocument) {
		input := cloudwatchlogs.PutIndexPolicyInput{
			LogGroupIdentifier: plan.LogGroupName.ValueStringPointer(),
		}

		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Test"))...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.PutIndexPolicy(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Logs, create.ErrActionUpdating, ResNameIndexPolicy, plan.LogGroupName.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.IndexPolicy == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Logs, create.ErrActionUpdating, ResNameIndexPolicy, "", nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceIndexPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LogsClient(ctx)

	var state resourceIndexPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cloudwatchlogs.DeleteIndexPolicyInput{
		LogGroupIdentifier: state.LogGroupName.ValueStringPointer(),
	}

	_, err := conn.DeleteIndexPolicy(ctx, &input)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Logs, create.ErrActionDeleting, ResNameIndexPolicy, "", err),
			err.Error(),
		)
		return
	}
}

func (r *resourceIndexPolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrLogGroupName), req, resp)
}

func findIndexPolicyByLogGroupName(ctx context.Context, conn *cloudwatchlogs.Client, logGroupName string) (*awstypes.IndexPolicy, error) {
	in := &cloudwatchlogs.DescribeIndexPoliciesInput{
		LogGroupIdentifiers: []string{logGroupName},
	}

	out, err := conn.DescribeIndexPolicies(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.IndexPolicies == nil || len(out.IndexPolicies) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &out.IndexPolicies[0], nil
}

type resourceIndexPolicyModel struct {
	LogGroupName   types.String `tfsdk:"log_group_name"`
	PolicyDocument types.String `tfsdk:"policy_document"`
}
