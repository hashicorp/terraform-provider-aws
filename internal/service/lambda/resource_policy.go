// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lambda_resource_policy", name="Resource Policy")
// @ArnIdentity("resource_arn")
// @Testing(hasNoPreExistingResource=true)
// Ignore `policy` because JSON is not normalized during attribute comparison.
// @Testing(importIgnore="policy")
func newResourcePolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourcePolicyResource{}, nil
}

type resourcePolicyResource struct {
	framework.ResourceWithModel[resourcePolicyResourceModel]
	framework.WithImportByIdentity
}

func (r *resourcePolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrPolicy: schema.StringAttribute{
				CustomType:  fwtypes.IAMPolicyType,
				Required:    true,
				Description: "JSON-formatted resource-based policy document to attach to the Lambda resource. This replaces the entire policy, including any statements added with aws_lambda_permission.",
			},
			names.AttrResourceARN: schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Required:    true,
				Description: "ARN of the Lambda function, version, or alias to attach the resource-based policy to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"revision_id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the current revision of the policy. Changes on every update, since PutResourcePolicy always issues a new revision.",
			},
		},
	}
}

func (r *resourcePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var plan resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input lambda.PutResourcePolicyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := tfresource.RetryWhenIsA[*lambda.PutResourcePolicyOutput, *awstypes.ResourceConflictException](ctx, resourcePolicyPropagationTimeout, func(ctx context.Context) (*lambda.PutResourcePolicyOutput, error) {
		return conn.PutResourcePolicy(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, fwflex.StringValueFromFramework(ctx, plan.ResourceARN))
		return
	}

	// Set values for unknowns.
	plan.RevisionID = fwflex.StringToFramework(ctx, out.RevisionId)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourcePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var state resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	resourceARN := fwflex.StringValueFromFramework(ctx, state.ResourceARN)
	out, err := findResourcePolicyByARN(ctx, conn, resourceARN)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceARN)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourcePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var plan resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input lambda.PutResourcePolicyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := tfresource.RetryWhenIsA[*lambda.PutResourcePolicyOutput, *awstypes.ResourceConflictException](ctx, resourcePolicyPropagationTimeout, func(ctx context.Context) (*lambda.PutResourcePolicyOutput, error) {
		return conn.PutResourcePolicy(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, fwflex.StringValueFromFramework(ctx, plan.ResourceARN))
		return
	}

	// Set values for unknowns.
	plan.RevisionID = fwflex.StringToFramework(ctx, out.RevisionId)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourcePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var state resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	resourceARN := fwflex.StringValueFromFramework(ctx, state.ResourceARN)
	input := lambda.DeleteResourcePolicyInput{
		ResourceArn: aws.String(resourceARN),
	}
	_, err := conn.DeleteResourcePolicy(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, resourceARN)
		return
	}
}

func (r *resourcePolicyResource) flatten(ctx context.Context, out *lambda.GetResourcePolicyOutput, data *resourcePolicyResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, out, data)...)
	return diags
}

const resourcePolicyPropagationTimeout = 2 * time.Minute

func findResourcePolicyByARN(ctx context.Context, conn *lambda.Client, resourceARN string) (*lambda.GetResourcePolicyOutput, error) {
	input := lambda.GetResourcePolicyInput{
		ResourceArn: aws.String(resourceARN),
	}
	return findResourcePolicy(ctx, conn, &input)
}

func findResourcePolicy(ctx context.Context, conn *lambda.Client, input *lambda.GetResourcePolicyInput) (*lambda.GetResourcePolicyOutput, error) {
	out, err := conn.GetResourcePolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Policy == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type resourcePolicyResourceModel struct {
	framework.WithRegionModel
	Policy      fwtypes.IAMPolicy `tfsdk:"policy"`
	ResourceARN fwtypes.ARN       `tfsdk:"resource_arn"`
	RevisionID  types.String      `tfsdk:"revision_id"`
}
