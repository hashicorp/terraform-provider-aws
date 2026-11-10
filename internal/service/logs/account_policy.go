// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"strings"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cloudwatch_log_account_policy", name="Account Policy")
// @IdentityAttribute("policy_name")
// @IdentityAttribute("policy_type")
// @ImportIDHandler("accountPolicyImportID")
// @Testing(preIdentityVersion="v6.51.0")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types;awstypes;awstypes.AccountPolicy")
// @Testing(importStateIdFunc=testAccAccountPolicyImportStateIDFunc)
// @Testing(importStateIdAttribute="policy_name")
func newAccountPolicyResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &accountPolicyResource{}, nil
}

type accountPolicyResource struct {
	framework.ResourceWithModel[accountPolicyResourceModel]
	framework.WithImportByIdentity
}

func (r *accountPolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"policy_document": schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Required:   true,
			},
			"policy_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PolicyType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrScope: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.Scope](),
				Optional:   true,
				Computed:   true,
				Default:    stringdefault.StaticString(string(awstypes.ScopeAll)),
			},
			"selection_criteria": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *accountPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan accountPolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	var input cloudwatchlogs.PutAccountPolicyInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.PutAccountPolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, "policy_name", plan.PolicyName.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output.AccountPolicy, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *accountPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accountPolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.PolicyName)
	output, err := findAccountPolicyByTwoPartKey(ctx, conn, name, data.PolicyType.ValueEnum())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, "policy_name", name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *accountPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan accountPolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	var input cloudwatchlogs.PutAccountPolicyInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.PutAccountPolicy(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, "policy_name", plan.PolicyName.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output.AccountPolicy, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *accountPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data accountPolicyResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().LogsClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.PolicyName)
	input := cloudwatchlogs.DeleteAccountPolicyInput{
		PolicyName: aws.String(name),
		PolicyType: data.PolicyType.ValueEnum(),
	}
	_, err := conn.DeleteAccountPolicy(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, "policy_name", name)
	}
}

// flatten sets the model from the API value, for Read and for an item the list resource includes.
func (r *accountPolicyResource) flatten(ctx context.Context, policy *awstypes.AccountPolicy, data *accountPolicyResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.Append(fwflex.Flatten(ctx, policy, data)...)
	if diags.HasError() {
		return diags
	}

	// AWS doesn't preserve key order. Sort keys, as the SDKv2 implementation did, so imported
	// and pre-migration state stay byte-for-byte comparable.
	var document any
	if err := json.Unmarshal([]byte(aws.ToString(policy.PolicyDocument)), &document); err != nil {
		diags.AddError("normalizing CloudWatch Logs Account Policy document", err.Error())
		return diags
	}
	b, err := json.Marshal(document)
	if err != nil {
		diags.AddError("normalizing CloudWatch Logs Account Policy document", err.Error())
		return diags
	}

	data.ID = fwflex.StringToFramework(ctx, policy.PolicyName)
	data.PolicyDocument = jsontypes.NewNormalizedValue(string(b))

	return diags
}

// UpgradeState converts SDKv2 state. SDKv2 stored an unset selection_criteria as "",
// which the Framework would plan as a change to null and replace the policy.
func (r *accountPolicyResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID:         schema.StringAttribute{Computed: true},
			"policy_document":    schema.StringAttribute{Required: true},
			"policy_name":        schema.StringAttribute{Required: true},
			"policy_type":        schema.StringAttribute{Required: true},
			names.AttrRegion:     schema.StringAttribute{Optional: true, Computed: true},
			names.AttrScope:      schema.StringAttribute{Optional: true, Computed: true},
			"selection_criteria": schema.StringAttribute{Optional: true},
		},
	}

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &schemaV0,
			StateUpgrader: upgradeAccountPolicyStateV0toV1,
		},
	}
}

func upgradeAccountPolicyStateV0toV1(ctx context.Context, request resource.UpgradeStateRequest, response *resource.UpgradeStateResponse) {
	type accountPolicyResourceModelV0 struct {
		ID                types.String `tfsdk:"id"`
		PolicyDocument    types.String `tfsdk:"policy_document"`
		PolicyName        types.String `tfsdk:"policy_name"`
		PolicyType        types.String `tfsdk:"policy_type"`
		Region            types.String `tfsdk:"region"`
		Scope             types.String `tfsdk:"scope"`
		SelectionCriteria types.String `tfsdk:"selection_criteria"`
	}

	var v0 accountPolicyResourceModelV0
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &v0))
	if response.Diagnostics.HasError() {
		return
	}

	v1 := accountPolicyResourceModel{
		ID:                v0.ID,
		PolicyDocument:    jsontypes.NewNormalizedValue(v0.PolicyDocument.ValueString()),
		PolicyName:        v0.PolicyName,
		PolicyType:        fwtypes.StringEnumValue(awstypes.PolicyType(v0.PolicyType.ValueString())),
		Scope:             fwtypes.StringEnumValue(awstypes.Scope(v0.Scope.ValueString())),
		SelectionCriteria: v0.SelectionCriteria,

		Region: v0.Region}
	if v0.SelectionCriteria.ValueString() == "" {
		v1.SelectionCriteria = types.StringNull()
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, v1))
}

type accountPolicyResourceModel struct {
	framework.WithRegionModel
	ID                types.String                            `tfsdk:"id"`
	PolicyDocument    jsontypes.Normalized                    `tfsdk:"policy_document"`
	PolicyName        types.String                            `tfsdk:"policy_name"`
	PolicyType        fwtypes.StringEnum[awstypes.PolicyType] `tfsdk:"policy_type"`
	Scope             fwtypes.StringEnum[awstypes.Scope]      `tfsdk:"scope"`
	SelectionCriteria types.String                            `tfsdk:"selection_criteria" autoflex:",omitempty"`
}

func findAccountPolicyByTwoPartKey(ctx context.Context, conn *cloudwatchlogs.Client, policyName string, policyType awstypes.PolicyType) (*awstypes.AccountPolicy, error) {
	input := cloudwatchlogs.DescribeAccountPoliciesInput{
		PolicyName: aws.String(policyName),
		PolicyType: policyType,
	}
	output, err := findAccountPolicy(ctx, conn, &input)
	if err != nil {
		return nil, err
	}

	if output.PolicyDocument == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}

func findAccountPolicy(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeAccountPoliciesInput) (*awstypes.AccountPolicy, error) {
	output, err := findAccountPolicies(ctx, conn, input)
	if err != nil {
		return nil, err
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output))
}

func findAccountPolicies(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeAccountPoliciesInput) ([]awstypes.AccountPolicy, error) {
	var output []awstypes.AccountPolicy
	for item, err := range listAccountPolicies(ctx, conn, input) {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		output = append(output, item)
	}

	return output, nil
}

func listAccountPolicies(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeAccountPoliciesInput) iter.Seq2[awstypes.AccountPolicy, error] {
	return tfiter.ConcatValuesWithError(listAccountPolicyPages(ctx, conn, input))
}

// listAccountPolicyPages pages DescribeAccountPolicies, which has no SDK paginator.
func listAccountPolicyPages(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeAccountPoliciesInput) iter.Seq2[[]awstypes.AccountPolicy, error] {
	return func(yield func([]awstypes.AccountPolicy, error) bool) {
		for {
			page, err := conn.DescribeAccountPolicies(ctx, input)
			if err != nil {
				yield(nil, fmt.Errorf("listing CloudWatch Logs Account Policies: %w", err))
				return
			}

			if !yield(page.AccountPolicies, nil) {
				return
			}

			if aws.ToString(page.NextToken) == "" {
				return
			}
			input.NextToken = page.NextToken
		}
	}
}

const accountPolicyImportIDSeparator = ":"

var _ inttypes.ImportIDParser = accountPolicyImportID{}

type accountPolicyImportID struct{}

func (accountPolicyImportID) Parse(id string) (string, map[string]any, error) {
	policyName, policyType, found := strings.Cut(id, accountPolicyImportIDSeparator)
	if !found || policyName == "" || policyType == "" || strings.Contains(policyType, accountPolicyImportIDSeparator) {
		return "", nil, fmt.Errorf("unexpected format for ID (%[1]s), expected <policy-name>%[2]s<policy-type>", id, accountPolicyImportIDSeparator)
	}

	result := map[string]any{
		"policy_name": policyName,
		"policy_type": policyType,
	}

	return policyName, result, nil
}
