// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package mailmanager

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mailmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mailmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_mailmanager_traffic_policy", name="Traffic Policy")
// @IdentityAttribute("id")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/mailmanager;mailmanager.GetTrafficPolicyOutput")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheck")
// @Testing(skipEmptyTags=true, skipNullTags=true)
func newTrafficPolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &trafficPolicyResource{}, nil
}

const ResNameTrafficPolicy = "Traffic Policy"

type trafficPolicyResource struct {
	framework.ResourceWithModel[trafficPolicyResourceModel]
	framework.WithImportByIdentity
}

func (r *trafficPolicyResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"created_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"default_action": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AcceptAction](),
				Required:   true,
			},
			names.AttrID: framework.IDAttribute(),
			"last_updated_timestamp": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"max_message_size_bytes": schema.Int32Attribute{
				Optional: true,
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_-]+$`), "must contain only alphanumeric characters, underscores, and hyphens"),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"policy_statement": policyStatementBlock(ctx),
		},
	}
}

func policyStatementBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[policyStatementModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtLeast(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"action": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.AcceptAction](),
					Required:   true,
				},
			},
			Blocks: map[string]schema.Block{"condition": conditionBlock(ctx)},
		},
	}
}

func conditionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[policyConditionModel](ctx),
		Validators: []validator.List{listvalidator.SizeAtLeast(1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"boolean_expression": booleanExpressionBlock(ctx),
				"ip_expression":      ipExpressionBlock(ctx),
				"ipv6_expression":    ipv6ExpressionBlock(ctx),
				"string_expression":  stringExpressionBlock(ctx),
				"tls_expression":     tlsExpressionBlock(ctx),
			},
		},
	}
}

func conditionUnionValidators(siblings ...string) []validator.List {
	paths := make([]path.Expression, len(siblings))
	for i, sibling := range siblings {
		paths[i] = path.MatchRelative().AtParent().AtName(sibling)
	}
	return []validator.List{listvalidator.SizeAtMost(1), listvalidator.ExactlyOneOf(paths...)}
}

func booleanExpressionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[booleanExpressionModel](ctx),
		Validators: conditionUnionValidators("ip_expression", "ipv6_expression", "string_expression", "tls_expression"),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"operator": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.IngressBooleanOperator](), Required: true},
			},
			Blocks: map[string]schema.Block{"evaluate": booleanEvaluateBlock(ctx)},
		},
	}
}

func booleanEvaluateBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[booleanEvaluateModel](ctx),
		Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
		NestedObject: schema.NestedBlockObject{
			Blocks: map[string]schema.Block{
				"analysis": schema.ListNestedBlock{
					CustomType:   fwtypes.NewListNestedObjectTypeOf[analysisModel](ctx),
					Validators:   conditionUnionValidators("is_in_address_list"),
					NestedObject: schema.NestedBlockObject{Attributes: analysisAttributes()},
				},
				"is_in_address_list": schema.ListNestedBlock{
					CustomType: fwtypes.NewListNestedObjectTypeOf[isInAddressListModel](ctx),
					Validators: conditionUnionValidators("analysis"),
					NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
						"address_lists": schema.ListAttribute{
							CustomType: fwtypes.ListOfStringType, ElementType: types.StringType, Required: true,
							Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
						},
						"attribute": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.IngressAddressListEmailAttribute](), Required: true},
					}},
				},
			},
		},
	}
}

func analysisAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"analyzer": schema.StringAttribute{CustomType: fwtypes.ARNType, Required: true, Validators: []validator.String{fwvalidators.ARN()}},
		"result_field": schema.StringAttribute{Required: true, Validators: []validator.String{
			stringvalidator.LengthBetween(1, 256),
			stringvalidator.RegexMatches(regexache.MustCompile(`^(addon\.)?[\sa-zA-Z0-9_]+$`), "must contain only spaces, alphanumeric characters, and underscores, optionally prefixed by addon."),
		}},
	}
}

func ipExpressionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ipExpressionModel](ctx),
		Validators: conditionUnionValidators("boolean_expression", "ipv6_expression", "string_expression", "tls_expression"),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"operator": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.IngressIpOperator](), Required: true},
				"values": schema.ListAttribute{
					CustomType: fwtypes.ListOfStringType, ElementType: types.StringType, Required: true,
					Validators: []validator.List{listvalidator.SizeAtLeast(1), listvalidator.ValueStringsAre(fwvalidators.IPv4CIDRNetworkAddress())},
				},
			},
			Blocks: map[string]schema.Block{"evaluate": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ipEvaluateModel](ctx),
				Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
				NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
					"attribute": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.IngressIpv4Attribute](), Required: true},
				}},
			}},
		},
	}
}

func ipv6ExpressionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ipv6ExpressionModel](ctx),
		Validators: conditionUnionValidators("boolean_expression", "ip_expression", "string_expression", "tls_expression"),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"operator": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.IngressIpOperator](), Required: true},
				"values": schema.ListAttribute{
					CustomType: fwtypes.ListOfStringType, ElementType: types.StringType, Required: true,
					Validators: []validator.List{listvalidator.SizeAtLeast(1), listvalidator.ValueStringsAre(fwvalidators.IPv6CIDRNetworkAddress())},
				},
			},
			Blocks: map[string]schema.Block{"evaluate": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[ipv6EvaluateModel](ctx),
				Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
				NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
					"attribute": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.IngressIpv6Attribute](), Required: true},
				}},
			}},
		},
	}
}

func stringExpressionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[stringExpressionModel](ctx),
		Validators: conditionUnionValidators("boolean_expression", "ip_expression", "ipv6_expression", "tls_expression"),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"operator": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.IngressStringOperator](), Required: true},
				"values":   schema.ListAttribute{CustomType: fwtypes.ListOfStringType, ElementType: types.StringType, Required: true, Validators: []validator.List{listvalidator.SizeAtLeast(1)}},
			},
			Blocks: map[string]schema.Block{"evaluate": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[stringEvaluateModel](ctx),
				Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{"attribute": schema.StringAttribute{
						CustomType: fwtypes.StringEnumType[awstypes.IngressStringEmailAttribute](), Optional: true,
						Validators: []validator.String{stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("analysis"))},
					}},
					Blocks: map[string]schema.Block{"analysis": schema.ListNestedBlock{
						CustomType:   fwtypes.NewListNestedObjectTypeOf[analysisModel](ctx),
						Validators:   conditionUnionValidators("attribute"),
						NestedObject: schema.NestedBlockObject{Attributes: analysisAttributes()},
					}},
				},
			}},
		},
	}
}

func tlsExpressionBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[tlsExpressionModel](ctx),
		Validators: conditionUnionValidators("boolean_expression", "ip_expression", "ipv6_expression", "string_expression"),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"operator": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.IngressTlsProtocolOperator](), Required: true},
				"value":    schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.IngressTlsProtocolAttribute](), Required: true},
			},
			Blocks: map[string]schema.Block{"evaluate": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[tlsEvaluateModel](ctx),
				Validators: []validator.List{listvalidator.SizeBetween(1, 1)},
				NestedObject: schema.NestedBlockObject{Attributes: map[string]schema.Attribute{
					"attribute": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.IngressTlsAttribute](), Required: true},
				}},
			}},
		},
	}
}

func (r *trafficPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().MailManagerClient(ctx)
	var plan trafficPolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input mailmanager.CreateTrafficPolicyInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("TrafficPolicy")))
	if resp.Diagnostics.HasError() {
		return
	}
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateTrafficPolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	id := aws.ToString(out.TrafficPolicyId)
	created, err := findTrafficPolicyByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, created, &plan, flex.WithFieldNamePrefix("TrafficPolicy")))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *trafficPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().MailManagerClient(ctx)
	var state trafficPolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	out, err := findTrafficPolicyByID(ctx, conn, id)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("TrafficPolicy")))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *trafficPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().MailManagerClient(ctx)
	var plan, state trafficPolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}
	if diff.HasChanges() {
		var input mailmanager.UpdateTrafficPolicyInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("TrafficPolicy")))
		if resp.Diagnostics.HasError() {
			return
		}
		_, err := conn.UpdateTrafficPolicy(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
			return
		}
	}

	out, err := findTrafficPolicyByID(ctx, conn, plan.ID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan, flex.WithFieldNamePrefix("TrafficPolicy")))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *trafficPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().MailManagerClient(ctx)
	var state trafficPolicyResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	var input mailmanager.DeleteTrafficPolicyInput
	input.TrafficPolicyId = aws.String(id)
	_, err := conn.DeleteTrafficPolicy(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
	}
}

func findTrafficPolicyByID(ctx context.Context, conn *mailmanager.Client, id string) (*mailmanager.GetTrafficPolicyOutput, error) {
	input := mailmanager.GetTrafficPolicyInput{TrafficPolicyId: aws.String(id)}
	out, err := conn.GetTrafficPolicy(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	return out, nil
}

var (
	_ flex.Expander  = policyConditionModel{}
	_ flex.Flattener = &policyConditionModel{}
	_ flex.Expander  = booleanEvaluateModel{}
	_ flex.Flattener = &booleanEvaluateModel{}
	_ flex.Expander  = stringEvaluateModel{}
	_ flex.Flattener = &stringEvaluateModel{}
	_ flex.Expander  = ipEvaluateModel{}
	_ flex.Flattener = &ipEvaluateModel{}
	_ flex.Expander  = ipv6EvaluateModel{}
	_ flex.Flattener = &ipv6EvaluateModel{}
	_ flex.Expander  = tlsEvaluateModel{}
	_ flex.Flattener = &tlsEvaluateModel{}
)

func (m policyConditionModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.BooleanExpression.IsNull():
		var r awstypes.PolicyConditionMemberBooleanExpression
		diags.Append(flex.Expand(ctx, m.BooleanExpression, &r.Value)...)
		return &r, diags
	case !m.IPExpression.IsNull():
		var r awstypes.PolicyConditionMemberIpExpression
		diags.Append(flex.Expand(ctx, m.IPExpression, &r.Value)...)
		return &r, diags
	case !m.IPv6Expression.IsNull():
		var r awstypes.PolicyConditionMemberIpv6Expression
		diags.Append(flex.Expand(ctx, m.IPv6Expression, &r.Value)...)
		return &r, diags
	case !m.StringExpression.IsNull():
		var r awstypes.PolicyConditionMemberStringExpression
		diags.Append(flex.Expand(ctx, m.StringExpression, &r.Value)...)
		return &r, diags
	case !m.TLSExpression.IsNull():
		var r awstypes.PolicyConditionMemberTlsExpression
		diags.Append(flex.Expand(ctx, m.TLSExpression, &r.Value)...)
		return &r, diags
	}
	return nil, diags
}

func (m *policyConditionModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.PolicyConditionMemberBooleanExpression:
		var model booleanExpressionModel
		diags.Append(flex.Flatten(ctx, t.Value, &model)...)
		m.BooleanExpression = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.PolicyConditionMemberIpExpression:
		var model ipExpressionModel
		diags.Append(flex.Flatten(ctx, t.Value, &model)...)
		m.IPExpression = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.PolicyConditionMemberIpv6Expression:
		var model ipv6ExpressionModel
		diags.Append(flex.Flatten(ctx, t.Value, &model)...)
		m.IPv6Expression = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.PolicyConditionMemberStringExpression:
		var model stringExpressionModel
		diags.Append(flex.Flatten(ctx, t.Value, &model)...)
		m.StringExpression = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.PolicyConditionMemberTlsExpression:
		var model tlsExpressionModel
		diags.Append(flex.Flatten(ctx, t.Value, &model)...)
		m.TLSExpression = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("policy condition flatten: %T", v))
	}
	return diags
}

func (m booleanEvaluateModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.Analysis.IsNull():
		var r awstypes.IngressBooleanToEvaluateMemberAnalysis
		diags.Append(flex.Expand(ctx, m.Analysis, &r.Value)...)
		return &r, diags
	case !m.IsInAddressList.IsNull():
		var r awstypes.IngressBooleanToEvaluateMemberIsInAddressList
		diags.Append(flex.Expand(ctx, m.IsInAddressList, &r.Value)...)
		return &r, diags
	}
	return nil, diags
}

func (m *booleanEvaluateModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.IngressBooleanToEvaluateMemberAnalysis:
		var model analysisModel
		diags.Append(flex.Flatten(ctx, t.Value, &model)...)
		m.Analysis = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.IngressBooleanToEvaluateMemberIsInAddressList:
		var model isInAddressListModel
		diags.Append(flex.Flatten(ctx, t.Value, &model)...)
		m.IsInAddressList = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("boolean evaluate flatten: %T", v))
	}
	return diags
}

func (m stringEvaluateModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.Analysis.IsNull():
		var r awstypes.IngressStringToEvaluateMemberAnalysis
		diags.Append(flex.Expand(ctx, m.Analysis, &r.Value)...)
		return &r, diags
	case !m.Attribute.IsNull():
		return &awstypes.IngressStringToEvaluateMemberAttribute{Value: m.Attribute.ValueEnum()}, diags
	}
	return nil, diags
}

func (m *stringEvaluateModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.IngressStringToEvaluateMemberAnalysis:
		var model analysisModel
		diags.Append(flex.Flatten(ctx, t.Value, &model)...)
		m.Analysis = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)
	case awstypes.IngressStringToEvaluateMemberAttribute:
		m.Attribute = fwtypes.StringEnumValue(t.Value)
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("string evaluate flatten: %T", v))
	}
	return diags
}

func (m ipEvaluateModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	return &awstypes.IngressIpToEvaluateMemberAttribute{Value: m.Attribute.ValueEnum()}, diags
}

func (m *ipEvaluateModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.IngressIpToEvaluateMemberAttribute:
		m.Attribute = fwtypes.StringEnumValue(t.Value)
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("IP evaluate flatten: %T", v))
	}
	return diags
}

func (m ipv6EvaluateModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	return &awstypes.IngressIpv6ToEvaluateMemberAttribute{Value: m.Attribute.ValueEnum()}, diags
}

func (m *ipv6EvaluateModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.IngressIpv6ToEvaluateMemberAttribute:
		m.Attribute = fwtypes.StringEnumValue(t.Value)
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("IPv6 evaluate flatten: %T", v))
	}
	return diags
}

func (m tlsEvaluateModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	return &awstypes.IngressTlsProtocolToEvaluateMemberAttribute{Value: m.Attribute.ValueEnum()}, diags
}

func (m *tlsEvaluateModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.IngressTlsProtocolToEvaluateMemberAttribute:
		m.Attribute = fwtypes.StringEnumValue(t.Value)
	default:
		diags.AddError("Unexpected Type", fmt.Sprintf("TLS evaluate flatten: %T", v))
	}
	return diags
}

type trafficPolicyResourceModel struct {
	framework.WithRegionModel
	ARN                  types.String                                          `tfsdk:"arn"`
	CreatedTimestamp     timetypes.RFC3339                                     `tfsdk:"created_timestamp"`
	DefaultAction        fwtypes.StringEnum[awstypes.AcceptAction]             `tfsdk:"default_action"`
	ID                   types.String                                          `tfsdk:"id"`
	LastUpdatedTimestamp timetypes.RFC3339                                     `tfsdk:"last_updated_timestamp"`
	MaxMessageSizeBytes  types.Int32                                           `tfsdk:"max_message_size_bytes"`
	Name                 types.String                                          `tfsdk:"name"`
	PolicyStatements     fwtypes.ListNestedObjectValueOf[policyStatementModel] `tfsdk:"policy_statement"`
	Tags                 tftags.Map                                            `tfsdk:"tags"`
	TagsAll              tftags.Map                                            `tfsdk:"tags_all"`
}

type policyStatementModel struct {
	Action     fwtypes.StringEnum[awstypes.AcceptAction]             `tfsdk:"action"`
	Conditions fwtypes.ListNestedObjectValueOf[policyConditionModel] `tfsdk:"condition"`
}
type policyConditionModel struct {
	BooleanExpression fwtypes.ListNestedObjectValueOf[booleanExpressionModel] `tfsdk:"boolean_expression"`
	IPExpression      fwtypes.ListNestedObjectValueOf[ipExpressionModel]      `tfsdk:"ip_expression"`
	IPv6Expression    fwtypes.ListNestedObjectValueOf[ipv6ExpressionModel]    `tfsdk:"ipv6_expression"`
	StringExpression  fwtypes.ListNestedObjectValueOf[stringExpressionModel]  `tfsdk:"string_expression"`
	TLSExpression     fwtypes.ListNestedObjectValueOf[tlsExpressionModel]     `tfsdk:"tls_expression"`
}
type booleanExpressionModel struct {
	Evaluate fwtypes.ListNestedObjectValueOf[booleanEvaluateModel] `tfsdk:"evaluate"`
	Operator fwtypes.StringEnum[awstypes.IngressBooleanOperator]   `tfsdk:"operator"`
}
type booleanEvaluateModel struct {
	Analysis        fwtypes.ListNestedObjectValueOf[analysisModel]        `tfsdk:"analysis"`
	IsInAddressList fwtypes.ListNestedObjectValueOf[isInAddressListModel] `tfsdk:"is_in_address_list"`
}
type analysisModel struct {
	Analyzer    fwtypes.ARN  `tfsdk:"analyzer"`
	ResultField types.String `tfsdk:"result_field"`
}
type isInAddressListModel struct {
	AddressLists fwtypes.ListValueOf[types.String]                             `tfsdk:"address_lists"`
	Attribute    fwtypes.StringEnum[awstypes.IngressAddressListEmailAttribute] `tfsdk:"attribute"`
}
type ipExpressionModel struct {
	Evaluate fwtypes.ListNestedObjectValueOf[ipEvaluateModel] `tfsdk:"evaluate"`
	Operator fwtypes.StringEnum[awstypes.IngressIpOperator]   `tfsdk:"operator"`
	Values   fwtypes.ListValueOf[types.String]                `tfsdk:"values"`
}
type ipEvaluateModel struct {
	Attribute fwtypes.StringEnum[awstypes.IngressIpv4Attribute] `tfsdk:"attribute"`
}
type ipv6ExpressionModel struct {
	Evaluate fwtypes.ListNestedObjectValueOf[ipv6EvaluateModel] `tfsdk:"evaluate"`
	Operator fwtypes.StringEnum[awstypes.IngressIpOperator]     `tfsdk:"operator"`
	Values   fwtypes.ListValueOf[types.String]                  `tfsdk:"values"`
}
type ipv6EvaluateModel struct {
	Attribute fwtypes.StringEnum[awstypes.IngressIpv6Attribute] `tfsdk:"attribute"`
}
type stringExpressionModel struct {
	Evaluate fwtypes.ListNestedObjectValueOf[stringEvaluateModel] `tfsdk:"evaluate"`
	Operator fwtypes.StringEnum[awstypes.IngressStringOperator]   `tfsdk:"operator"`
	Values   fwtypes.ListValueOf[types.String]                    `tfsdk:"values"`
}
type stringEvaluateModel struct {
	Analysis  fwtypes.ListNestedObjectValueOf[analysisModel]           `tfsdk:"analysis"`
	Attribute fwtypes.StringEnum[awstypes.IngressStringEmailAttribute] `tfsdk:"attribute"`
}
type tlsExpressionModel struct {
	Evaluate fwtypes.ListNestedObjectValueOf[tlsEvaluateModel]        `tfsdk:"evaluate"`
	Operator fwtypes.StringEnum[awstypes.IngressTlsProtocolOperator]  `tfsdk:"operator"`
	Value    fwtypes.StringEnum[awstypes.IngressTlsProtocolAttribute] `tfsdk:"value"`
}
type tlsEvaluateModel struct {
	Attribute fwtypes.StringEnum[awstypes.IngressTlsAttribute] `tfsdk:"attribute"`
}
