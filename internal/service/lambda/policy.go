// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lambda_policy", name="Policy")
// @ArnIdentity("resource_arn")
// @Testing(hasNoPreExistingResource=true)
// Ignore `policy` because JSON is not normalized during attribute comparison.
// @Testing(importIgnore="policy")
func newPolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &policyResource{}, nil
}

type policyResource struct {
	framework.ResourceWithModel[policyResourceModel]
	framework.WithImportByIdentity
}

func (r *policyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrResourceARN: schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN of the Lambda function, version, or alias to attach the resource-based policy to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrPolicy: schema.StringAttribute{
				CustomType:  fwtypes.IAMPolicyType,
				Required:    true,
				Description: "JSON-formatted resource-based policy document to attach to the Lambda resource. This replaces the entire policy, including any statements added with aws_lambda_permission.",
			},
			"revision_id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the current revision of the policy. Changes on every update, since PutResourcePolicy always issues a new revision.",
			},
		},
	}
}

func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var plan policyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	input := lambda.PutResourcePolicyInput{
		Policy:      plan.Policy.ValueStringPointer(),
		ResourceArn: plan.ResourceARN.ValueStringPointer(),
	}

	out, err := tfresource.RetryWhenIsA[*lambda.PutResourcePolicyOutput, *awstypes.ResourceConflictException](ctx, resourcePolicyPropagationTimeout, func(ctx context.Context) (*lambda.PutResourcePolicyOutput, error) {
		return conn.PutResourcePolicy(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, names.AttrResourceARN, plan.ResourceARN.String())
		return
	}
	if out == nil || out.Policy == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), names.AttrResourceARN, plan.ResourceARN.String())
		return
	}
	plan.Policy = fwtypes.IAMPolicyValue(aws.ToString(out.Policy))
	plan.RevisionID = flex.StringToFramework(ctx, out.RevisionId)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var state policyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findResourcePolicyByARN(ctx, conn, state.ResourceARN.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, names.AttrResourceARN, state.ResourceARN.String())
		return
	}

	state.Policy = fwtypes.IAMPolicyValue(aws.ToString(out.Policy))
	state.RevisionID = flex.StringToFramework(ctx, out.RevisionId)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var plan policyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	input := lambda.PutResourcePolicyInput{
		Policy:      plan.Policy.ValueStringPointer(),
		ResourceArn: plan.ResourceARN.ValueStringPointer(),
	}

	out, err := tfresource.RetryWhenIsA[*lambda.PutResourcePolicyOutput, *awstypes.ResourceConflictException](ctx, resourcePolicyPropagationTimeout, func(ctx context.Context) (*lambda.PutResourcePolicyOutput, error) {
		return conn.PutResourcePolicy(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, names.AttrResourceARN, plan.ResourceARN.String())
		return
	}
	if out == nil || out.Policy == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), names.AttrResourceARN, plan.ResourceARN.String())
		return
	}
	plan.Policy = fwtypes.IAMPolicyValue(aws.ToString(out.Policy))
	plan.RevisionID = flex.StringToFramework(ctx, out.RevisionId)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var state policyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := lambda.DeleteResourcePolicyInput{
		ResourceArn: state.ResourceARN.ValueStringPointer(),
	}

	_, err := conn.DeleteResourcePolicy(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, names.AttrResourceARN, state.ResourceARN.String())
		return
	}
}

const resourcePolicyPropagationTimeout = 2 * time.Minute

func findResourcePolicyByARN(ctx context.Context, conn *lambda.Client, resourceARN string) (*lambda.GetResourcePolicyOutput, error) {
	input := lambda.GetResourcePolicyInput{
		ResourceArn: aws.String(resourceARN),
	}

	out, err := conn.GetResourcePolicy(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Policy == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type policyResourceModel struct {
	framework.WithRegionModel
	Policy      fwtypes.IAMPolicy `tfsdk:"policy"`
	ResourceARN fwtypes.ARN       `tfsdk:"resource_arn"`
	RevisionID  types.String      `tfsdk:"revision_id"`
}
