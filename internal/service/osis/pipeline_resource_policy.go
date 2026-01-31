// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package osis

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/osis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/osis/types"
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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_osis_pipeline_resource_policy", name="Pipeline Resource Policy")
// @ArnIdentity("resource_arn", identityDuplicateAttributes="id")
// @Testing(importIgnore="policy")
func newPipelineResourcePolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &pipelineResourcePolicyResource{}

	return r, nil
}

func (r *pipelineResourcePolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
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

func (r *pipelineResourcePolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data pipelineResourcePolicyResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchIngestionClient(ctx)

	input := &osis.PutResourcePolicyInput{
		Policy:      fwflex.StringFromFramework(ctx, data.Policy),
		ResourceArn: fwflex.StringFromFramework(ctx, data.ResourceARN),
	}

	_, err := conn.PutResourcePolicy(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating OpenSearch Ingestion Resource Policy (%s)", data.ResourceARN.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = data.ResourceARN.StringValue

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *pipelineResourcePolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data pipelineResourcePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchIngestionClient(ctx)

	resourceArn := data.ResourceARN.ValueString()

	input := &osis.GetResourcePolicyInput{
		ResourceArn: aws.String(resourceArn),
	}

	output, err := conn.GetResourcePolicy(ctx, input)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading OpenSearch Ingestion Pipeline Resource Policy (%s)", resourceArn), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *pipelineResourcePolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data pipelineResourcePolicyResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchIngestionClient(ctx)

	input := &osis.PutResourcePolicyInput{
		Policy:      fwflex.StringFromFramework(ctx, data.Policy),
		ResourceArn: fwflex.StringFromFramework(ctx, data.ResourceARN),
	}

	_, err := conn.PutResourcePolicy(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating OpenSearch Ingestion Resource Policy (%s)", data.ResourceARN.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *pipelineResourcePolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data pipelineResourcePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchIngestionClient(ctx)

	input := &osis.DeleteResourcePolicyInput{
		ResourceArn: fwflex.StringFromFramework(ctx, data.ResourceARN),
	}

	_, err := conn.DeleteResourcePolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting OpenSearch Ingestion Pipeline Resource Policy (%s)", data.ID.ValueString()), err.Error())
		return
	}
}

type pipelineResourcePolicyResource struct {
	framework.ResourceWithModel[pipelineResourcePolicyResourceModel]
	framework.WithImportByIdentity
}

type pipelineResourcePolicyResourceModel struct {
	framework.WithRegionModel
	ID          types.String      `tfsdk:"id"`
	ResourceARN fwtypes.ARN       `tfsdk:"resource_arn"`
	Policy      fwtypes.IAMPolicy `tfsdk:"policy"`
}
