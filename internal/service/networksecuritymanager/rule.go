// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networksecuritymanager_rule", name="Rule")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/networksecuritymanager;networksecuritymanager.GetRuleOutput")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheckRule")
// @Testing(identityRegionOverrideTest=false)
func newRuleResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &ruleResource{}, nil
}

type ruleResource struct {
	framework.ResourceWithModel[ruleResourceModel]
	framework.WithImportByIdentity
}

func (r *ruleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrConfiguration: schema.StringAttribute{
				CustomType: fwtypes.NewSmithyJSONType(ctx, document.NewLazyDocument),
				Required:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(256),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z _.:/=+\-@]*$`), ""),
				},
			},
			"firewall_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RuleFirewallType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9A-Za-z][0-9A-Za-z _.:/=+\-@]*$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RuleType](),
				Required:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *ruleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ruleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input networksecuritymanager.CreateRuleInput
	// The API's name, description and ARN fields are prefixed with "Rule".
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Rule")))
	if resp.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.IsPublished = aws.Bool(true)
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateRule(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenRule(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *ruleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ruleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	out, err := findRuleByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flattenRule(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *ruleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ruleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		// Updates are guarded by the token returned from the most recent read or write.
		current, err := findRuleByARN(ctx, conn, arn)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		var input networksecuritymanager.UpdateRuleInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Rule")))
		if resp.Diagnostics.HasError() {
			return
		}

		input.ClientToken = aws.String(create.UniqueId(ctx))
		// Always publish so that a Terraform-managed rule never carries a dangling draft.
		input.IsPublished = aws.Bool(true)
		input.RuleIdentifier = aws.String(arn)
		input.UpdateToken = current.UpdateToken

		out, err := conn.UpdateRule(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, flattenRule(ctx, out, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *ruleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ruleResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NetworkSecurityManagerClient(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ARN)
	input := networksecuritymanager.DeleteRuleInput{
		RuleIdentifier: aws.String(arn),
	}
	_, err := conn.DeleteRule(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}
}

func findRuleByARN(ctx context.Context, conn *networksecuritymanager.Client, arn string) (*networksecuritymanager.GetRuleOutput, error) {
	input := networksecuritymanager.GetRuleInput{
		RuleIdentifier: aws.String(arn),
	}

	out, err := conn.GetRule(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

// flattenRule flattens a CreateRule, GetRule or UpdateRule response into the model.
//
// The service returns every numeric value in a rule configuration as a string
// (for example "ImmunityTime": 300 is read back as "ImmunityTime": "300") while
// rejecting strings on input. When the returned configuration only differs from
// the model's value in that way, the model's value is kept so that the planned
// configuration does not appear to have changed.
func flattenRule(ctx context.Context, apiObject any, data *ruleResourceModel) diag.Diagnostics { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	prior := data.Configuration

	diags := fwflex.Flatten(ctx, apiObject, data, fwflex.WithFieldNamePrefix("Rule"))
	if diags.HasError() {
		return diags
	}

	if prior.IsNull() || prior.IsUnknown() || data.Configuration.IsNull() {
		return diags
	}

	if ruleConfigurationEquivalent(prior.ValueString(), data.Configuration.ValueString()) {
		data.Configuration = prior
	}

	return diags
}

// ruleConfigurationEquivalent reports whether two JSON documents are equal,
// treating a JSON number and a string containing the same decimal number as equal.
// Invalid JSON is never equivalent.
func ruleConfigurationEquivalent(a, b string) bool {
	va, err := decodeJSONUseNumber(a)
	if err != nil {
		return false
	}
	vb, err := decodeJSONUseNumber(b)
	if err != nil {
		return false
	}

	return jsonValuesEquivalent(va, vb)
}

func decodeJSONUseNumber(s string) (any, error) {
	dec := json.NewDecoder(bytes.NewReader([]byte(s)))
	dec.UseNumber()

	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}

	return v, nil
}

func jsonValuesEquivalent(a, b any) bool {
	switch a := a.(type) {
	case map[string]any:
		b, ok := b.(map[string]any)
		if !ok || len(a) != len(b) {
			return false
		}
		for k, va := range a {
			vb, ok := b[k]
			if !ok || !jsonValuesEquivalent(va, vb) {
				return false
			}
		}
		return true

	case []any:
		b, ok := b.([]any)
		if !ok || len(a) != len(b) {
			return false
		}
		for i := range a {
			if !jsonValuesEquivalent(a[i], b[i]) {
				return false
			}
		}
		return true

	case json.Number:
		return jsonNumberEquivalent(a, b)

	case string:
		if n, ok := b.(json.Number); ok {
			return jsonNumberEquivalent(n, a)
		}
		b, ok := b.(string)
		return ok && a == b

	case bool:
		b, ok := b.(bool)
		return ok && a == b

	case nil:
		return b == nil
	}

	return false
}

// jsonNumberEquivalent reports whether v is a JSON number, or a string, with the same numeric value as n.
func jsonNumberEquivalent(n json.Number, v any) bool {
	var s string
	switch v := v.(type) {
	case json.Number:
		s = v.String()
	case string:
		s = v
	default:
		return false
	}

	x, ok := new(big.Rat).SetString(n.String())
	if !ok {
		return false
	}
	y, ok := new(big.Rat).SetString(s)
	if !ok {
		return false
	}

	return x.Cmp(y) == 0
}

type ruleResourceModel struct {
	framework.WithRegionModel
	ARN           types.String                                  `tfsdk:"arn"`
	Configuration fwtypes.SmithyJSON[document.Interface]        `tfsdk:"configuration"`
	Description   types.String                                  `tfsdk:"description"`
	FirewallType  fwtypes.StringEnum[awstypes.RuleFirewallType] `tfsdk:"firewall_type"`
	Name          types.String                                  `tfsdk:"name"`
	RuleType      fwtypes.StringEnum[awstypes.RuleType]         `tfsdk:"rule_type"`
	Tags          tftags.Map                                    `tfsdk:"tags"`
	TagsAll       tftags.Map                                    `tfsdk:"tags_all"`
}
