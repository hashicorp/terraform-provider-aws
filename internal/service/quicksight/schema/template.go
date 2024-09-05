// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TemplateDefinitionSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TemplateVersionDefinition.html
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Computed: true,
		ExactlyOneOf: []string{
			"definition",
			"source_entity",
		},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"data_set_configuration": dataSetConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataSetConfiguration.html
				"analysis_defaults":      analysisDefaultSchema(),      // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AnalysisDefaults.html
				"calculated_fields":      calculatedFieldsSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CalculatedField.html
				"column_configurations": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnConfiguration.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 200,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"column":               columnSchema(true),          // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
							"format_configuration": formatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FormatConfiguration.html
							names.AttrRole:         stringSchema(false, enum.Validate[awstypes.ColumnRole]()),
						},
					},
				},
				"filter_groups": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterGroup.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 2000,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"cross_dataset":       stringSchema(true, enum.Validate[awstypes.CrossDatasetTypes]()),
							"filter_group_id":     idSchema(),
							"filters":             filtersSchema(),                  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Filter.html
							"scope_configuration": filterScopeConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterScopeConfiguration.html
							names.AttrStatus:      stringSchema(false, enum.Validate[awstypes.Status]()),
						},
					},
				},
				"parameters_declarations": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterDeclaration.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 200,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"date_time_parameter_declaration": dateTimeParameterDeclarationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimeParameterDeclaration.html
							"decimal_parameter_declaration":   decimalParameterDeclarationSchema(),  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DecimalParameterDeclaration.html
							"integer_parameter_declaration":   integerParameterDeclarationSchema(),  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_IntegerParameterDeclaration.html
							"string_parameter_declaration":    stringParameterDeclarationSchema(),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_StringParameterDeclaration.html
						},
					},
				},
				"sheets": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetDefinition.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 20,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"sheet_id": idSchema(),
							names.AttrContentType: {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[awstypes.SheetContentType](),
							},
							names.AttrDescription:   stringSchema(false, validation.StringLenBetween(1, 1024)),
							"filter_controls":       filterControlsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterControl.html
							"layouts":               layoutSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Layout.html
							names.AttrName:          stringSchema(false, validation.StringLenBetween(1, 2048)),
							"parameter_controls":    parameterControlsSchema(),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterControl.html
							"sheet_control_layouts": sheetControlLayoutsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetControlLayout.html
							"text_boxes": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetTextBox.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 100,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"sheet_text_box_id": idSchema(),
										names.AttrContent:   stringSchema(false, validation.StringLenBetween(1, 150000)),
									},
								},
							},
							"title":   stringSchema(false, validation.StringLenBetween(1, 1024)),
							"visuals": visualsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Visual.html
						},
					},
				},
			},
		},
	}
}

func stringOptionalComputedSchema(validateFunc any) *schema.Schema {
	switch v := validateFunc.(type) {
	case schema.SchemaValidateDiagFunc:
		return &schema.Schema{
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: v,
		}
	case schema.SchemaValidateFunc:
		return stringOptionalComputedSchema(validation.ToDiagFunc(v))
	default:
		panic(fmt.Sprintf("unsupported validateFunc type: %T", v)) //lintignore:R009
	}
}

func stringSchema(required bool, validateFunc any) *schema.Schema {
	switch v := validateFunc.(type) {
	case schema.SchemaValidateDiagFunc:
		return &schema.Schema{
			Type:             schema.TypeString,
			Required:         required,
			Optional:         !required,
			ValidateDiagFunc: v,
		}
	case schema.SchemaValidateFunc:
		return stringSchema(required, validation.ToDiagFunc(v))
	case func(interface{}, string) ([]string, []error):
		return stringSchema(required, schema.SchemaValidateFunc(v))
	default:
		panic(fmt.Sprintf("unsupported validateFunc type: %T", v)) //lintignore:R009
	}
}

func intSchema(required bool, validateFunc any) *schema.Schema {
	switch v := validateFunc.(type) {
	case schema.SchemaValidateDiagFunc:
		return &schema.Schema{
			Type:             schema.TypeInt,
			Required:         required,
			Optional:         !required,
			ValidateDiagFunc: v,
		}
	case schema.SchemaValidateFunc:
		return intSchema(required, validation.ToDiagFunc(v))
	case func(interface{}, string) ([]string, []error):
		return intSchema(required, schema.SchemaValidateFunc(v))
	default:
		panic(fmt.Sprintf("unsupported validateFunc type: %T", v)) //lintignore:R009
	}
}

func floatSchema(required bool, validateFunc any) *schema.Schema {
	switch v := validateFunc.(type) {
	case schema.SchemaValidateDiagFunc:
		return &schema.Schema{
			Type:             schema.TypeFloat,
			Required:         required,
			Optional:         !required,
			ValidateDiagFunc: v,
		}
	case schema.SchemaValidateFunc:
		return floatSchema(required, validation.ToDiagFunc(v))
	case func(interface{}, string) ([]string, []error):
		return floatSchema(required, schema.SchemaValidateFunc(v))
	default:
		panic(fmt.Sprintf("unsupported validateFunc type: %T", v)) //lintignore:R009
	}
}

func aggregationFunctionSchema(required bool) *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AggregationFunction.html
		Type:     schema.TypeList,
		Required: required,
		Optional: !required,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"categorical_aggregation_function": stringSchema(false, enum.Validate[awstypes.CategoricalAggregationFunction]()),
				"date_aggregation_function":        stringSchema(false, enum.Validate[awstypes.DateAggregationFunction]()),
				"numerical_aggregation_function":   numericalAggregationFunctionSchema(false), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericalAggregationFunction.html
			},
		},
	}
}

func calculatedFieldsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CalculatedField.html
		Type:     schema.TypeSet,
		MinItems: 1,
		MaxItems: 500,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"data_set_identifier": stringSchema(true, validation.StringLenBetween(1, 2048)),
				names.AttrExpression:  stringSchema(true, validation.StringLenBetween(1, 32000)),
				names.AttrName:        stringSchema(true, validation.StringLenBetween(1, 128)),
			},
		},
	}
}

func numericalAggregationFunctionSchema(required bool) *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericalAggregationFunction.html
		Type:     schema.TypeList,
		Required: required,
		Optional: !required,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"percentile_aggregation": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PercentileAggregation.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"percentile_value": {
								Type:         schema.TypeFloat,
								Optional:     true,
								ValidateFunc: validation.IntBetween(0, 100),
							},
						},
					},
				},
				"simple_numerical_aggregation": stringSchema(false, enum.Validate[awstypes.SimpleNumericalAggregationFunction]()),
			},
		},
	}
}

func idSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ValidateFunc: validation.All(
			validation.StringLenBetween(1, 512),
			validation.StringMatch(regexache.MustCompile(`[\w\-]+`), "must contain only alphanumeric, hyphen, and underscore characters"),
		),
	}
}

func columnSchema(required bool) *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Required: required,
		Optional: !required,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"column_name":         stringSchema(true, validation.StringLenBetween(1, 128)),
				"data_set_identifier": stringSchema(true, validation.StringLenBetween(1, 2048)),
			},
		},
	}
}

func dataSetConfigurationSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataSetConfiguration.html
		Type:     schema.TypeList,
		MaxItems: 30,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"column_group_schema_list": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnGroupSchema.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 500,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"column_group_column_schema_list": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnGroupColumnSchema.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 500,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrName: {
											Type:     schema.TypeString,
											Optional: true,
										},
									},
								},
							},
							names.AttrName: {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"data_set_schema": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataSetSchema.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"column_schema_list": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnSchema.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 500,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"data_type": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"geographic_role": {
											Type:     schema.TypeString,
											Optional: true,
										},
										names.AttrName: {
											Type:     schema.TypeString,
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
				"placeholder": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func rollingDateConfigurationSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RollingDateConfiguration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"data_set_identifier": stringSchema(false, validation.StringLenBetween(1, 2048)),
				names.AttrExpression:  stringSchema(true, validation.StringLenBetween(1, 4096)),
			},
		},
	}
}

func TemplateSourceEntitySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		ExactlyOneOf: []string{
			"definition",
			"source_entity",
		},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"source_analysis": {
					Type:         schema.TypeList,
					MaxItems:     1,
					Optional:     true,
					ExactlyOneOf: []string{"source_entity.0.source_analysis", "source_entity.0.source_template"},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"data_set_references": dataSetReferencesSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataSetReference.html
						},
					},
				},
				"source_template": {
					Type:         schema.TypeList,
					MaxItems:     1,
					Optional:     true,
					ExactlyOneOf: []string{"source_entity.0.source_analysis", "source_entity.0.source_template"},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
						},
					},
				},
			},
		},
	}
}

func ExpandTemplateSourceEntity(tfList []interface{}) *awstypes.TemplateSourceEntity {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.TemplateSourceEntity{}

	if v, ok := tfMap["source_analysis"].([]interface{}); ok && len(v) > 0 {
		apiObject.SourceAnalysis = expandSourceAnalysis(v[0].(map[string]interface{}))
	} else if v, ok := tfMap["source_template"].([]interface{}); ok && len(v) > 0 {
		apiObject.SourceTemplate = expandTemplateSourceTemplate(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandSourceAnalysis(tfMap map[string]interface{}) *awstypes.TemplateSourceAnalysis {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TemplateSourceAnalysis{}

	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		apiObject.Arn = aws.String(v)
	}
	if v, ok := tfMap["data_set_references"].([]interface{}); ok && len(v) > 0 {
		apiObject.DataSetReferences = expandDataSetReferences(v)
	}

	return apiObject
}

func expandDataSetReferences(tfList []interface{}) []awstypes.DataSetReference {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.DataSetReference

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandDataSetReference(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandDataSetReference(tfMap map[string]interface{}) *awstypes.DataSetReference {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DataSetReference{}

	if v, ok := tfMap["data_set_arn"].(string); ok {
		apiObject.DataSetArn = aws.String(v)
	}
	if v, ok := tfMap["data_set_placeholder"].(string); ok {
		apiObject.DataSetPlaceholder = aws.String(v)
	}

	return apiObject
}

func expandTemplateSourceTemplate(tfMap map[string]interface{}) *awstypes.TemplateSourceTemplate {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TemplateSourceTemplate{}

	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		apiObject.Arn = aws.String(v)
	}

	return apiObject
}

func ExpandTemplateDefinition(tfList []interface{}) *awstypes.TemplateVersionDefinition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.TemplateVersionDefinition{}

	if v, ok := tfMap["analysis_defaults"].([]interface{}); ok && len(v) > 0 {
		apiObject.AnalysisDefaults = expandAnalysisDefaults(v)
	}
	if v, ok := tfMap["calculated_fields"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CalculatedFields = expandCalculatedFields(v.List())
	}
	if v, ok := tfMap["column_configurations"].([]interface{}); ok && len(v) > 0 {
		apiObject.ColumnConfigurations = expandColumnConfigurations(v)
	}
	if v, ok := tfMap["data_set_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.DataSetConfigurations = expandDataSetConfigurations(v)
	}
	if v, ok := tfMap["filter_groups"].([]interface{}); ok && len(v) > 0 {
		apiObject.FilterGroups = expandFilterGroups(v)
	}
	if v, ok := tfMap["parameters_declarations"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ParameterDeclarations = expandParameterDeclarations(v.List())
	}
	if v, ok := tfMap["sheets"].([]interface{}); ok && len(v) > 0 {
		apiObject.Sheets = expandSheetDefinitions(v)
	}

	return apiObject
}

func expandCalculatedFields(tfList []interface{}) []awstypes.CalculatedField {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.CalculatedField

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandCalculatedField(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandCalculatedField(tfMap map[string]interface{}) *awstypes.CalculatedField {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CalculatedField{}

	if v, ok := tfMap["data_set_identifier"].(string); ok && v != "" {
		apiObject.DataSetIdentifier = aws.String(v)
	}
	if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
		apiObject.Expression = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandColumnConfigurations(tfList []interface{}) []awstypes.ColumnConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ColumnConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandColumnConfiguration(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandColumnConfiguration(tfMap map[string]interface{}) *awstypes.ColumnConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ColumnConfiguration{}

	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}

	if v, ok := tfMap["format_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.FormatConfiguration = expandFormatConfiguration(v)
	}

	if v, ok := tfMap[names.AttrRole].(string); ok && v != "" {
		apiObject.Role = awstypes.ColumnRole(v)
	}

	return apiObject
}

func expandColumnIdentifier(tfList []interface{}) *awstypes.ColumnIdentifier {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return expandColumnIdentifierInternal(tfMap)
}

func expandColumnIdentifierInternal(tfMap map[string]interface{}) *awstypes.ColumnIdentifier {
	apiObject := &awstypes.ColumnIdentifier{}

	if v, ok := tfMap["data_set_identifier"].(string); ok && v != "" {
		apiObject.DataSetIdentifier = aws.String(v)
	}
	if v, ok := tfMap["column_name"].(string); ok && v != "" {
		apiObject.ColumnName = aws.String(v)
	}

	return apiObject
}

func expandColumnIdentifiers(tfList []interface{}) []awstypes.ColumnIdentifier {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ColumnIdentifier

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandColumnIdentifierInternal(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandDataSetConfigurations(tfList []interface{}) []awstypes.DataSetConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.DataSetConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandDataSetConfiguration(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandDataSetConfiguration(tfMap map[string]interface{}) *awstypes.DataSetConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DataSetConfiguration{}

	if v, ok := tfMap["column_group_schema_list"].([]interface{}); ok && len(v) > 0 {
		apiObject.ColumnGroupSchemaList = expandColumnGroupSchemas(v)
	}
	if v, ok := tfMap["data_set_schema"].([]interface{}); ok && len(v) > 0 {
		apiObject.DataSetSchema = expandDataSetSchema(v)
	}
	if v, ok := tfMap["placeholder"].(string); ok && v != "" {
		apiObject.Placeholder = aws.String(v)
	}

	return apiObject
}

func expandColumnGroupSchemas(tfList []interface{}) []awstypes.ColumnGroupSchema {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ColumnGroupSchema

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandColumnGroupSchema(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandColumnGroupSchema(tfMap map[string]interface{}) *awstypes.ColumnGroupSchema {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ColumnGroupSchema{}

	if v, ok := tfMap["column_group_schema_list"].([]interface{}); ok && len(v) > 0 {
		apiObject.ColumnGroupColumnSchemaList = expandColumnGroupColumnSchemas(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandColumnGroupColumnSchemas(tfList []interface{}) []awstypes.ColumnGroupColumnSchema {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ColumnGroupColumnSchema

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandColumnGroupColumnSchema(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandColumnGroupColumnSchema(tfMap map[string]interface{}) *awstypes.ColumnGroupColumnSchema {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ColumnGroupColumnSchema{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandDataSetSchema(tfList []interface{}) *awstypes.DataSetSchema {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DataSetSchema{}

	if v, ok := tfMap["column_schema_list"].([]interface{}); ok && len(v) > 0 {
		apiObject.ColumnSchemaList = expandColumnSchemas(v)
	}

	return apiObject
}

func expandColumnSchemas(tfList []interface{}) []awstypes.ColumnSchema {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ColumnSchema

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandColumnSchema(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandColumnSchema(tfMap map[string]interface{}) *awstypes.ColumnSchema {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ColumnSchema{}

	if v, ok := tfMap["data_type"].(string); ok && v != "" {
		apiObject.DataType = aws.String(v)
	}
	if v, ok := tfMap["geographic_role"].(string); ok && v != "" {
		apiObject.GeographicRole = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandFilterGroups(tfList []interface{}) []awstypes.FilterGroup {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.FilterGroup

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandFilterGroup(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandFilterGroup(tfMap map[string]interface{}) *awstypes.FilterGroup {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FilterGroup{}

	if v, ok := tfMap["cross_dataset"].(string); ok && v != "" {
		apiObject.CrossDataset = awstypes.CrossDatasetTypes(v)
	}
	if v, ok := tfMap["filter_group_id"].(string); ok && v != "" {
		apiObject.FilterGroupId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = awstypes.WidgetStatus(v)
	}
	if v, ok := tfMap["filters"].([]interface{}); ok && len(v) > 0 {
		apiObject.Filters = expandFilters(v)
	}
	if v, ok := tfMap["scope_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.ScopeConfiguration = expandFilterScopeConfiguration(v)
	}

	return apiObject
}

func expandAggregationFunction(tfList []interface{}) *awstypes.AggregationFunction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.AggregationFunction{}

	if v, ok := tfMap["categorical_aggregation_function"].(string); ok && v != "" {
		apiObject.CategoricalAggregationFunction = awstypes.CategoricalAggregationFunction(v)
	}
	if v, ok := tfMap["date_aggregation_function"].(string); ok && v != "" {
		apiObject.DateAggregationFunction = awstypes.DateAggregationFunction(v)
	}
	if v, ok := tfMap["numerical_aggregation_function"].([]interface{}); ok && len(v) > 0 {
		apiObject.NumericalAggregationFunction = expandNumericalAggregationFunction(v)
	}

	return apiObject
}

func expandNumericalAggregationFunction(tfList []interface{}) *awstypes.NumericalAggregationFunction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.NumericalAggregationFunction{}

	if v, ok := tfMap["simple_numerical_aggregation"].(string); ok && v != "" {
		apiObject.SimpleNumericalAggregation = awstypes.SimpleNumericalAggregationFunction(v)
	}
	if v, ok := tfMap["percentile_aggregation"].([]interface{}); ok && len(v) > 0 {
		apiObject.PercentileAggregation = expandPercentileAggregation(v)
	}

	return apiObject
}

func expandPercentileAggregation(tfList []interface{}) *awstypes.PercentileAggregation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.PercentileAggregation{}

	if v, ok := tfMap["simple_numerical_aggregation"].(float64); ok {
		apiObject.PercentileValue = aws.Float64(v)
	}

	return apiObject
}

func expandRollingDateConfiguration(tfList []interface{}) *awstypes.RollingDateConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.RollingDateConfiguration{}

	if v, ok := tfMap["data_set_identifier"].(string); ok && v != "" {
		apiObject.DataSetIdentifier = aws.String(v)
	}
	if v, ok := tfMap[names.AttrExpression].(string); ok {
		apiObject.Expression = aws.String(v)
	}

	return apiObject
}

func expandParameterDeclarations(tfList []interface{}) []awstypes.ParameterDeclaration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ParameterDeclaration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandParameterDeclaration(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandParameterDeclaration(tfMap map[string]interface{}) *awstypes.ParameterDeclaration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ParameterDeclaration{}

	if v, ok := tfMap["date_time_parameter_declaration"].([]interface{}); ok && len(v) > 0 {
		apiObject.DateTimeParameterDeclaration = expandDateTimeParameterDeclaration(v)
	}
	if v, ok := tfMap["decimal_parameter_declaration"].([]interface{}); ok && len(v) > 0 {
		apiObject.DecimalParameterDeclaration = expandDecimalParameterDeclaration(v)
	}
	if v, ok := tfMap["integer_parameter_declaration"].([]interface{}); ok && len(v) > 0 {
		apiObject.IntegerParameterDeclaration = expandIntegerParameterDeclaration(v)
	}
	if v, ok := tfMap["string_parameter_declaration"].([]interface{}); ok && len(v) > 0 {
		apiObject.StringParameterDeclaration = expandStringParameterDeclaration(v)
	}

	return apiObject
}

func expandSheetDefinitions(tfList []interface{}) []awstypes.SheetDefinition {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.SheetDefinition

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandSheetDefinition(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func FlattenTemplateDefinition(apiObject *awstypes.TemplateVersionDefinition) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.AnalysisDefaults != nil {
		tfMap["analysis_defaults"] = flattenAnalysisDefaults(apiObject.AnalysisDefaults)
	}
	if apiObject.CalculatedFields != nil {
		tfMap["calculated_fields"] = flattenCalculatedFields(apiObject.CalculatedFields)
	}
	if apiObject.ColumnConfigurations != nil {
		tfMap["column_configurations"] = flattenColumnConfigurations(apiObject.ColumnConfigurations)
	}
	if apiObject.DataSetConfigurations != nil {
		tfMap["data_set_configuration"] = flattenDataSetConfigurations(apiObject.DataSetConfigurations)
	}
	if apiObject.FilterGroups != nil {
		tfMap["filter_groups"] = flattenFilterGroups(apiObject.FilterGroups)
	}
	if apiObject.ParameterDeclarations != nil {
		tfMap["parameters_declarations"] = flattenParameterDeclarations(apiObject.ParameterDeclarations)
	}
	if apiObject.Sheets != nil {
		tfMap["sheets"] = flattenSheetDefinitions(apiObject.Sheets)
	}

	return []interface{}{tfMap}
}

func flattenCalculatedFields(apiObjects []awstypes.CalculatedField) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.DataSetIdentifier != nil {
			tfMap["data_set_identifier"] = aws.ToString(apiObject.DataSetIdentifier)
		}
		if apiObject.Expression != nil {
			tfMap[names.AttrExpression] = aws.ToString(apiObject.Expression)
		}
		if apiObject.Name != nil {
			tfMap[names.AttrName] = aws.ToString(apiObject.Name)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnConfigurations(apiObjects []awstypes.ColumnConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.Column != nil {
			tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
		}
		if apiObject.FormatConfiguration != nil {
			tfMap["format_configuration"] = flattenFormatConfiguration(apiObject.FormatConfiguration)
		}
		tfMap[names.AttrRole] = apiObject.Role

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnIdentifier(apiObject *awstypes.ColumnIdentifier) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.ToString(apiObject.ColumnName)
	}
	if apiObject.DataSetIdentifier != nil {
		tfMap["data_set_identifier"] = aws.ToString(apiObject.DataSetIdentifier)
	}

	return []interface{}{tfMap}
}

func flattenDataSetConfigurations(apiObjects []awstypes.DataSetConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.ColumnGroupSchemaList != nil {
			tfMap["column_group_schema_list"] = flattenColumnGroupSchemas(apiObject.ColumnGroupSchemaList)
		}
		if apiObject.DataSetSchema != nil {
			tfMap["data_set_schema"] = flattenDataSetSchema(apiObject.DataSetSchema)
		}
		if apiObject.Placeholder != nil {
			tfMap["placeholder"] = aws.ToString(apiObject.Placeholder)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnGroupSchemas(apiObjects []awstypes.ColumnGroupSchema) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.ColumnGroupColumnSchemaList != nil {
			tfMap["column_group_column_schema_list"] = flattenColumnGroupColumnSchemas(apiObject.ColumnGroupColumnSchemaList)
		}
		if apiObject.Name != nil {
			tfMap[names.AttrName] = aws.ToString(apiObject.Name)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnGroupColumnSchemas(apiObjects []awstypes.ColumnGroupColumnSchema) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.Name != nil {
			tfMap[names.AttrName] = aws.ToString(apiObject.Name)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDataSetSchema(apiObject *awstypes.DataSetSchema) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ColumnSchemaList != nil {
		tfMap["column_schema_list"] = flattenColumnSchemas(apiObject.ColumnSchemaList)
	}

	return []interface{}{tfMap}
}

func flattenColumnSchemas(apiObjects []awstypes.ColumnSchema) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.DataType != nil {
			tfMap["data_type"] = aws.ToString(apiObject.DataType)
		}
		if apiObject.GeographicRole != nil {
			tfMap["geographic_role"] = aws.ToString(apiObject.GeographicRole)
		}
		if apiObject.Name != nil {
			tfMap[names.AttrName] = aws.ToString(apiObject.Name)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenFilterGroups(apiObjects []awstypes.FilterGroup) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		tfMap["cross_dataset"] = apiObject.CrossDataset
		if apiObject.FilterGroupId != nil {
			tfMap["filter_group_id"] = aws.ToString(apiObject.FilterGroupId)
		}
		if apiObject.Filters != nil {
			tfMap["filters"] = flattenFilters(apiObject.Filters)
		}
		if apiObject.ScopeConfiguration != nil {
			tfMap["scope_configuration"] = flattenFilterScopeConfiguration(apiObject.ScopeConfiguration)
		}
		tfMap[names.AttrStatus] = apiObject.Status

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenAggregationFunction(apiObject *awstypes.AggregationFunction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["categorical_aggregation_function"] = apiObject.CategoricalAggregationFunction
	tfMap["date_aggregation_function"] = apiObject.DateAggregationFunction
	if apiObject.NumericalAggregationFunction != nil {
		tfMap["numerical_aggregation_function"] = flattenNumericalAggregationFunction(apiObject.NumericalAggregationFunction)
	}

	return []interface{}{tfMap}
}

func flattenNumericalAggregationFunction(apiObject *awstypes.NumericalAggregationFunction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.PercentileAggregation != nil {
		tfMap["percentile_aggregation"] = flattenPercentileAggregation(apiObject.PercentileAggregation)
	}
	tfMap["simple_numerical_aggregation"] = apiObject.SimpleNumericalAggregation

	return []interface{}{tfMap}
}

func flattenPercentileAggregation(apiObject *awstypes.PercentileAggregation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.PercentileValue != nil {
		tfMap["percentile_value"] = aws.ToFloat64(apiObject.PercentileValue)
	}

	return []interface{}{tfMap}
}

func flattenRollingDateConfiguration(apiObject *awstypes.RollingDateConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.DataSetIdentifier != nil {
		tfMap["data_set_identifier"] = aws.ToString(apiObject.DataSetIdentifier)
	}
	if apiObject.Expression != nil {
		tfMap[names.AttrExpression] = aws.ToString(apiObject.Expression)
	}

	return []interface{}{tfMap}
}

func flattenParameterDeclarations(apiObjects []awstypes.ParameterDeclaration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.DateTimeParameterDeclaration != nil {
			tfMap["date_time_parameter_declaration"] = flattenDateTimeParameterDeclaration(apiObject.DateTimeParameterDeclaration)
		}
		if apiObject.DecimalParameterDeclaration != nil {
			tfMap["decimal_parameter_declaration"] = flattenDecimalParameterDeclaration(apiObject.DecimalParameterDeclaration)
		}
		if apiObject.IntegerParameterDeclaration != nil {
			tfMap["integer_parameter_declaration"] = flattenIntegerParameterDeclaration(apiObject.IntegerParameterDeclaration)
		}
		if apiObject.StringParameterDeclaration != nil {
			tfMap["string_parameter_declaration"] = flattenStringParameterDeclaration(apiObject.StringParameterDeclaration)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSheetDefinitions(apiObjects []awstypes.SheetDefinition) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"sheet_id": aws.ToString(apiObject.SheetId),
		}

		tfMap[names.AttrContentType] = apiObject.ContentType
		if apiObject.Description != nil {
			tfMap[names.AttrDescription] = aws.ToString(apiObject.Description)
		}
		if apiObject.FilterControls != nil {
			tfMap["filter_controls"] = flattenFilterControls(apiObject.FilterControls)
		}
		if apiObject.Layouts != nil {
			tfMap["layouts"] = flattenLayouts(apiObject.Layouts)
		}
		if apiObject.Name != nil {
			tfMap[names.AttrName] = aws.ToString(apiObject.Name)
		}
		if apiObject.ParameterControls != nil {
			tfMap["parameter_controls"] = flattenParameterControls(apiObject.ParameterControls)
		}
		if apiObject.SheetControlLayouts != nil {
			tfMap["sheet_control_layouts"] = flattenSheetControlLayouts(apiObject.SheetControlLayouts)
		}
		if apiObject.TextBoxes != nil {
			tfMap["text_boxes"] = flattenTextBoxes(apiObject.TextBoxes)
		}
		if apiObject.Title != nil {
			tfMap["title"] = aws.ToString(apiObject.Title)
		}
		if apiObject.Visuals != nil {
			tfMap["visuals"] = flattenVisuals(apiObject.Visuals)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTextBoxes(apiObjects []awstypes.SheetTextBox) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"sheet_text_box_id": aws.ToString(apiObject.SheetTextBoxId),
		}

		if apiObject.Content != nil {
			tfMap[names.AttrContent] = aws.ToString(apiObject.Content)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
