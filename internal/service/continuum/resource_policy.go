// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/continuum"
	awstypes "github.com/aws/aws-sdk-go-v2/service/continuum/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_continuum_resource_policy", name="Resource Policy")
// @ArnIdentity("resource_arn", identityDuplicateAttributes="id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(identityTest=false)
// @Testing(tagsTest=false)
// @Testing(importIgnore="policy")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/continuum;continuum.GetResourcePolicyOutput")
func newResourcePolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourcePolicyResource{}, nil
}

type resourcePolicyResource struct {
	framework.ResourceWithModel[resourcePolicyResourceModel]
	framework.WithImportByIdentity
}

func (r *resourcePolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
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
			"revision_id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *resourcePolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	arn := plan.ResourceARN.ValueString()
	input := continuum.PutResourcePolicyInput{
		Arn:    aws.String(arn),
		Policy: plan.Policy.ValueStringPointer(),
	}

	output, err := conn.PutResourcePolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}
	if output == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, arn)
		return
	}

	plan.ID = types.StringValue(arn)
	plan.RevisionID = types.StringPointerValue(output.RevisionId)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *resourcePolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	arn := state.ResourceARN.ValueString()
	output, err := findResourcePolicyByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}

	state.ID = types.StringValue(arn)
	state.Policy = fwtypes.IAMPolicyValue(aws.ToString(output.Policy))
	state.RevisionID = types.StringPointerValue(output.RevisionId)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *resourcePolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	arn := plan.ResourceARN.ValueString()
	input := continuum.PutResourcePolicyInput{
		Arn:                aws.String(arn),
		ExpectedRevisionId: state.RevisionID.ValueStringPointer(),
		Policy:             plan.Policy.ValueStringPointer(),
	}

	output, err := conn.PutResourcePolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}
	if output == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, arn)
		return
	}

	plan.ID = types.StringValue(arn)
	plan.RevisionID = types.StringPointerValue(output.RevisionId)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *resourcePolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resourcePolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ContinuumClient(ctx)

	// The API has no delete operation; putting a policy without a document removes it.
	input := continuum.PutResourcePolicyInput{
		Arn: state.ResourceARN.ValueStringPointer(),
	}
	_, err := conn.PutResourcePolicy(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.ResourceARN.String())
		return
	}
}

func findResourcePolicyByARN(ctx context.Context, conn *continuum.Client, arn string) (*continuum.GetResourcePolicyOutput, error) {
	input := continuum.GetResourcePolicyInput{
		Arn: aws.String(arn),
	}

	output, err := conn.GetResourcePolicy(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}
	if err != nil {
		return nil, err
	}
	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}
	if aws.ToString(output.Policy) == "" {
		return nil, &retry.NotFoundError{
			Message: "empty resource policy",
		}
	}

	return output, nil
}

type resourcePolicyResourceModel struct {
	framework.WithRegionModel
	ID          types.String      `tfsdk:"id"`
	Policy      fwtypes.IAMPolicy `tfsdk:"policy"`
	ResourceARN fwtypes.ARN       `tfsdk:"resource_arn"`
	RevisionID  types.String      `tfsdk:"revision_id"`
}
