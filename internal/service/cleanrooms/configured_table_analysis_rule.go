package cleanrooms

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cleanrooms_configured_table_analysis_rule")
// @Tags(identifierAttribute="arn")
func ResourceConfiguredTableAnalysisRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfiguredTableAnalysisRuleCreate,
		ReadWithoutTimeout:   resourceConfiguredTableAnalysisRuleRead,
		UpdateWithoutTimeout: resourceConfiguredTableAnalysisRuleUpdate,
		DeleteWithoutTimeout: resourceConfiguredTableAnalysisRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"analysis_rule_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"configured_table_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"analysis_rule_policy": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: false,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"v1": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: false,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"list": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 1,
										ExactlyOneOf: []string{
											"analysis_rule_policy.0.v1.0.list",
											"analysis_rule_policy.0.v1.0.aggregation",
											"analysis_rule_policy.0.v1.0.custom",
										},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"join_columns": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
												"allowed_join_operators": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
												"list_columns": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
											},
										},
									},
									"aggregation": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 1,
										ExactlyOneOf: []string{
											"analysis_rule_policy.0.v1.0.list",
											"analysis_rule_policy.0.v1.0.aggregation",
											"analysis_rule_policy.0.v1.0.custom",
										},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"aggregate_columns": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"column_names": {
																Type:     schema.TypeList,
																Optional: true,
																ForceNew: false,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},
															"function": {
																Type:     schema.TypeString,
																Optional: true,
																ForceNew: false,
															},
														},
													},
												},
												"join_aggregate_columns": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
												"join_required": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: false,
												},
												"allowed_join_operators": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
												"dimension_columns": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
												"scalar_functions": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
												"output_constraints": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"column_name": {
																Type:     schema.TypeString,
																Optional: true,
																ForceNew: false,
															},
															"minimum": {
																Type:     schema.TypeInt,
																Optional: true,
																ForceNew: false,
															},
															"type": {
																Type:     schema.TypeString,
																Optional: true,
																ForceNew: false,
															},
														},
													},
												},
											},
										},
									},
									"custom": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 1,
										ExactlyOneOf: []string{
											"analysis_rule_policy.0.v1.0.list",
											"analysis_rule_policy.0.v1.0.aggregation",
											"analysis_rule_policy.0.v1.0.custom",
										},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"allowed_custom_analyses": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
												"allowed_analyses_providers": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
												"differential_privacy": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: false,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"columns": {
																Type:     schema.TypeList,
																Optional: true,
																ForceNew: false,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"name": {
																			Type:     schema.TypeString,
																			Optional: true,
																			ForceNew: false,
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

const (
	ResNameConfiguredTableAnalysisRule = "Configured Table Analysis Rule"
)

func resourceConfiguredTableAnalysisRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CleanRoomsClient(ctx)

	input := &cleanrooms.CreateConfiguredTableAnalysisRuleInput{
		ConfiguredTableIdentifier: aws.String(d.Get("configured_table_identifier").(string)),
	}

	if v, ok := d.GetOk("analysis_rule_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AnalysisRulePolicy = expandAnalysisRulePolicy(v.([]interface{}))
	}

	analysisRuleType, err := expandAnalysisRuleType(d.Get("analysis_rule_type").(string))
	if err != nil {
		return create.DiagError(names.CleanRooms, create.ErrActionCreating, ResNameConfiguredTableAnalysisRule, d.Get("name").(string), err)
	}

	if analysisRuleType == "LIST" {
		if listPolicy := d.Get("analysis_rule_policy.0.v1.0.list").([]interface{}); len(listPolicy) == 0 {
			return diag.Errorf("analysis rule type is LIST but got empty analysis rule policy for list.")
		}
	}
	if analysisRuleType == "AGGREGATION" {
		if listPolicy := d.Get("analysis_rule_policy.0.v1.0.aggregation").([]interface{}); len(listPolicy) == 0 {
			return diag.Errorf("analysis rule type is AGGREGATION but got empty analysis rule policy for aggregation.")
		}
	}
	if analysisRuleType == "CUSTOM" {
		if listPolicy := d.Get("analysis_rule_policy.0.v1.0.custom").([]interface{}); len(listPolicy) == 0 {
			return diag.Errorf("analysis rule type is CUSTOM but got empty analysis rule policy for custom.")
		}
	}

	input.AnalysisRuleType = analysisRuleType

	out, err := conn.CreateConfiguredTableAnalysisRule(ctx, input)

	if err != nil {
		return create.DiagError(names.CleanRooms, create.ErrActionCreating, ResNameConfiguredTableAnalysisRule, d.Get("name").(string), err)
	}

	if out == nil || out.AnalysisRule == nil {
		return create.DiagError(names.CleanRooms, create.ErrActionCreating, ResNameConfiguredTableAnalysisRule, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.AnalysisRule.ConfiguredTableId))

	return nil
}

func resourceConfiguredTableAnalysisRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CleanRoomsClient(ctx)

	_, tableErr := findConfiguredTableByID(ctx, conn, d.Get("configured_table_identifier").(string))
	ruleOut, ruleErr := findAnalysisRuleByID(ctx, conn, d.Get("configured_table_identifier").(string), d.Get("analysis_rule_type").(string))

	if !d.IsNewResource() && tfresource.NotFound(ruleErr) {
		log.Printf("[WARN] CleanRooms Configured Table Analysis Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if tableErr == nil && ruleErr != nil {
		log.Printf("[WARN] CleanRooms Configured Table found but associated Analysis Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if ruleErr != nil {
		return create.DiagError(names.CleanRooms, create.ErrActionReading, ResNameConfiguredTableAnalysisRule, d.Id(), ruleErr)
	}

	analysisRule := ruleOut.AnalysisRule
	d.Set(names.AttrARN, analysisRule.ConfiguredTableArn)
	d.Set(names.AttrID, analysisRule.ConfiguredTableId)
	d.Set("create_time", analysisRule.CreateTime.Format(time.RFC3339))
	d.Set("analysis_rule_type", analysisRule.Type)

	return diags
}

func resourceConfiguredTableAnalysisRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CleanRoomsClient(ctx)

	input := &cleanrooms.UpdateConfiguredTableAnalysisRuleInput{
		ConfiguredTableIdentifier: aws.String(d.Id()),
	}

	analysisRuleType, analysisErr := expandAnalysisRuleType(d.Get("analysis_rule_type").(string))
	if analysisErr != nil {
		return create.DiagError(names.CleanRooms, create.ErrActionCreating, ResNameConfiguredTableAnalysisRule, d.Get("name").(string), analysisErr)
	}
	input.AnalysisRuleType = analysisRuleType

	if d.HasChange("analysis_rule_policy") {
		if v, ok := d.GetOk("analysis_rule_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.AnalysisRulePolicy = expandAnalysisRulePolicy(v.([]interface{}))
		}
	}

	_, err := conn.UpdateConfiguredTableAnalysisRule(ctx, input)
	if err != nil {
		return create.DiagError(names.CleanRooms, create.ErrActionUpdating, ResNameConfiguredTableAnalysisRule, d.Id(), err)
	}

	return append(diags, resourceConfiguredTableAnalysisRuleRead(ctx, d, meta)...)
}

func resourceConfiguredTableAnalysisRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).CleanRoomsClient(ctx)

	log.Printf("[INFO] Deleting Configured Table Analysis Rule %s", d.Id())

	ruleType, _ := expandAnalysisRuleType(d.Get("analysis_rule_type").(string))

	_, tableErr := findConfiguredTableByID(ctx, conn, d.Get("configured_table_identifier").(string))
	_, ruleErr := findAnalysisRuleByID(ctx, conn, d.Get("configured_table_identifier").(string), d.Get("analysis_rule_type").(string))

	if tableErr == nil && ruleErr != nil {
		log.Printf("[WARN] CleanRooms Configured Table found but associated Analysis Rule (%s) not found, removing from state", d.Id())
		return diags
	}
	_, err := conn.DeleteConfiguredTableAnalysisRule(ctx, &cleanrooms.DeleteConfiguredTableAnalysisRuleInput{
		ConfiguredTableIdentifier: aws.String(d.Get("configured_table_identifier").(string)),
		AnalysisRuleType:          ruleType,
	})

	if errs.IsA[*types.AccessDeniedException](err) {
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.CleanRooms, create.ErrActionDeleting, ResNameConfiguredTableAnalysisRule, d.Id(), err)
	}

	return diags
}

func expandAnalysisRuleType(rule string) (types.ConfiguredTableAnalysisRuleType, error) {
	switch rule {
	case "AGGREGATION":
		return types.ConfiguredTableAnalysisRuleTypeAggregation, nil
	case "LIST":
		return types.ConfiguredTableAnalysisRuleTypeList, nil
	case "CUSTOM":
		return types.ConfiguredTableAnalysisRuleTypeCustom, nil
	default:
		return types.ConfiguredTableAnalysisRuleTypeAggregation, fmt.Errorf("invalid analysis_rule_type: %s", rule)
	}
}

func expandAnalysisRulePolicy(tfList []interface{}) types.ConfiguredTableAnalysisRulePolicy {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	log.Printf("tfMap expandAnalysisRulePolicy: %+v\n", tfMap)

	if v, ok := tfMap["v1"].([]interface{}); ok && len(v) > 0 {
		return &types.ConfiguredTableAnalysisRulePolicyMemberV1{
			Value: expandAnalysisPolicyV1(v),
		}
	} else {
		return nil
	}
}

func expandAnalysisPolicyV1(tfList []interface{}) types.ConfiguredTableAnalysisRulePolicyV1 {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	log.Printf("tfMap expandAnalysisPolicyV1: %+v\n", tfMap)

	if v, ok := tfMap["list"].([]interface{}); ok && len(v) > 0 {
		return &types.ConfiguredTableAnalysisRulePolicyV1MemberList{
			Value: *expandAnalysisPolicyList(v),
		}
	} else if v, ok := tfMap["aggregation"].([]interface{}); ok && len(v) > 0 {
		return &types.ConfiguredTableAnalysisRulePolicyV1MemberAggregation{
			Value: *expandAnalysisPolicyAggregation(v),
		}
	} else if v, ok := tfMap["custom"].([]interface{}); ok && len(v) > 0 {
		return &types.ConfiguredTableAnalysisRulePolicyV1MemberCustom{
			Value: *expandAnalysisPolicyCustom(v),
		}
	} else {
		return nil
	}

}

func expandAnalysisPolicyList(tfList []interface{}) *types.AnalysisRuleList {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	a := &types.AnalysisRuleList{}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	if v, ok := tfMap["join_columns"].([]interface{}); ok && len(v) > 0 {
		a.JoinColumns = convertInterfaceSliceToStringSlice(v)
	}
	if v, ok := tfMap["allowed_join_operators"].([]interface{}); ok && len(v) > 0 {
		a.AllowedJoinOperators = expandAllowedJoinOperators(v)
	}
	if v, ok := tfMap["list_columns"].([]interface{}); ok && len(v) > 0 {
		a.ListColumns = convertInterfaceSliceToStringSlice(v)
	}

	return a
}

func convertInterfaceSliceToStringSlice(interfaces []interface{}) []string {
	strings := make([]string, len(interfaces))
	for i, v := range interfaces {
		if str, ok := v.(string); ok {
			strings[i] = str
		} else {
			continue
		}
	}
	return strings
}

func expandAllowedJoinOperators(data []interface{}) []types.JoinOperator {
	var joinOperator []types.JoinOperator
	for _, v := range data {
		switch v {
		case "OR":
			joinOperator = append(joinOperator, types.JoinOperatorOr)
		case "AND":
			joinOperator = append(joinOperator, types.JoinOperatorAnd)
		}
	}
	return joinOperator
}

func expandAnalysisPolicyAggregation(tfList []interface{}) *types.AnalysisRuleAggregation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	a := &types.AnalysisRuleAggregation{}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	if v, ok := tfMap["aggregate_columns"].([]interface{}); ok && len(v) > 0 {
		a.AggregateColumns = expandAggregateColumns(v)
	}

	if v, ok := tfMap["join_aggregate_columns"].([]interface{}); ok && len(v) > 0 {
		a.JoinColumns = convertInterfaceSliceToStringSlice(v)
	}

	if v, ok := tfMap["join_required"].(string); ok && len(v) > 0 {
		a.JoinRequired = expandJoinedRequiredOption(v)
	}

	if v, ok := tfMap["dimension_columns"].([]interface{}); ok && len(v) > 0 {
		a.DimensionColumns = convertInterfaceSliceToStringSlice(v)
	}

	if v, ok := tfMap["scalar_functions"].([]interface{}); ok && len(v) > 0 {
		a.ScalarFunctions = expandScalarFunctions(convertInterfaceSliceToStringSlice(v))
	}

	if v, ok := tfMap["allowed_join_operators"].([]interface{}); ok && len(v) > 0 {
		a.AllowedJoinOperators = expandAllowedJoinOperators(v)
	}

	if v, ok := tfMap["output_constraints"].([]interface{}); ok && len(v) > 0 {
		a.OutputConstraints = expandOutputConstraints(v)
	}

	return a

}

func expandAggregateColumns(tfList []interface{}) []types.AggregateColumn {
	if len(tfList) == 0 {
		return nil
	}

	a := make([]types.AggregateColumn, 0, len(tfList))

	for _, tfListRaw := range tfList {

		tfMap, ok := tfListRaw.(map[string]interface{})
		if !ok {
			continue
		}

		a = append(a, types.AggregateColumn{
			ColumnNames: convertInterfaceSliceToStringSlice(tfMap["column_names"].([]interface{})),
			Function:    expandFunction(tfMap["function"].(string)),
		})

	}

	return a

}

func expandFunction(value string) types.AggregateFunctionName {
	switch value {
	case "SUM":
		return types.AggregateFunctionNameSum
	case "SUM_DISTINCT":
		return types.AggregateFunctionNameSumDistinct
	case "COUNT":
		return types.AggregateFunctionNameCount
	case "COUNT_DISTINCT":
		return types.AggregateFunctionNameCountDistinct
	case "AVG":
		return types.AggregateFunctionNameAvg
	default:
		return types.AggregateFunctionNameSum
	}
}

func expandOutputConstraints(tfList []interface{}) []types.AggregationConstraint {
	if len(tfList) == 0 {
		return nil
	}

	a := make([]types.AggregationConstraint, 0, len(tfList))

	for _, tfListRaw := range tfList {

		tfMap, ok := tfListRaw.(map[string]interface{})
		if !ok {
			continue
		}

		minimum := int32(tfMap["minimum"].(int))

		a = append(a, types.AggregationConstraint{
			ColumnName: aws.String(tfMap["column_name"].(string)),
			Minimum:    &minimum,
			Type:       expandOutputConstraintsType(tfMap["type"].(string)),
		})

	}

	return a

}

func expandOutputConstraintsType(value string) types.AggregationType {
	switch value {
	case "COUNT_DISTINCT":
		return types.AggregationTypeCountDistinct
	default:
		return types.AggregationTypeCountDistinct
	}
}

func expandJoinedRequiredOption(value string) types.JoinRequiredOption {
	switch value {
	case "QUERY_RUNNER":
		return types.JoinRequiredOptionQueryRunner
	default:
		return types.JoinRequiredOptionQueryRunner
	}
}

func expandScalarFunctions(data []string) []types.ScalarFunctions {
	var scalarFunction []types.ScalarFunctions
	for _, v := range data {
		switch v {
		case "ABS":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsAbs)
		case "CAST":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsCast)
		case "CEILING":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsCeiling)
		case "COALESCE":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsCoalesce)
		case "CONVERT":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsConvert)
		case "CURRENT_DATE":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsCurrentDate)
		case "DATE_ADD":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsDateadd)
		case "EXTRACT":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsExtract)
		case "FLOOR":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsFloor)
		case "GETDATE":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsGetdate)
		case "LN":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsLn)
		case "LOG":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsLog)
		case "LOWER":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsLower)
		case "ROUND":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsRound)
		case "RTRIM":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsRtrim)
		case "SQRT":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsSqrt)
		case "SUBSTRING":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsSubstring)
		case "TO_CHAR":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsToChar)
		case "TO_DATE":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsToDate)
		case "TO_NUMBER":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsToNumber)
		case "TO_TIMESTAMP":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsToTimestamp)
		case "TRIM":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsTrim)
		case "TRUNC":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsTrunc)
		case "UPPER":
			scalarFunction = append(scalarFunction, types.ScalarFunctionsUpper)
		}
	}
	return scalarFunction
}

func expandAnalysisPolicyCustom(tfList []interface{}) *types.AnalysisRuleCustom {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	a := &types.AnalysisRuleCustom{}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	if v, ok := tfMap["allowed_custom_analyses"].([]interface{}); ok && len(v) > 0 {
		a.AllowedAnalyses = convertInterfaceSliceToStringSlice(v)
	}

	if v, ok := tfMap["allowed_analyses_providers"].([]interface{}); ok && len(v) > 0 {
		a.AllowedAnalysisProviders = convertInterfaceSliceToStringSlice(v)
	}

	if v, ok := tfMap["differential_privacy"].([]interface{}); ok && len(v) > 0 {
		a.DifferentialPrivacy = expandExpandDifferentialPrivacy(v)
	}

	return a
}

func expandExpandDifferentialPrivacy(tfList []interface{}) *types.DifferentialPrivacyConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.DifferentialPrivacyConfiguration{}

	if v, ok := tfMap["columns"].([]interface{}); ok && len(v) > 0 {
		config.Columns = expandDifferentialPrivacyColumns(v)
	}

	return config
}

func expandDifferentialPrivacyColumns(tfList []interface{}) []types.DifferentialPrivacyColumn {
	var columns []types.DifferentialPrivacyColumn

	for _, v := range tfList {
		if columnMap, ok := v.(map[string]interface{}); ok {
			column := types.DifferentialPrivacyColumn{}
			if name, ok := columnMap["name"].(string); ok {
				column.Name = &name
			}
			columns = append(columns, column)
		}
	}

	return columns
}
