// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_wafv2_rule_group_permission_policy", name="Rule Group Permission Policy")
// @ArnIdentity("resource_arn", identityDuplicateAttributes="id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/wafv2;wafv2.GetPermissionPolicyOutput")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importIgnore="policy")
func newRuleGroupPermissionPolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &ruleGroupPermissionPolicyResource{}

	return r, nil
}

type ruleGroupPermissionPolicyResource struct {
	framework.ResourceWithModel[ruleGroupPermissionPolicyResourceModel]
	framework.WithImportByIdentity
}

func (r *ruleGroupPermissionPolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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

func (r *ruleGroupPermissionPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data ruleGroupPermissionPolicyResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WAFV2Client(ctx)

	input := &wafv2.PutPermissionPolicyInput{
		Policy:      flex.StringFromFramework(ctx, data.Policy),
		ResourceArn: flex.StringFromFramework(ctx, data.ResourceARN),
	}

	_, err := conn.PutPermissionPolicy(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating WAFv2 Rule Group Permission Policy (%s)", data.ResourceARN.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *ruleGroupPermissionPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data ruleGroupPermissionPolicyResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WAFV2Client(ctx)

	output, err := findPermissionPolicyByARN(ctx, conn, data.ResourceARN.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading WAFv2 Rule Group Permission Policy (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// GetPermissionPolicy only returns Policy, not ResourceArn.
	data.Policy = fwtypes.IAMPolicyValue(aws.ToString(output.Policy))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *ruleGroupPermissionPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new ruleGroupPermissionPolicyResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WAFV2Client(ctx)

	input := &wafv2.PutPermissionPolicyInput{
		Policy:      flex.StringFromFramework(ctx, new.Policy),
		ResourceArn: flex.StringFromFramework(ctx, new.ResourceARN),
	}

	_, err := conn.PutPermissionPolicy(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating WAFv2 Rule Group Permission Policy (%s)", new.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *ruleGroupPermissionPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data ruleGroupPermissionPolicyResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().WAFV2Client(ctx)

	input := wafv2.DeletePermissionPolicyInput{
		ResourceArn: flex.StringFromFramework(ctx, data.ResourceARN),
	}

	_, err := conn.DeletePermissionPolicy(ctx, &input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting WAFv2 Rule Group Permission Policy (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findPermissionPolicyByARN(ctx context.Context, conn *wafv2.Client, resourceARN string) (*wafv2.GetPermissionPolicyOutput, error) {
	input := &wafv2.GetPermissionPolicyInput{
		ResourceArn: aws.String(resourceARN),
	}

	output, err := conn.GetPermissionPolicy(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type ruleGroupPermissionPolicyResourceModel struct {
	framework.WithRegionModel
	ID          types.String      `tfsdk:"id"`
	Policy      fwtypes.IAMPolicy `tfsdk:"policy"`
	ResourceARN fwtypes.ARN       `tfsdk:"resource_arn"`
}

func (data *ruleGroupPermissionPolicyResourceModel) setID() {
	data.ID = data.ResourceARN.StringValue
}
