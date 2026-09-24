// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ssm

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ssm_resource_policy", name="Resource Policy")
func newResourcePolicyResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePolicyResource{}

	return r, nil
}

type resourcePolicyResource struct {
	framework.ResourceWithModel[resourcePolicyResourceModel]
}

func (r *resourcePolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrPolicy: schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
			},
			"policy_hash": schema.StringAttribute{
				Computed: true,
			},
			"policy_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

	conn := r.Meta().SSMClient(ctx)

	// SSM PutResourcePolicy validates `policy` against ^(?!\s*$).+ which only matches single-line strings,
	// so multi-line JSON (e.g. from aws_iam_policy_document) is rejected. Compact it before sending.
	policy, err := tfjson.CompactString(data.Policy.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating SSM Resource Policy (%s)", data.ResourceARN.ValueString()), err.Error())

		return
	}

	input := &ssm.PutResourcePolicyInput{
		Policy:      aws.String(policy),
		ResourceArn: flex.StringFromFramework(ctx, data.ResourceARN),
	}

	output, err := conn.PutResourcePolicy(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating SSM Resource Policy (%s)", data.ResourceARN.ValueString()), err.Error())

		return
	}

	data.PolicyID = flex.StringToFramework(ctx, output.PolicyId)
	data.PolicyHash = flex.StringToFramework(ctx, output.PolicyHash)
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourcePolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourcePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSMClient(ctx)

	output, err := findResourcePolicyByTwoPartKey(ctx, conn, data.ResourceARN.ValueString(), data.PolicyID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SSM Resource Policy (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.Policy = fwtypes.IAMPolicyValue(aws.ToString(output.Policy))
	data.PolicyHash = flex.StringToFramework(ctx, output.PolicyHash)
	data.PolicyID = flex.StringToFramework(ctx, output.PolicyId)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourcePolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new resourcePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSMClient(ctx)

	policy, err := tfjson.CompactString(new.Policy.ValueString())
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating SSM Resource Policy (%s)", new.ID.ValueString()), err.Error())

		return
	}

	input := &ssm.PutResourcePolicyInput{
		Policy:      aws.String(policy),
		PolicyHash:  flex.StringFromFramework(ctx, old.PolicyHash),
		PolicyId:    flex.StringFromFramework(ctx, old.PolicyID),
		ResourceArn: flex.StringFromFramework(ctx, new.ResourceARN),
	}

	output, err := conn.PutResourcePolicy(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating SSM Resource Policy (%s)", new.ID.ValueString()), err.Error())

		return
	}

	new.PolicyID = flex.StringToFramework(ctx, output.PolicyId)
	new.PolicyHash = flex.StringToFramework(ctx, output.PolicyHash)

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourcePolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourcePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSMClient(ctx)

	input := ssm.DeleteResourcePolicyInput{
		PolicyHash:  flex.StringFromFramework(ctx, data.PolicyHash),
		PolicyId:    flex.StringFromFramework(ctx, data.PolicyID),
		ResourceArn: flex.StringFromFramework(ctx, data.ResourceARN),
	}

	_, err := conn.DeleteResourcePolicy(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.ResourcePolicyNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting SSM Resource Policy (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *resourcePolicyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const importIDSeparator = ","

	parts := strings.Split(request.ID, importIDSeparator)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		response.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: <resource_arn>%s<policy_id>. Got: %q", importIDSeparator, request.ID),
		)
		return
	}

	resourceARN, policyID := parts[0], parts[1]

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrResourceARN), resourceARN)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("policy_id"), policyID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), resourcePolicyCreateID(resourceARN, policyID))...)
}

func findResourcePolicyByTwoPartKey(ctx context.Context, conn *ssm.Client, resourceARN, policyID string) (*awstypes.GetResourcePoliciesResponseEntry, error) {
	input := &ssm.GetResourcePoliciesInput{
		ResourceArn: aws.String(resourceARN),
	}

	pages := ssm.NewGetResourcePoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.ResourcePolicyNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, policy := range page.Policies {
			if aws.ToString(policy.PolicyId) == policyID {
				return &policy, nil
			}
		}
	}

	return nil, &retry.NotFoundError{}
}

func resourcePolicyCreateID(resourceARN, policyID string) string {
	return fmt.Sprintf("%s,%s", resourceARN, policyID)
}

type resourcePolicyResourceModel struct {
	framework.WithRegionModel
	ID          types.String      `tfsdk:"id"`
	Policy      fwtypes.IAMPolicy `tfsdk:"policy"`
	PolicyHash  types.String      `tfsdk:"policy_hash"`
	PolicyID    types.String      `tfsdk:"policy_id"`
	ResourceARN fwtypes.ARN       `tfsdk:"resource_arn"`
}

func (data *resourcePolicyResourceModel) setID() {
	data.ID = types.StringValue(resourcePolicyCreateID(data.ResourceARN.ValueString(), data.PolicyID.ValueString()))
}
