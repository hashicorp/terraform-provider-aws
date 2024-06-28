// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
							names.AttrRole:         stringSchema(false, validation.StringInSlice(quicksight.ColumnRole_Values(), false)),
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
							"cross_dataset":       stringSchema(true, validation.StringInSlice(quicksight.CrossDatasetTypes_Values(), false)),
							"filter_group_id":     idSchema(),
							"filters":             filtersSchema(),                  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Filter.html
							"scope_configuration": filterScopeConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterScopeConfiguration.html
							names.AttrStatus:      stringSchema(false, validation.StringInSlice(quicksight.Status_Values(), false)),
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
								Type:         schema.TypeString,
								Optional:     true,
								Computed:     true,
								ValidateFunc: validation.StringInSlice(quicksight.SheetContentType_Values(), false),
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

func stringOptionalComputedSchema(validateFunc schema.SchemaValidateFunc) *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validateFunc,
	}
}

func stringSchema(required bool, validateFunc schema.SchemaValidateFunc) *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Required:     required,
		Optional:     !required,
		ValidateFunc: validateFunc,
	}
}

func intSchema(required bool, validateFunc schema.SchemaValidateFunc) *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeInt,
		Required:     required,
		Optional:     !required,
		ValidateFunc: validateFunc,
	}
}

func floatSchema(required bool, validateFunc schema.SchemaValidateFunc) *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeFloat,
		Required:     required,
		Optional:     !required,
		ValidateFunc: validateFunc,
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
				"categorical_aggregation_function": stringSchema(false, validation.StringInSlice(quicksight.CategoricalAggregationFunction_Values(), false)),
				"date_aggregation_function":        stringSchema(false, validation.StringInSlice(quicksight.DateAggregationFunction_Values(), false)),
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
				"simple_numerical_aggregation": stringSchema(false, validation.StringInSlice(quicksight.SimpleNumericalAggregationFunction_Values(), false)),
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

func ExpandTemplateSourceEntity(tfList []interface{}) *quicksight.TemplateSourceEntity {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	sourceEntity := &quicksight.TemplateSourceEntity{}

	if v, ok := tfMap["source_analysis"].([]interface{}); ok && len(v) > 0 {
		sourceEntity.SourceAnalysis = expandSourceAnalysis(v[0].(map[string]interface{}))
	} else if v, ok := tfMap["source_template"].([]interface{}); ok && len(v) > 0 {
		sourceEntity.SourceTemplate = expandTemplateSourceTemplate(v[0].(map[string]interface{}))
	}

	return sourceEntity
}

func expandSourceAnalysis(tfMap map[string]interface{}) *quicksight.TemplateSourceAnalysis {
	if tfMap == nil {
		return nil
	}

	sourceAnalysis := &quicksight.TemplateSourceAnalysis{}
	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		sourceAnalysis.Arn = aws.String(v)
	}
	if v, ok := tfMap["data_set_references"].([]interface{}); ok && len(v) > 0 {
		sourceAnalysis.DataSetReferences = expandDataSetReferences(v)
	}

	return sourceAnalysis
}

func expandDataSetReferences(tfList []interface{}) []*quicksight.DataSetReference {
	if len(tfList) == 0 {
		return nil
	}

	var dataSetReferences []*quicksight.DataSetReference
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		dataSetReference := expandDataSetReference(tfMap)
		if dataSetReference == nil {
			continue
		}

		dataSetReferences = append(dataSetReferences, dataSetReference)
	}

	return dataSetReferences
}

func expandDataSetReference(tfMap map[string]interface{}) *quicksight.DataSetReference {
	if tfMap == nil {
		return nil
	}

	dataSetReference := &quicksight.DataSetReference{}
	if v, ok := tfMap["data_set_arn"].(string); ok {
		dataSetReference.DataSetArn = aws.String(v)
	}
	if v, ok := tfMap["data_set_placeholder"].(string); ok {
		dataSetReference.DataSetPlaceholder = aws.String(v)
	}

	return dataSetReference
}

func expandTemplateSourceTemplate(tfMap map[string]interface{}) *quicksight.TemplateSourceTemplate {
	if tfMap == nil {
		return nil
	}

	sourceTemplate := &quicksight.TemplateSourceTemplate{}
	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		sourceTemplate.Arn = aws.String(v)
	}

	return sourceTemplate
}

func ExpandTemplateDefinition(tfList []interface{}) *quicksight.TemplateVersionDefinition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	definition := &quicksight.TemplateVersionDefinition{}

	if v, ok := tfMap["analysis_defaults"].([]interface{}); ok && len(v) > 0 {
		definition.AnalysisDefaults = expandAnalysisDefaults(v)
	}
	if v, ok := tfMap["calculated_fields"].(*schema.Set); ok && v.Len() > 0 {
		definition.CalculatedFields = expandCalculatedFields(v.List())
	}
	if v, ok := tfMap["column_configurations"].([]interface{}); ok && len(v) > 0 {
		definition.ColumnConfigurations = expandColumnConfigurations(v)
	}
	if v, ok := tfMap["data_set_configuration"].([]interface{}); ok && len(v) > 0 {
		definition.DataSetConfigurations = expandDataSetConfigurations(v)
	}
	if v, ok := tfMap["filter_groups"].([]interface{}); ok && len(v) > 0 {
		definition.FilterGroups = expandFilterGroups(v)
	}
	if v, ok := tfMap["parameters_declarations"].(*schema.Set); ok && v.Len() > 0 {
		definition.ParameterDeclarations = expandParameterDeclarations(v.List())
	}
	if v, ok := tfMap["sheets"].([]interface{}); ok && len(v) > 0 {
		definition.Sheets = expandSheetDefinitions(v)
	}

	return definition
}

func expandCalculatedFields(tfList []interface{}) []*quicksight.CalculatedField {
	if len(tfList) == 0 {
		return nil
	}

	var fields []*quicksight.CalculatedField
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		field := expandCalculatedField(tfMap)
		if field == nil {
			continue
		}

		fields = append(fields, field)
	}

	return fields
}

func expandCalculatedField(tfMap map[string]interface{}) *quicksight.CalculatedField {
	if tfMap == nil {
		return nil
	}

	field := &quicksight.CalculatedField{}

	if v, ok := tfMap["data_set_identifier"].(string); ok && v != "" {
		field.DataSetIdentifier = aws.String(v)
	}
	if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
		field.Expression = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		field.Name = aws.String(v)
	}

	return field
}

func expandColumnConfigurations(tfList []interface{}) []*quicksight.ColumnConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var configs []*quicksight.ColumnConfiguration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		column := expandColumnConfiguration(tfMap)
		if column == nil {
			continue
		}

		configs = append(configs, column)
	}

	return configs
}

func expandColumnConfiguration(tfMap map[string]interface{}) *quicksight.ColumnConfiguration {
	if tfMap == nil {
		return nil
	}

	column := &quicksight.ColumnConfiguration{}

	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		column.Column = expandColumnIdentifier(v)
	}

	if v, ok := tfMap["format_configuration"].([]interface{}); ok && len(v) > 0 {
		column.FormatConfiguration = expandFormatConfiguration(v)
	}

	if v, ok := tfMap[names.AttrRole].(string); ok && v != "" {
		column.Role = aws.String(v)
	}

	return column
}

func expandColumnIdentifier(tfList []interface{}) *quicksight.ColumnIdentifier {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return expandColumnIdentifierInternal(tfMap)
}

func expandColumnIdentifierInternal(tfMap map[string]interface{}) *quicksight.ColumnIdentifier {
	column := &quicksight.ColumnIdentifier{}

	if v, ok := tfMap["data_set_identifier"].(string); ok && v != "" {
		column.DataSetIdentifier = aws.String(v)
	}
	if v, ok := tfMap["column_name"].(string); ok && v != "" {
		column.ColumnName = aws.String(v)
	}

	return column
}

func expandColumnIdentifiers(tfList []interface{}) []*quicksight.ColumnIdentifier {
	if len(tfList) == 0 {
		return nil
	}

	var columns []*quicksight.ColumnIdentifier
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		col := expandColumnIdentifierInternal(tfMap)
		if col == nil {
			continue
		}

		columns = append(columns, col)
	}

	return columns
}

func expandDataSetConfigurations(tfList []interface{}) []*quicksight.DataSetConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var configs []*quicksight.DataSetConfiguration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		config := expandDataSetConfiguration(tfMap)
		if config == nil {
			continue
		}

		configs = append(configs, config)
	}

	return configs
}

func expandDataSetConfiguration(tfMap map[string]interface{}) *quicksight.DataSetConfiguration {
	if tfMap == nil {
		return nil
	}

	config := &quicksight.DataSetConfiguration{}

	if v, ok := tfMap["column_group_schema_list"].([]interface{}); ok && len(v) > 0 {
		config.ColumnGroupSchemaList = expandColumnGroupSchemas(v)
	}
	if v, ok := tfMap["data_set_schema"].([]interface{}); ok && len(v) > 0 {
		config.DataSetSchema = expandDataSetSchema(v)
	}
	if v, ok := tfMap["placeholder"].(string); ok && v != "" {
		config.Placeholder = aws.String(v)
	}

	return config
}

func expandColumnGroupSchemas(tfList []interface{}) []*quicksight.ColumnGroupSchema {
	if len(tfList) == 0 {
		return nil
	}

	var groups []*quicksight.ColumnGroupSchema
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		group := expandColumnGroupSchema(tfMap)
		if group == nil {
			continue
		}

		groups = append(groups, group)
	}

	return groups
}

func expandColumnGroupSchema(tfMap map[string]interface{}) *quicksight.ColumnGroupSchema {
	if tfMap == nil {
		return nil
	}

	group := &quicksight.ColumnGroupSchema{}

	if v, ok := tfMap["column_group_schema_list"].([]interface{}); ok && len(v) > 0 {
		group.ColumnGroupColumnSchemaList = expandColumnGroupColumnSchemas(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		group.Name = aws.String(v)
	}

	return group
}

func expandColumnGroupColumnSchemas(tfList []interface{}) []*quicksight.ColumnGroupColumnSchema {
	if len(tfList) == 0 {
		return nil
	}

	var columns []*quicksight.ColumnGroupColumnSchema
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		column := expandColumnGroupColumnSchema(tfMap)
		if column == nil {
			continue
		}

		columns = append(columns, column)
	}

	return columns
}

func expandColumnGroupColumnSchema(tfMap map[string]interface{}) *quicksight.ColumnGroupColumnSchema {
	if tfMap == nil {
		return nil
	}

	column := &quicksight.ColumnGroupColumnSchema{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		column.Name = aws.String(v)
	}

	return column
}

func expandDataSetSchema(tfList []interface{}) *quicksight.DataSetSchema {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}
	schema := &quicksight.DataSetSchema{}

	if v, ok := tfMap["column_schema_list"].([]interface{}); ok && len(v) > 0 {
		schema.ColumnSchemaList = expandColumnSchemas(v)
	}

	return schema
}

func expandColumnSchemas(tfList []interface{}) []*quicksight.ColumnSchema {
	if len(tfList) == 0 {
		return nil
	}

	var columns []*quicksight.ColumnSchema
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		column := expandColumnSchema(tfMap)
		if column == nil {
			continue
		}

		columns = append(columns, column)
	}

	return columns
}

func expandColumnSchema(tfMap map[string]interface{}) *quicksight.ColumnSchema {
	if tfMap == nil {
		return nil
	}

	column := &quicksight.ColumnSchema{}

	if v, ok := tfMap["data_type"].(string); ok && v != "" {
		column.DataType = aws.String(v)
	}
	if v, ok := tfMap["geographic_role"].(string); ok && v != "" {
		column.GeographicRole = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		column.Name = aws.String(v)
	}

	return column
}

func expandFilterGroups(tfList []interface{}) []*quicksight.FilterGroup {
	if len(tfList) == 0 {
		return nil
	}

	var groups []*quicksight.FilterGroup
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		group := expandFilterGroup(tfMap)
		if group == nil {
			continue
		}

		groups = append(groups, group)
	}

	return groups
}

func expandFilterGroup(tfMap map[string]interface{}) *quicksight.FilterGroup {
	if tfMap == nil {
		return nil
	}

	group := &quicksight.FilterGroup{}

	if v, ok := tfMap["cross_dataset"].(string); ok && v != "" {
		group.CrossDataset = aws.String(v)
	}
	if v, ok := tfMap["filter_group_id"].(string); ok && v != "" {
		group.FilterGroupId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		group.Status = aws.String(v)
	}
	if v, ok := tfMap["filters"].([]interface{}); ok && len(v) > 0 {
		group.Filters = expandFilters(v)
	}
	if v, ok := tfMap["scope_configuration"].([]interface{}); ok && len(v) > 0 {
		group.ScopeConfiguration = expandFilterScopeConfiguration(v)
	}

	return group
}

func expandAggregationFunction(tfList []interface{}) *quicksight.AggregationFunction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	function := &quicksight.AggregationFunction{}

	if v, ok := tfMap["categorical_aggregation_function"].(string); ok && v != "" {
		function.CategoricalAggregationFunction = aws.String(v)
	}
	if v, ok := tfMap["date_aggregation_function"].(string); ok && v != "" {
		function.DateAggregationFunction = aws.String(v)
	}
	if v, ok := tfMap["numerical_aggregation_function"].([]interface{}); ok && len(v) > 0 {
		function.NumericalAggregationFunction = expandNumericalAggregationFunction(v)
	}

	return function
}

func expandNumericalAggregationFunction(tfList []interface{}) *quicksight.NumericalAggregationFunction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	function := &quicksight.NumericalAggregationFunction{}

	if v, ok := tfMap["simple_numerical_aggregation"].(string); ok && v != "" {
		function.SimpleNumericalAggregation = aws.String(v)
	}
	if v, ok := tfMap["percentile_aggregation"].([]interface{}); ok && len(v) > 0 {
		function.PercentileAggregation = expandPercentileAggregation(v)
	}

	return function
}

func expandPercentileAggregation(tfList []interface{}) *quicksight.PercentileAggregation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	agg := &quicksight.PercentileAggregation{}

	if v, ok := tfMap["simple_numerical_aggregation"].(float64); ok {
		agg.PercentileValue = aws.Float64(v)
	}

	return agg
}

func expandRollingDateConfiguration(tfList []interface{}) *quicksight.RollingDateConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.RollingDateConfiguration{}

	if v, ok := tfMap["data_set_identifier"].(string); ok && v != "" {
		config.DataSetIdentifier = aws.String(v)
	}
	if v, ok := tfMap[names.AttrExpression].(string); ok {
		config.Expression = aws.String(v)
	}

	return config
}

func expandParameterDeclarations(tfList []interface{}) []*quicksight.ParameterDeclaration {
	if len(tfList) == 0 {
		return nil
	}

	var params []*quicksight.ParameterDeclaration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		param := expandParameterDeclaration(tfMap)
		if param == nil {
			continue
		}

		params = append(params, param)
	}

	return params
}

func expandParameterDeclaration(tfMap map[string]interface{}) *quicksight.ParameterDeclaration {
	if tfMap == nil {
		return nil
	}

	param := &quicksight.ParameterDeclaration{}

	if v, ok := tfMap["date_time_parameter_declaration"].([]interface{}); ok && len(v) > 0 {
		param.DateTimeParameterDeclaration = expandDateTimeParameterDeclaration(v)
	}
	if v, ok := tfMap["decimal_parameter_declaration"].([]interface{}); ok && len(v) > 0 {
		param.DecimalParameterDeclaration = expandDecimalParameterDeclaration(v)
	}
	if v, ok := tfMap["integer_parameter_declaration"].([]interface{}); ok && len(v) > 0 {
		param.IntegerParameterDeclaration = expandIntegerParameterDeclaration(v)
	}
	if v, ok := tfMap["string_parameter_declaration"].([]interface{}); ok && len(v) > 0 {
		param.StringParameterDeclaration = expandStringParameterDeclaration(v)
	}

	return param
}

func expandSheetDefinitions(tfList []interface{}) []*quicksight.SheetDefinition {
	if len(tfList) == 0 {
		return nil
	}

	var sheets []*quicksight.SheetDefinition
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		sheet := expandSheetDefinition(tfMap)
		if sheet == nil {
			continue
		}

		sheets = append(sheets, sheet)
	}

	return sheets
}

func FlattenTemplateDefinition(apiObject *quicksight.TemplateVersionDefinition) []interface{} {
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

func flattenCalculatedFields(apiObject []*quicksight.CalculatedField) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, field := range apiObject {
		if field == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if field.DataSetIdentifier != nil {
			tfMap["data_set_identifier"] = aws.StringValue(field.DataSetIdentifier)
		}
		if field.Expression != nil {
			tfMap[names.AttrExpression] = aws.StringValue(field.Expression)
		}
		if field.Name != nil {
			tfMap[names.AttrName] = aws.StringValue(field.Name)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnConfigurations(apiObject []*quicksight.ColumnConfiguration) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, column := range apiObject {
		if column == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if column.Column != nil {
			tfMap["column"] = flattenColumnIdentifier(column.Column)
		}
		if column.FormatConfiguration != nil {
			tfMap["format_configuration"] = flattenFormatConfiguration(column.FormatConfiguration)
		}
		if column.Role != nil {
			tfMap[names.AttrRole] = aws.StringValue(column.Role)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnIdentifier(apiObject *quicksight.ColumnIdentifier) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ColumnName != nil {
		tfMap["column_name"] = aws.StringValue(apiObject.ColumnName)
	}
	if apiObject.DataSetIdentifier != nil {
		tfMap["data_set_identifier"] = aws.StringValue(apiObject.DataSetIdentifier)
	}

	return []interface{}{tfMap}
}

func flattenDataSetConfigurations(apiObject []*quicksight.DataSetConfiguration) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.ColumnGroupSchemaList != nil {
			tfMap["column_group_schema_list"] = flattenColumnGroupSchemas(config.ColumnGroupSchemaList)
		}
		if config.DataSetSchema != nil {
			tfMap["data_set_schema"] = flattenDataSetSchema(config.DataSetSchema)
		}
		if config.Placeholder != nil {
			tfMap["placeholder"] = aws.StringValue(config.Placeholder)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnGroupSchemas(apiObject []*quicksight.ColumnGroupSchema) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, schema := range apiObject {
		if schema == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if schema.ColumnGroupColumnSchemaList != nil {
			tfMap["column_group_column_schema_list"] = flattenColumnGroupColumnSchemas(schema.ColumnGroupColumnSchemaList)
		}
		if schema.Name != nil {
			tfMap[names.AttrName] = aws.StringValue(schema.Name)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenColumnGroupColumnSchemas(apiObject []*quicksight.ColumnGroupColumnSchema) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, schema := range apiObject {
		if schema == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if schema.Name != nil {
			tfMap[names.AttrName] = aws.StringValue(schema.Name)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDataSetSchema(apiObject *quicksight.DataSetSchema) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ColumnSchemaList != nil {
		tfMap["column_schema_list"] = flattenColumnSchemas(apiObject.ColumnSchemaList)
	}

	return []interface{}{tfMap}
}

func flattenColumnSchemas(apiObject []*quicksight.ColumnSchema) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, column := range apiObject {
		if column == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if column.DataType != nil {
			tfMap["data_type"] = aws.StringValue(column.DataType)
		}
		if column.GeographicRole != nil {
			tfMap["geographic_role"] = aws.StringValue(column.GeographicRole)
		}
		if column.Name != nil {
			tfMap[names.AttrName] = aws.StringValue(column.Name)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenFilterGroups(apiObject []*quicksight.FilterGroup) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, group := range apiObject {
		if group == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if group.CrossDataset != nil {
			tfMap["cross_dataset"] = aws.StringValue(group.CrossDataset)
		}
		if group.FilterGroupId != nil {
			tfMap["filter_group_id"] = aws.StringValue(group.FilterGroupId)
		}
		if group.Filters != nil {
			tfMap["filters"] = flattenFilters(group.Filters)
		}
		if group.ScopeConfiguration != nil {
			tfMap["scope_configuration"] = flattenFilterScopeConfiguration(group.ScopeConfiguration)
		}
		if group.Status != nil {
			tfMap[names.AttrStatus] = aws.StringValue(group.Status)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenAggregationFunction(apiObject *quicksight.AggregationFunction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CategoricalAggregationFunction != nil {
		tfMap["categorical_aggregation_function"] = aws.StringValue(apiObject.CategoricalAggregationFunction)
	}
	if apiObject.DateAggregationFunction != nil {
		tfMap["date_aggregation_function"] = aws.StringValue(apiObject.DateAggregationFunction)
	}
	if apiObject.NumericalAggregationFunction != nil {
		tfMap["numerical_aggregation_function"] = flattenNumericalAggregationFunction(apiObject.NumericalAggregationFunction)
	}

	return []interface{}{tfMap}
}

func flattenNumericalAggregationFunction(apiObject *quicksight.NumericalAggregationFunction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.PercentileAggregation != nil {
		tfMap["percentile_aggregation"] = flattenPercentileAggregation(apiObject.PercentileAggregation)
	}
	if apiObject.SimpleNumericalAggregation != nil {
		tfMap["simple_numerical_aggregation"] = aws.StringValue(apiObject.SimpleNumericalAggregation)
	}

	return []interface{}{tfMap}
}

func flattenPercentileAggregation(apiObject *quicksight.PercentileAggregation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.PercentileValue != nil {
		tfMap["percentile_value"] = aws.Float64Value(apiObject.PercentileValue)
	}

	return []interface{}{tfMap}
}

func flattenRollingDateConfiguration(apiObject *quicksight.RollingDateConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DataSetIdentifier != nil {
		tfMap["data_set_identifier"] = aws.StringValue(apiObject.DataSetIdentifier)
	}
	if apiObject.Expression != nil {
		tfMap[names.AttrExpression] = aws.StringValue(apiObject.Expression)
	}

	return []interface{}{tfMap}
}

func flattenParameterDeclarations(apiObject []*quicksight.ParameterDeclaration) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.DateTimeParameterDeclaration != nil {
			tfMap["date_time_parameter_declaration"] = flattenDateTimeParameterDeclaration(config.DateTimeParameterDeclaration)
		}
		if config.DecimalParameterDeclaration != nil {
			tfMap["decimal_parameter_declaration"] = flattenDecimalParameterDeclaration(config.DecimalParameterDeclaration)
		}
		if config.IntegerParameterDeclaration != nil {
			tfMap["integer_parameter_declaration"] = flattenIntegerParameterDeclaration(config.IntegerParameterDeclaration)
		}
		if config.StringParameterDeclaration != nil {
			tfMap["string_parameter_declaration"] = flattenStringParameterDeclaration(config.StringParameterDeclaration)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSheetDefinitions(apiObject []*quicksight.SheetDefinition) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"sheet_id": aws.StringValue(config.SheetId),
		}
		if config.ContentType != nil {
			tfMap[names.AttrContentType] = aws.StringValue(config.ContentType)
		}
		if config.Description != nil {
			tfMap[names.AttrDescription] = aws.StringValue(config.Description)
		}
		if config.FilterControls != nil {
			tfMap["filter_controls"] = flattenFilterControls(config.FilterControls)
		}
		if config.Layouts != nil {
			tfMap["layouts"] = flattenLayouts(config.Layouts)
		}
		if config.Name != nil {
			tfMap[names.AttrName] = aws.StringValue(config.Name)
		}
		if config.ParameterControls != nil {
			tfMap["parameter_controls"] = flattenParameterControls(config.ParameterControls)
		}
		if config.SheetControlLayouts != nil {
			tfMap["sheet_control_layouts"] = flattenSheetControlLayouts(config.SheetControlLayouts)
		}
		if config.TextBoxes != nil {
			tfMap["text_boxes"] = flattenTextBoxes(config.TextBoxes)
		}
		if config.Title != nil {
			tfMap["title"] = aws.StringValue(config.Title)
		}
		if config.Visuals != nil {
			tfMap["visuals"] = flattenVisuals(config.Visuals)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenTextBoxes(apiObject []*quicksight.SheetTextBox) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"sheet_text_box_id": aws.StringValue(config.SheetTextBoxId),
		}
		if config.Content != nil {
			tfMap[names.AttrContent] = aws.StringValue(config.Content)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}
