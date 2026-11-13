// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package osis

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/osis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/osis/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_osis_resource_policy", name="Resource Policy")
// @ArnIdentity("resource_arn")
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator="randomPipelineName(t)")
func newResourcePolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePolicyResource{}

	return r, nil
}

type resourcePolicyResource struct {
	framework.ResourceWithModel[resourcePolicyResourceModel]
	framework.WithImportByIdentity
}

func (r *resourcePolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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

func (r *resourcePolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourcePolicyResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchIngestionClient(ctx)

	input := osis.PutResourcePolicyInput{
		Policy:      fwflex.StringFromFramework(ctx, data.Policy),
		ResourceArn: fwflex.StringFromFramework(ctx, data.ResourceARN),
	}

	_, err := conn.PutResourcePolicy(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating OpenSearch Ingestion Resource Policy (%s)", data.ResourceARN.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourcePolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourcePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchIngestionClient(ctx)

	resourceArn := data.ResourceARN.ValueString()

	output, err := findResourcePolicyByResourceARN(ctx, conn, resourceArn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading OpenSearch Ingestion Resource Policy (%s)", resourceArn), err.Error())
		return
	}

	response.Diagnostics.Append(r.flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourcePolicyResource) flatten(ctx context.Context, output *osis.GetResourcePolicyOutput, data *resourcePolicyResourceModel) (diags diag.Diagnostics) {
	diags.Append(fwflex.Flatten(ctx, output, data)...)
	return diags
}

func (r *resourcePolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data resourcePolicyResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchIngestionClient(ctx)

	input := osis.PutResourcePolicyInput{
		Policy:      fwflex.StringFromFramework(ctx, data.Policy),
		ResourceArn: fwflex.StringFromFramework(ctx, data.ResourceARN),
	}

	_, err := conn.PutResourcePolicy(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating OpenSearch Ingestion Resource Policy (%s)", data.ResourceARN.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourcePolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourcePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchIngestionClient(ctx)

	input := osis.DeleteResourcePolicyInput{
		ResourceArn: fwflex.StringFromFramework(ctx, data.ResourceARN),
	}

	_, err := conn.DeleteResourcePolicy(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting OpenSearch Ingestion Resource Policy (%s)", data.ResourceARN.ValueString()), err.Error())
		return
	}
}

type resourcePolicyResourceModel struct {
	framework.WithRegionModel
	ResourceARN fwtypes.ARN       `tfsdk:"resource_arn"`
	Policy      fwtypes.IAMPolicy `tfsdk:"policy"`
}

// findResourcePolicyByResourceARN retrieves a resource policy by its resource ARN.
// Returns a NotFoundError if the policy is empty or the resource is not found.
func findResourcePolicyByResourceARN(ctx context.Context, conn *osis.Client, resourceARN string) (*osis.GetResourcePolicyOutput, error) {
	input := &osis.GetResourcePolicyInput{
		ResourceArn: aws.String(resourceARN),
	}

	output, err := conn.GetResourcePolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil || aws.ToString(output.Policy) == "{}" {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}
