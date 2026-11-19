// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrock_guardrail_resource_policy", name="Guardrail Resource Policy")
// @ArnIdentity("resource_arn")
// @Testing(hasNoPreExistingResource=true)
// Ignore `policy` because JSON is not normalized during attribute comparison.
// @Testing(importIgnore="policy")
func newGuardrailResourcePolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &guardrailResourcePolicyResource{}, nil
}

const (
	ResNameGuardrailResourcePolicy = "Guardrail Resource Policy"
)

type guardrailResourcePolicyResource struct {
	framework.ResourceWithModel[guardrailResourcePolicyResourceModel]
	framework.WithImportByIdentity
}

func (r *guardrailResourcePolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrPolicy: schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
			},
			names.AttrResourceARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *guardrailResourcePolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data guardrailResourcePolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	resourceARN := data.ResourceARN.ValueString()
	// PutResourcePolicy rejects whitespace characters such as newlines, which aws_iam_policy_document output contains.
	policy, err := tfjson.CompactString(data.Policy.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, resourceARN)
		return
	}
	input := bedrock.PutResourcePolicyInput{
		ResourceArn:    aws.String(resourceARN),
		ResourcePolicy: aws.String(policy),
	}

	_, err = conn.PutResourcePolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, resourceARN)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *guardrailResourcePolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data guardrailResourcePolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	resourceARN := data.ResourceARN.ValueString()
	policy, err := findGuardrailResourcePolicyByARN(ctx, conn, resourceARN)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err), smerr.ID, resourceARN)
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, resourceARN)
		return
	}

	policyToSet, err := verify.PolicyToSet(data.Policy.ValueString(), aws.ToString(policy))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, resourceARN)
		return
	}
	data.Policy = fwtypes.IAMPolicyValue(policyToSet)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *guardrailResourcePolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data guardrailResourcePolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	resourceARN := data.ResourceARN.ValueString()
	// PutResourcePolicy rejects whitespace characters such as newlines, which aws_iam_policy_document output contains.
	policy, err := tfjson.CompactString(data.Policy.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, resourceARN)
		return
	}
	input := bedrock.PutResourcePolicyInput{
		ResourceArn:    aws.String(resourceARN),
		ResourcePolicy: aws.String(policy),
	}

	_, err = conn.PutResourcePolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, resourceARN)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *guardrailResourcePolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data guardrailResourcePolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	input := bedrock.DeleteResourcePolicyInput{
		ResourceArn: data.ResourceARN.ValueStringPointer(),
	}
	_, err := conn.DeleteResourcePolicy(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ResourceARN.ValueString())
		return
	}
}

type guardrailResourcePolicyResourceModel struct {
	framework.WithRegionModel
	Policy      fwtypes.IAMPolicy `tfsdk:"policy"`
	ResourceARN fwtypes.ARN       `tfsdk:"resource_arn"`
}

func findGuardrailResourcePolicyByARN(ctx context.Context, conn *bedrock.Client, arn string) (*string, error) {
	input := bedrock.GetResourcePolicyInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetResourcePolicy(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	// GetResourcePolicy returns "{}" rather than ResourceNotFoundException when no policy is attached.
	if policy := aws.ToString(output.ResourcePolicy); policy == "" || policy == "{}" {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.ResourcePolicy, nil
}
