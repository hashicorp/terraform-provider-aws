// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networkfirewall_proxy_rule_group", name="Proxy Rule Group")
// @Tags(identifierAttribute="arn")
// @ArnIdentity(identityDuplicateAttributes="id")
// @ArnFormat("proxy-rule-group/{name}")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preIdentityVersion="v5.100.0")
func newResourceProxyRuleGroup(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceProxyRuleGroup{}

	return r, nil
}

type resourceProxyRuleGroup struct {
	framework.ResourceWithModel[resourceProxyRuleGroupModel]
	framework.WithImportByIdentity
}

func (r *resourceProxyRuleGroup) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrARN)),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"update_token": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceProxyRuleGroup) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var plan resourceProxyRuleGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &networkfirewall.CreateProxyRuleGroupInput{
		ProxyRuleGroupName: aws.String(plan.ProxyRuleGroupName.ValueString()),
		Tags:               getTagsIn(ctx),
	}

	if !plan.Description.IsNull() {
		input.Description = aws.String(plan.Description.ValueString())
	}

	out, err := conn.CreateProxyRuleGroup(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"creating Network Firewall Proxy Rule Group",
			err.Error(),
		)
		return
	}

	if out == nil || out.ProxyRuleGroup == nil {
		resp.Diagnostics.AddError(
			"creating Network Firewall Proxy Rule Group",
			"empty output",
		)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.ProxyRuleGroup, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = plan.ProxyRuleGroupArn
	plan.UpdateToken = flex.StringToFramework(ctx, out.UpdateToken)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceProxyRuleGroup) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var state resourceProxyRuleGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findProxyRuleGroupByARN(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"reading Network Firewall Proxy Rule Group",
			err.Error(),
		)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.ProxyRuleGroup, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	state.UpdateToken = flex.StringToFramework(ctx, out.UpdateToken)

	setTagsOut(ctx, out.ProxyRuleGroup.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceProxyRuleGroup) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceProxyRuleGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Tags are updated via the service package framework; only state sync is needed here.
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceProxyRuleGroup) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var state resourceProxyRuleGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &networkfirewall.DeleteProxyRuleGroupInput{
		ProxyRuleGroupArn: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteProxyRuleGroup(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			"deleting Network Firewall Proxy Rule Group",
			err.Error(),
		)
		return
	}
}

func findProxyRuleGroupByARN(ctx context.Context, conn *networkfirewall.Client, arn string) (*networkfirewall.DescribeProxyRuleGroupOutput, error) {
	input := &networkfirewall.DescribeProxyRuleGroupInput{
		ProxyRuleGroupArn: aws.String(arn),
	}

	output, err := conn.DescribeProxyRuleGroup(ctx, input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.ProxyRuleGroup == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	if output.ProxyRuleGroup.DeleteTime != nil {
		return nil, &retry.NotFoundError{
			Message: "resource is deleted",
		}
	}

	return output, nil
}

type resourceProxyRuleGroupModel struct {
	framework.WithRegionModel
	Description        types.String `tfsdk:"description"`
	ID                 types.String `tfsdk:"id"`
	ProxyRuleGroupArn  types.String `tfsdk:"arn"`
	ProxyRuleGroupName types.String `tfsdk:"name"`
	Tags               tftags.Map   `tfsdk:"tags"`
	TagsAll            tftags.Map   `tfsdk:"tags_all"`
	UpdateToken        types.String `tfsdk:"update_token"`
}
