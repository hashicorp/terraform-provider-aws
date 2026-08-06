// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_iam_custom_policy_simulation", name="Custom Policy Simulation")
func newCustomPolicySimulationDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &customPolicySimulationDataSource{}, nil
}

type customPolicySimulationDataSource struct {
	framework.DataSourceWithModel[customPolicySimulationDataSourceModel]
}

func (d *customPolicySimulationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"action_names": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Required:    true,
			},
			"all_allowed": schema.BoolAttribute{
				Computed: true,
			},
			"caller_arn": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					fwvalidators.ARN(),
				},
			},
			"permissions_boundary_policies_json": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(fwvalidators.JSON()),
				},
			},
			"policies_json": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(fwvalidators.JSON()),
				},
			},
			"resource_arns": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(fwvalidators.ARN()),
				},
			},
			"resource_handling_option": schema.StringAttribute{
				Optional: true,
			},
			"resource_owner_account_id": schema.StringAttribute{
				Optional: true,
			},
			"resource_policy_json": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					fwvalidators.JSON(),
				},
			},
			"results": framework.DataSourceComputedListOfObjectAttribute[evaluationResultModel](ctx),
		},
		Blocks: map[string]schema.Block{
			"context": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[contextEntryModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrKey: schema.StringAttribute{
							Required: true,
						},
						names.AttrType: schema.StringAttribute{
							Required: true,
						},
						names.AttrValues: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (d *customPolicySimulationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().IAMClient(ctx)

	var data customPolicySimulationDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	input := iam.SimulateCustomPolicyInput{
		ActionNames:                        fwflex.ExpandFrameworkStringValueSet(ctx, data.ActionNames),
		PolicyInputList:                    fwflex.ExpandFrameworkStringValueSet(ctx, data.PoliciesJSON),
		PermissionsBoundaryPolicyInputList: fwflex.ExpandFrameworkStringValueSet(ctx, data.PermissionsBoundaryPoliciesJSON),
		ResourceArns:                       fwflex.ExpandFrameworkStringValueSet(ctx, data.ResourceARNs),
	}

	input.CallerArn = data.CallerARN.ValueStringPointer()
	input.ResourceHandlingOption = data.ResourceHandlingOption.ValueStringPointer()
	input.ResourceOwner = data.ResourceOwnerAccountID.ValueStringPointer()
	input.ResourcePolicy = data.ResourcePolicyJSON.ValueStringPointer()

	contextEntries, diags := data.Context.ToSlice(ctx)
	smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
	if resp.Diagnostics.HasError() {
		return
	}
	for _, entry := range contextEntries {
		input.ContextEntries = append(input.ContextEntries, awstypes.ContextEntry{
			ContextKeyName:   entry.Key.ValueStringPointer(),
			ContextKeyType:   awstypes.ContextKeyTypeEnum(entry.Type.ValueString()),
			ContextKeyValues: fwflex.ExpandFrameworkStringValueSet(ctx, entry.Values),
		})
	}

	apiResults, err := findCustomPolicySimulationResults(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	allowedCount := 0
	deniedCount := 0

	results := make([]evaluationResultModel, len(apiResults))
	for i, result := range apiResults {
		allowed := string(result.EvalDecision) == "allowed"
		if allowed {
			allowedCount++
		} else {
			deniedCount++
		}

		var decisionDetails map[string]string
		for k, v := range result.EvalDecisionDetails {
			if v != "" {
				if decisionDetails == nil {
					decisionDetails = make(map[string]string)
				}
				decisionDetails[k] = string(v)
			}
		}

		var matchedStmts []matchedStatementModel
		for _, stmt := range result.MatchedStatements {
			matchedStmts = append(matchedStmts, matchedStatementModel{
				SourcePolicyID:   fwflex.StringToFramework(ctx, stmt.SourcePolicyId),
				SourcePolicyType: fwflex.StringValueToFramework(ctx, stmt.SourcePolicyType),
			})
		}

		var missingContextKeys []string
		for _, v := range result.MissingContextValues {
			if v != "" {
				missingContextKeys = append(missingContextKeys, v)
			}
		}

		results[i] = evaluationResultModel{
			ActionName:         fwflex.StringToFramework(ctx, result.EvalActionName),
			Allowed:            types.BoolValue(allowed),
			Decision:           types.StringValue(string(result.EvalDecision)),
			DecisionDetails:    fwtypes.MapOfString{MapValue: fwflex.FlattenFrameworkStringValueMapLegacy(ctx, decisionDetails)},
			ResourceARN:        fwflex.StringToFramework(ctx, result.EvalResourceName),
			MatchedStatements:  fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, matchedStmts),
			MissingContextKeys: fwflex.FlattenFrameworkStringValueSetOfStringLegacy(ctx, missingContextKeys),
		}
	}

	data.AllAllowed = types.BoolValue(allowedCount > 0 && deniedCount == 0)
	data.Results = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, results)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func findCustomPolicySimulationResults(ctx context.Context, conn *iam.Client, input *iam.SimulateCustomPolicyInput) ([]awstypes.EvaluationResult, error) {
	var output []awstypes.EvaluationResult

	paginator := iam.NewSimulateCustomPolicyPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		output = append(output, page.EvaluationResults...)
	}

	return output, nil
}

type customPolicySimulationDataSourceModel struct {
	ActionNames                     fwtypes.SetOfString                                    `tfsdk:"action_names"`
	AllAllowed                      types.Bool                                             `tfsdk:"all_allowed"`
	CallerARN                       types.String                                           `tfsdk:"caller_arn"`
	Context                         fwtypes.SetNestedObjectValueOf[contextEntryModel]      `tfsdk:"context"`
	PermissionsBoundaryPoliciesJSON fwtypes.SetOfString                                    `tfsdk:"permissions_boundary_policies_json"`
	PoliciesJSON                    fwtypes.SetOfString                                    `tfsdk:"policies_json"`
	ResourceARNs                    fwtypes.SetOfString                                    `tfsdk:"resource_arns"`
	ResourceHandlingOption          types.String                                           `tfsdk:"resource_handling_option"`
	ResourceOwnerAccountID          types.String                                           `tfsdk:"resource_owner_account_id"`
	ResourcePolicyJSON              types.String                                           `tfsdk:"resource_policy_json"`
	Results                         fwtypes.ListNestedObjectValueOf[evaluationResultModel] `tfsdk:"results"`
}

type contextEntryModel struct {
	Key    types.String        `tfsdk:"key"`
	Type   types.String        `tfsdk:"type"`
	Values fwtypes.SetOfString `tfsdk:"values"`
}

type evaluationResultModel struct {
	ActionName         types.String                                           `tfsdk:"action_name"`
	Allowed            types.Bool                                             `tfsdk:"allowed"`
	Decision           types.String                                           `tfsdk:"decision"`
	DecisionDetails    fwtypes.MapOfString                                    `tfsdk:"decision_details"`
	ResourceARN        types.String                                           `tfsdk:"resource_arn"`
	MatchedStatements  fwtypes.ListNestedObjectValueOf[matchedStatementModel] `tfsdk:"matched_statements"`
	MissingContextKeys fwtypes.SetOfString                                    `tfsdk:"missing_context_keys"`
}

type matchedStatementModel struct {
	SourcePolicyID   types.String `tfsdk:"source_policy_id"`
	SourcePolicyType types.String `tfsdk:"source_policy_type"`
}
