// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networkfirewall_proxy_configuration", name="Proxy Configuration")
// @Tags(identifierAttribute="arn")
// @ArnIdentity(identityDuplicateAttributes="id")
// @ArnFormat("proxy-configuration/{name}")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preIdentityVersion="v5.100.0")
func newResourceProxyConfiguration(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceProxyConfiguration{}

	return r, nil
}

type resourceProxyConfiguration struct {
	framework.ResourceWithModel[resourceProxyConfigurationModel]
	framework.WithImportByIdentity
}

func (r *resourceProxyConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
		Blocks: map[string]schema.Block{
			"default_rule_phase_actions": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[defaultRulePhaseActionsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"post_response": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ProxyRulePhaseAction](),
							Required:   true,
						},
						"pre_dns": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ProxyRulePhaseAction](),
							Required:   true,
						},
						"pre_request": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ProxyRulePhaseAction](),
							Required:   true,
						},
					},
				},
			},
		},
	}
}

func (r *resourceProxyConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var data resourceProxyConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input networkfirewall.CreateProxyConfigurationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, data, &input, flex.WithFieldNamePrefix("ProxyConfiguration")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateProxyConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ProxyConfigurationName.String())
		return
	}
	if out == nil || out.ProxyConfiguration == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, data.ProxyConfigurationName.String())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.ProxyConfiguration, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = data.ProxyConfigurationArn
	data.UpdateToken = flex.StringToFramework(ctx, out.UpdateToken)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, data))
}

func (r *resourceProxyConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var state resourceProxyConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findProxyConfigurationByARN(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.ProxyConfiguration, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	state.UpdateToken = flex.StringToFramework(ctx, out.UpdateToken)

	setTagsOut(ctx, out.ProxyConfiguration.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceProxyConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var plan, state resourceProxyConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state, flex.WithIgnoredField("Tags"), flex.WithIgnoredField("TagsAll"))
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input networkfirewall.UpdateProxyConfigurationInput
		input.UpdateToken = state.UpdateToken.ValueStringPointer()

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan.DefaultRulePhaseActions, &input.DefaultRulePhaseActions))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateProxyConfiguration(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
			return
		}
		if out == nil || out.ProxyConfiguration == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, state.ID.String())
			return
		}

		plan.UpdateToken = flex.StringToFramework(ctx, out.UpdateToken)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceProxyConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var state resourceProxyConfigurationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := networkfirewall.DeleteProxyConfigurationInput{
		ProxyConfigurationArn: state.ProxyConfigurationArn.ValueStringPointer(),
	}

	_, err := conn.DeleteProxyConfiguration(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}
}

type resourceProxyConfigurationModel struct {
	framework.WithRegionModel
	DefaultRulePhaseActions fwtypes.ListNestedObjectValueOf[defaultRulePhaseActionsModel] `tfsdk:"default_rule_phase_actions"`
	Description             types.String                                                  `tfsdk:"description"`
	ID                      types.String                                                  `tfsdk:"id"`
	ProxyConfigurationArn   types.String                                                  `tfsdk:"arn"`
	ProxyConfigurationName  types.String                                                  `tfsdk:"name"`
	Tags                    tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                    `tfsdk:"tags_all"`
	UpdateToken             types.String                                                  `tfsdk:"update_token"`
}

type defaultRulePhaseActionsModel struct {
	PostResponse fwtypes.StringEnum[awstypes.ProxyRulePhaseAction] `tfsdk:"post_response"`
	PreDNS       fwtypes.StringEnum[awstypes.ProxyRulePhaseAction] `tfsdk:"pre_dns"`
	PreRequest   fwtypes.StringEnum[awstypes.ProxyRulePhaseAction] `tfsdk:"pre_request"`
}

func findProxyConfigurationByARN(ctx context.Context, conn *networkfirewall.Client, arn string) (*networkfirewall.DescribeProxyConfigurationOutput, error) {
	input := networkfirewall.DescribeProxyConfigurationInput{
		ProxyConfigurationArn: aws.String(arn),
	}

	out, err := conn.DescribeProxyConfiguration(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.ProxyConfiguration == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	if out.ProxyConfiguration.DeleteTime != nil {
		return nil, smarterr.NewError(&retry.NotFoundError{
			Message: "resource is deleted",
		})
	}

	return out, nil
}
