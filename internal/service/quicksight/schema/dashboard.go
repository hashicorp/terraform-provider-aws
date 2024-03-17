// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DashboardDefinitionSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DashboardVersionDefinition.html
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
				"data_set_identifiers_declarations": dataSetIdentifierDeclarationsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataSetIdentifierDeclaration.html
				"analysis_defaults":                 analysisDefaultSchema(),               // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AnalysisDefaults.html
				"calculated_fields":                 calculatedFieldsSchema(),              // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CalculatedField.html
				"column_configurations": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnConfiguration.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 200,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"column":               columnSchema(true),          // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
							"format_configuration": formatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FormatConfiguration.html
							"role":                 stringSchema(false, enum.Validate[types.ColumnRole]()),
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
							"cross_dataset":       stringSchema(true, enum.Validate[types.CrossDatasetTypes]()),
							"filter_group_id":     idSchema(),
							"filters":             filtersSchema(),                  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Filter.html
							"scope_configuration": filterScopeConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterScopeConfiguration.html
							"status":              stringSchema(false, enum.Validate[types.WidgetStatus]()),
						},
					},
				},
				"parameter_declarations": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterDeclaration.html
					Type:     schema.TypeSet,
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
							"content_type": {
								Type:             schema.TypeString,
								Optional:         true,
								Computed:         true,
								ValidateDiagFunc: enum.Validate[types.SheetContentType](),
							},
							"description":           stringSchema(false, validation.ToDiagFunc(validation.StringLenBetween(1, 1024))),
							"filter_controls":       filterControlsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterControl.html
							"layouts":               layoutSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Layout.html
							"name":                  stringSchema(false, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
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
										"content":           stringSchema(false, validation.ToDiagFunc(validation.StringLenBetween(1, 150000))),
									},
								},
							},
							"title":   stringSchema(false, validation.ToDiagFunc(validation.StringLenBetween(1, 1024))),
							"visuals": visualsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Visual.html
						},
					},
				},
			},
		},
	}
}

func DashboardPublishOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DashboardPublishOptions.html
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"ad_hoc_filtering_option": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AdHocFilteringOption.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"availability_status": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.DashboardBehaviorEnabled,
								ValidateDiagFunc: enum.Validate[types.DashboardBehavior](),
							},
						},
					},
				},
				"data_point_drill_up_down_option": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataPointDrillUpDownOption.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"availability_status": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.DashboardBehaviorEnabled,
								ValidateDiagFunc: enum.Validate[types.DashboardBehavior](),
							}},
					},
				},
				"data_point_menu_label_option": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataPointMenuLabelOption.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"availability_status": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.DashboardBehaviorEnabled,
								ValidateDiagFunc: enum.Validate[types.DashboardBehavior](),
							}},
					},
				},
				"data_point_tooltip_option": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataPointTooltipOption.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"availability_status": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.DashboardBehaviorEnabled,
								ValidateDiagFunc: enum.Validate[types.DashboardBehavior](),
							}},
					},
				},
				"export_to_csv_option": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ExportToCSVOption.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"availability_status": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.DashboardBehaviorEnabled,
								ValidateDiagFunc: enum.Validate[types.DashboardBehavior](),
							}},
					},
				},
				"export_with_hidden_fields_option": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ExportWithHiddenFieldsOption.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"availability_status": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.DashboardBehaviorEnabled,
								ValidateDiagFunc: enum.Validate[types.DashboardBehavior](),
							}},
					},
				},
				"sheet_controls_option": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetControlsOption.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"visibility_state": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.DashboardUIStateCollapsed,
								ValidateDiagFunc: enum.Validate[types.DashboardUIState](),
							},
						},
					},
				},
				"sheet_layout_element_maximization_option": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetLayoutElementMaximizationOption.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"availability_status": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.DashboardBehaviorEnabled,
								ValidateDiagFunc: enum.Validate[types.DashboardBehavior](),
							}},
					},
				},
				"visual_axis_sort_option": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualAxisSortOption.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"availability_status": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.DashboardBehaviorEnabled,
								ValidateDiagFunc: enum.Validate[types.DashboardBehavior](),
							}},
					},
				},
				"visual_menu_option": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualMenuOption.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"availability_status": {
								Type:             schema.TypeString,
								Optional:         true,
								Default:          types.DashboardBehaviorEnabled,
								ValidateDiagFunc: enum.Validate[types.DashboardBehavior](),
							}},
					},
				},
			},
		},
	}
}

func DashboardSourceEntitySchema() *schema.Schema {
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
				"source_template": {
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"arn": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
							},
							"data_set_references": dataSetReferencesSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataSetReference.html
						},
					},
				},
			},
		},
	}
}

func ExpandDashboardSourceEntity(tfList []interface{}) *types.DashboardSourceEntity {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	sourceEntity := &types.DashboardSourceEntity{}

	if v, ok := tfMap["source_template"].([]interface{}); ok && len(v) > 0 {
		sourceEntity.SourceTemplate = expandDashboardSourceTemplate(v[0].(map[string]interface{}))
	}

	return sourceEntity
}

func expandDashboardSourceTemplate(tfMap map[string]interface{}) *types.DashboardSourceTemplate {
	if tfMap == nil {
		return nil
	}

	sourceTemplate := &types.DashboardSourceTemplate{}
	if v, ok := tfMap["arn"].(string); ok && v != "" {
		sourceTemplate.Arn = aws.String(v)
	}
	if v, ok := tfMap["data_set_references"].([]interface{}); ok && len(v) > 0 {
		sourceTemplate.DataSetReferences = expandDataSetReferences(v)
	}

	return sourceTemplate
}

func ExpandDashboardDefinition(tfList []interface{}) *types.DashboardVersionDefinition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	definition := &types.DashboardVersionDefinition{}

	if v, ok := tfMap["analysis_defaults"].([]interface{}); ok && len(v) > 0 {
		definition.AnalysisDefaults = expandAnalysisDefaults(v)
	}
	if v, ok := tfMap["calculated_fields"].(*schema.Set); ok && v.Len() > 0 {
		definition.CalculatedFields = expandCalculatedFields(v.List())
	}
	if v, ok := tfMap["column_configurations"].([]interface{}); ok && len(v) > 0 {
		definition.ColumnConfigurations = expandColumnConfigurations(v)
	}
	if v, ok := tfMap["data_set_identifiers_declarations"].([]interface{}); ok && len(v) > 0 {
		definition.DataSetIdentifierDeclarations = expandDataSetIdentifierDeclarations(v)
	}
	if v, ok := tfMap["filter_groups"].([]interface{}); ok && len(v) > 0 {
		definition.FilterGroups = expandFilterGroups(v)
	}
	if v, ok := tfMap["parameter_declarations"].(*schema.Set); ok && v.Len() > 0 {
		definition.ParameterDeclarations = expandParameterDeclarations(v.List())
	}
	if v, ok := tfMap["sheets"].([]interface{}); ok && len(v) > 0 {
		definition.Sheets = expandSheetDefinitions(v)
	}

	return definition
}

func ExpandDashboardPublishOptions(tfList []interface{}) *types.DashboardPublishOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.DashboardPublishOptions{}

	if v, ok := tfMap["ad_hoc_filtering_option"].([]interface{}); ok && len(v) > 0 {
		options.AdHocFilteringOption = expandAdHocFilteringOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["data_point_drill_up_down_option"].([]interface{}); ok && len(v) > 0 {
		options.DataPointDrillUpDownOption = expandDataPointDrillUpDownOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["data_point_menu_label_option"].([]interface{}); ok && len(v) > 0 {
		options.DataPointMenuLabelOption = expandDataPointMenuLabelOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["data_point_tooltip_option"].([]interface{}); ok && len(v) > 0 {
		options.DataPointTooltipOption = expandDataPointTooltipOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["export_to_csv_option"].([]interface{}); ok && len(v) > 0 {
		options.ExportToCSVOption = expandExportToCSVOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["export_with_hidden_fields_option"].([]interface{}); ok && len(v) > 0 {
		options.ExportWithHiddenFieldsOption = expandExportWithHiddenFieldsOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["sheet_controls_option"].([]interface{}); ok && len(v) > 0 {
		options.SheetControlsOption = expandSheetControlsOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["sheet_layout_element_maximization_option"].([]interface{}); ok && len(v) > 0 {
		options.SheetLayoutElementMaximizationOption = expandSheetLayoutElementMaximizationOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["visual_axis_sort_option"].([]interface{}); ok && len(v) > 0 {
		options.VisualAxisSortOption = expandVisualAxisSortOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["visual_menu_option"].([]interface{}); ok && len(v) > 0 {
		options.VisualMenuOption = expandVisualMenuOption(v[0].(map[string]interface{}))
	}

	return options
}

func expandAdHocFilteringOption(tfMap map[string]interface{}) *types.AdHocFilteringOption {
	if tfMap == nil {
		return nil
	}

	options := &types.AdHocFilteringOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = types.DashboardBehavior(v)
	}

	return options
}

func expandDataPointDrillUpDownOption(tfMap map[string]interface{}) *types.DataPointDrillUpDownOption {
	if tfMap == nil {
		return nil
	}

	options := &types.DataPointDrillUpDownOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = types.DashboardBehavior(v)
	}

	return options
}

func expandDataPointMenuLabelOption(tfMap map[string]interface{}) *types.DataPointMenuLabelOption {
	if tfMap == nil {
		return nil
	}

	options := &types.DataPointMenuLabelOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = types.DashboardBehavior(v)
	}

	return options
}

func expandDataPointTooltipOption(tfMap map[string]interface{}) *types.DataPointTooltipOption {
	if tfMap == nil {
		return nil
	}

	options := &types.DataPointTooltipOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = types.DashboardBehavior(v)
	}

	return options
}

func expandExportToCSVOption(tfMap map[string]interface{}) *types.ExportToCSVOption {
	if tfMap == nil {
		return nil
	}

	options := &types.ExportToCSVOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = types.DashboardBehavior(v)
	}

	return options
}

func expandExportWithHiddenFieldsOption(tfMap map[string]interface{}) *types.ExportWithHiddenFieldsOption {
	if tfMap == nil {
		return nil
	}

	options := &types.ExportWithHiddenFieldsOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = types.DashboardBehavior(v)
	}

	return options
}

func expandSheetLayoutElementMaximizationOption(tfMap map[string]interface{}) *types.SheetLayoutElementMaximizationOption {
	if tfMap == nil {
		return nil
	}

	options := &types.SheetLayoutElementMaximizationOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = types.DashboardBehavior(v)
	}

	return options
}

func expandSheetControlsOption(tfMap map[string]interface{}) *types.SheetControlsOption {
	if tfMap == nil {
		return nil
	}

	options := &types.SheetControlsOption{}
	if v, ok := tfMap["visibility_state"].(string); ok && v != "" {
		options.VisibilityState = types.DashboardUIState(v)
	}

	return options
}

func expandVisualAxisSortOption(tfMap map[string]interface{}) *types.VisualAxisSortOption {
	if tfMap == nil {
		return nil
	}

	options := &types.VisualAxisSortOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = types.DashboardBehavior(v)
	}

	return options
}

func expandVisualMenuOption(tfMap map[string]interface{}) *types.VisualMenuOption {
	if tfMap == nil {
		return nil
	}

	options := &types.VisualMenuOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = types.DashboardBehavior(v)
	}

	return options
}

func FlattenDashboardDefinition(apiObject *types.DashboardVersionDefinition) []interface{} {
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
	if apiObject.DataSetIdentifierDeclarations != nil {
		tfMap["data_set_identifiers_declarations"] = flattenDataSetIdentifierDeclarations(apiObject.DataSetIdentifierDeclarations)
	}
	if apiObject.FilterGroups != nil {
		tfMap["filter_groups"] = flattenFilterGroups(apiObject.FilterGroups)
	}
	if apiObject.ParameterDeclarations != nil {
		tfMap["parameter_declarations"] = flattenParameterDeclarations(apiObject.ParameterDeclarations)
	}
	if apiObject.Sheets != nil {
		tfMap["sheets"] = flattenSheetDefinitions(apiObject.Sheets)
	}

	return []interface{}{tfMap}
}

func FlattenDashboardPublishOptions(apiObject *types.DashboardPublishOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AdHocFilteringOption != nil {
		tfMap["ad_hoc_filtering_option"] = flattenAdHocFilteringOption(apiObject.AdHocFilteringOption)
	}
	if apiObject.DataPointDrillUpDownOption != nil {
		tfMap["data_point_drill_up_down_option"] = flattenDataPointDrillUpDownOption(apiObject.DataPointDrillUpDownOption)
	}
	if apiObject.DataPointMenuLabelOption != nil {
		tfMap["data_point_menu_label_option"] = flattenDataPointMenuLabelOption(apiObject.DataPointMenuLabelOption)
	}
	if apiObject.DataPointTooltipOption != nil {
		tfMap["data_point_tooltip_option"] = flattenDataPointTooltipOption(apiObject.DataPointTooltipOption)
	}
	if apiObject.ExportToCSVOption != nil {
		tfMap["export_to_csv_option"] = flattenExportToCSVOption(apiObject.ExportToCSVOption)
	}
	if apiObject.ExportWithHiddenFieldsOption != nil {
		tfMap["export_with_hidden_fields_option"] = flattenExportWithHiddenFieldsOption(apiObject.ExportWithHiddenFieldsOption)
	}
	if apiObject.SheetControlsOption != nil {
		tfMap["sheet_controls_option"] = flattenSheetControlsOption(apiObject.SheetControlsOption)
	}
	if apiObject.SheetLayoutElementMaximizationOption != nil {
		tfMap["sheet_layout_element_maximization_option"] = flattenSheetLayoutElementMaximizationOption(apiObject.SheetLayoutElementMaximizationOption)
	}
	if apiObject.VisualAxisSortOption != nil {
		tfMap["visual_axis_sort_option"] = flattenVisualAxisSortOption(apiObject.VisualAxisSortOption)
	}
	if apiObject.VisualMenuOption != nil {
		tfMap["visual_menu_option"] = flattenVisualMenuOption(apiObject.VisualMenuOption)
	}

	return []interface{}{tfMap}
}

func flattenAdHocFilteringOption(apiObject *types.AdHocFilteringOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["availability_status"] = types.DashboardBehavior(apiObject.AvailabilityStatus)

	return []interface{}{tfMap}
}

func flattenDataPointDrillUpDownOption(apiObject *types.DataPointDrillUpDownOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["availability_status"] = types.DashboardBehavior(apiObject.AvailabilityStatus)

	return []interface{}{tfMap}
}

func flattenDataPointMenuLabelOption(apiObject *types.DataPointMenuLabelOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["availability_status"] = types.DashboardBehavior(apiObject.AvailabilityStatus)

	return []interface{}{tfMap}
}

func flattenDataPointTooltipOption(apiObject *types.DataPointTooltipOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["availability_status"] = types.DashboardBehavior(apiObject.AvailabilityStatus)

	return []interface{}{tfMap}
}

func flattenExportToCSVOption(apiObject *types.ExportToCSVOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["availability_status"] = types.DashboardBehavior(apiObject.AvailabilityStatus)

	return []interface{}{tfMap}
}

func flattenExportWithHiddenFieldsOption(apiObject *types.ExportWithHiddenFieldsOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["availability_status"] = types.DashboardBehavior(apiObject.AvailabilityStatus)

	return []interface{}{tfMap}
}

func flattenSheetControlsOption(apiObject *types.SheetControlsOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["visibility_state"] = types.DashboardUIState(apiObject.VisibilityState)

	return []interface{}{tfMap}
}

func flattenSheetLayoutElementMaximizationOption(apiObject *types.SheetLayoutElementMaximizationOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["availability_status"] = types.DashboardBehavior(apiObject.AvailabilityStatus)

	return []interface{}{tfMap}
}

func flattenVisualAxisSortOption(apiObject *types.VisualAxisSortOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["availability_status"] = types.DashboardBehavior(apiObject.AvailabilityStatus)

	return []interface{}{tfMap}
}

func flattenVisualMenuOption(apiObject *types.VisualMenuOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["availability_status"] = types.DashboardBehavior(apiObject.AvailabilityStatus)

	return []interface{}{tfMap}
}
