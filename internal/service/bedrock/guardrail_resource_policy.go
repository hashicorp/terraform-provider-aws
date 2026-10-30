// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @FrameworkResource("aws_bedrock_guardrail_resource_policy", name="Guardrail Resource Policy")
// @Testing(tagsTest=false)
func newGuardrailResourcePolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &guardrailResourcePolicyResource{}, nil
}

const (
	ResNameGuardrailResourcePolicy = "Guardrail Resource Policy"
)

type guardrailResourcePolicyResource struct {
	framework.ResourceWithModel[guardrailResourcePolicyResourceModel]
}

func (r *guardrailResourcePolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
		},
	}
}

func (r *guardrailResourcePolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data guardrailResourcePolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	resourceARN := data.ResourceARN.ValueString()
	input := &bedrock.PutResourcePolicyInput{
		ResourceArn:    aws.String(resourceARN),
		ResourcePolicy: aws.String(data.Policy.ValueString()),
	}

	_, err := conn.PutResourcePolicy(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Bedrock Guardrail Resource Policy (%s)", resourceARN), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *guardrailResourcePolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data guardrailResourcePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	resourceARN := data.ResourceARN.ValueString()
	policy, err := findGuardrailResourcePolicyByARN(ctx, conn, resourceARN)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Guardrail Resource Policy (%s)", resourceARN), err.Error())
		return
	}

	policyToSet, err := verify.PolicyToSet(data.Policy.ValueString(), aws.ToString(policy))
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Guardrail Resource Policy (%s)", resourceARN), err.Error())
		return
	}
	data.Policy = fwtypes.IAMPolicyValue(policyToSet)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *guardrailResourcePolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data guardrailResourcePolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	resourceARN := data.ResourceARN.ValueString()
	input := &bedrock.PutResourcePolicyInput{
		ResourceArn:    aws.String(resourceARN),
		ResourcePolicy: aws.String(data.Policy.ValueString()),
	}

	_, err := conn.PutResourcePolicy(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Bedrock Guardrail Resource Policy (%s)", resourceARN), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *guardrailResourcePolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data guardrailResourcePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	_, err := conn.DeleteResourcePolicy(ctx, &bedrock.DeleteResourcePolicyInput{
		ResourceArn: data.ResourceARN.ValueStringPointer(),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Guardrail Resource Policy (%s)", data.ResourceARN.ValueString()), err.Error())
		return
	}
}

func (r *guardrailResourcePolicyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("resource_arn"), request.ID)...)
}

type guardrailResourcePolicyResourceModel struct {
	framework.WithRegionModel
	Policy      fwtypes.IAMPolicy `tfsdk:"policy"`
	ResourceARN fwtypes.ARN       `tfsdk:"resource_arn"`
}

func findGuardrailResourcePolicyByARN(ctx context.Context, conn *bedrock.Client, arn string) (*string, error) {
	input := &bedrock.GetResourcePolicyInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetResourcePolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{LastError: err}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ResourcePolicy == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.ResourcePolicy, nil
}
