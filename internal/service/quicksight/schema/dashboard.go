// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/names"
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
							names.AttrRole:         stringEnumSchema[awstypes.ColumnRole](attrOptional),
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
							"cross_dataset":       stringEnumSchema[awstypes.CrossDatasetTypes](attrRequired),
							"filter_group_id":     idSchema(),
							"filters":             filtersSchema(),                  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Filter.html
							"scope_configuration": filterScopeConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterScopeConfiguration.html
							names.AttrStatus:      stringEnumSchema[awstypes.Status](attrOptional),
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
							"sheet_id":              idSchema(),
							names.AttrContentType:   stringEnumSchema[awstypes.SheetContentType](attrOptionalComputed),
							names.AttrDescription:   stringLenBetweenSchema(attrOptional, 1, 1024),
							"filter_controls":       filterControlsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterControl.html
							"layouts":               layoutSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Layout.html
							names.AttrName:          stringLenBetweenSchema(attrOptional, 1, 2048),
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
										names.AttrContent:   stringLenBetweenSchema(attrOptional, 1, 150000),
									},
								},
							},
							"title":   stringLenBetweenSchema(attrOptional, 1, 1024),
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
								Default:          awstypes.StatusEnabled,
								ValidateDiagFunc: enum.Validate[awstypes.Status](),
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
								Default:          awstypes.StatusEnabled,
								ValidateDiagFunc: enum.Validate[awstypes.Status](),
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
								Default:          awstypes.StatusEnabled,
								ValidateDiagFunc: enum.Validate[awstypes.Status](),
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
								Default:          awstypes.StatusEnabled,
								ValidateDiagFunc: enum.Validate[awstypes.Status](),
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
								Default:          awstypes.StatusEnabled,
								ValidateDiagFunc: enum.Validate[awstypes.Status](),
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
								Default:          awstypes.StatusDisabled,
								ValidateDiagFunc: enum.Validate[awstypes.Status](),
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
								Default:          awstypes.DashboardUIStateCollapsed,
								ValidateDiagFunc: enum.Validate[awstypes.DashboardUIState](),
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
								Default:          awstypes.StatusEnabled,
								ValidateDiagFunc: enum.Validate[awstypes.Status](),
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
								Default:          awstypes.StatusEnabled,
								ValidateDiagFunc: enum.Validate[awstypes.Status](),
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
								Default:          awstypes.StatusEnabled,
								ValidateDiagFunc: enum.Validate[awstypes.Status](),
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
							names.AttrARN:         arnStringSchema(attrRequired),
							"data_set_references": dataSetReferencesSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataSetReference.html
						},
					},
				},
			},
		},
	}
}

func ExpandDashboardSourceEntity(tfList []interface{}) *awstypes.DashboardSourceEntity {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DashboardSourceEntity{}

	if v, ok := tfMap["source_template"].([]interface{}); ok && len(v) > 0 {
		apiObject.SourceTemplate = expandDashboardSourceTemplate(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandDashboardSourceTemplate(tfMap map[string]interface{}) *awstypes.DashboardSourceTemplate {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DashboardSourceTemplate{}

	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		apiObject.Arn = aws.String(v)
	}
	if v, ok := tfMap["data_set_references"].([]interface{}); ok && len(v) > 0 {
		apiObject.DataSetReferences = expandDataSetReferences(v)
	}

	return apiObject
}

func ExpandDashboardDefinition(tfList []interface{}) *awstypes.DashboardVersionDefinition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DashboardVersionDefinition{}

	if v, ok := tfMap["analysis_defaults"].([]interface{}); ok && len(v) > 0 {
		apiObject.AnalysisDefaults = expandAnalysisDefaults(v)
	}
	if v, ok := tfMap["calculated_fields"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CalculatedFields = expandCalculatedFields(v.List())
	}
	if v, ok := tfMap["column_configurations"].([]interface{}); ok && len(v) > 0 {
		apiObject.ColumnConfigurations = expandColumnConfigurations(v)
	}
	if v, ok := tfMap["data_set_identifiers_declarations"].([]interface{}); ok && len(v) > 0 {
		apiObject.DataSetIdentifierDeclarations = expandDataSetIdentifierDeclarations(v)
	}
	if v, ok := tfMap["filter_groups"].([]interface{}); ok && len(v) > 0 {
		apiObject.FilterGroups = expandFilterGroups(v)
	}
	if v, ok := tfMap["parameter_declarations"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ParameterDeclarations = expandParameterDeclarations(v.List())
	}
	if v, ok := tfMap["sheets"].([]interface{}); ok && len(v) > 0 {
		apiObject.Sheets = expandSheetDefinitions(v)
	}

	return apiObject
}

func ExpandDashboardPublishOptions(tfList []interface{}) *awstypes.DashboardPublishOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DashboardPublishOptions{}

	if v, ok := tfMap["ad_hoc_filtering_option"].([]interface{}); ok && len(v) > 0 {
		apiObject.AdHocFilteringOption = expandAdHocFilteringOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["data_point_drill_up_down_option"].([]interface{}); ok && len(v) > 0 {
		apiObject.DataPointDrillUpDownOption = expandDataPointDrillUpDownOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["data_point_menu_label_option"].([]interface{}); ok && len(v) > 0 {
		apiObject.DataPointMenuLabelOption = expandDataPointMenuLabelOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["data_point_tooltip_option"].([]interface{}); ok && len(v) > 0 {
		apiObject.DataPointTooltipOption = expandDataPointTooltipOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["export_to_csv_option"].([]interface{}); ok && len(v) > 0 {
		apiObject.ExportToCSVOption = expandExportToCSVOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["export_with_hidden_fields_option"].([]interface{}); ok && len(v) > 0 {
		apiObject.ExportWithHiddenFieldsOption = expandExportWithHiddenFieldsOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["sheet_controls_option"].([]interface{}); ok && len(v) > 0 {
		apiObject.SheetControlsOption = expandSheetControlsOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["sheet_layout_element_maximization_option"].([]interface{}); ok && len(v) > 0 {
		apiObject.SheetLayoutElementMaximizationOption = expandSheetLayoutElementMaximizationOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["visual_axis_sort_option"].([]interface{}); ok && len(v) > 0 {
		apiObject.VisualAxisSortOption = expandVisualAxisSortOption(v[0].(map[string]interface{}))
	}
	if v, ok := tfMap["visual_menu_option"].([]interface{}); ok && len(v) > 0 {
		apiObject.VisualMenuOption = expandVisualMenuOption(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandAdHocFilteringOption(tfMap map[string]interface{}) *awstypes.AdHocFilteringOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AdHocFilteringOption{}

	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		apiObject.AvailabilityStatus = awstypes.DashboardBehavior(v)
	}

	return apiObject
}

func expandDataPointDrillUpDownOption(tfMap map[string]interface{}) *awstypes.DataPointDrillUpDownOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DataPointDrillUpDownOption{}

	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		apiObject.AvailabilityStatus = awstypes.DashboardBehavior(v)
	}

	return apiObject
}

func expandDataPointMenuLabelOption(tfMap map[string]interface{}) *awstypes.DataPointMenuLabelOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DataPointMenuLabelOption{}

	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		apiObject.AvailabilityStatus = awstypes.DashboardBehavior(v)
	}

	return apiObject
}

func expandDataPointTooltipOption(tfMap map[string]interface{}) *awstypes.DataPointTooltipOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DataPointTooltipOption{}

	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		apiObject.AvailabilityStatus = awstypes.DashboardBehavior(v)
	}

	return apiObject
}

func expandExportToCSVOption(tfMap map[string]interface{}) *awstypes.ExportToCSVOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ExportToCSVOption{}

	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		apiObject.AvailabilityStatus = awstypes.DashboardBehavior(v)
	}

	return apiObject
}

func expandExportWithHiddenFieldsOption(tfMap map[string]interface{}) *awstypes.ExportWithHiddenFieldsOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ExportWithHiddenFieldsOption{}

	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		apiObject.AvailabilityStatus = awstypes.DashboardBehavior(v)
	}

	return apiObject
}

func expandSheetLayoutElementMaximizationOption(tfMap map[string]interface{}) *awstypes.SheetLayoutElementMaximizationOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SheetLayoutElementMaximizationOption{}

	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		apiObject.AvailabilityStatus = awstypes.DashboardBehavior(v)
	}

	return apiObject
}

func expandSheetControlsOption(tfMap map[string]interface{}) *awstypes.SheetControlsOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SheetControlsOption{}

	if v, ok := tfMap["visibility_state"].(string); ok && v != "" {
		apiObject.VisibilityState = awstypes.DashboardUIState(v)
	}

	return apiObject
}

func expandVisualAxisSortOption(tfMap map[string]interface{}) *awstypes.VisualAxisSortOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.VisualAxisSortOption{}

	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		apiObject.AvailabilityStatus = awstypes.DashboardBehavior(v)
	}

	return apiObject
}

func expandVisualMenuOption(tfMap map[string]interface{}) *awstypes.VisualMenuOption {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.VisualMenuOption{}

	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		apiObject.AvailabilityStatus = awstypes.DashboardBehavior(v)
	}

	return apiObject
}

func FlattenDashboardDefinition(apiObject *awstypes.DashboardVersionDefinition) []interface{} {
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

func FlattenDashboardPublishOptions(apiObject *awstypes.DashboardPublishOptions) []interface{} {
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

func flattenAdHocFilteringOption(apiObject *awstypes.AdHocFilteringOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"availability_status": apiObject.AvailabilityStatus,
	}

	return []interface{}{tfMap}
}

func flattenDataPointDrillUpDownOption(apiObject *awstypes.DataPointDrillUpDownOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"availability_status": apiObject.AvailabilityStatus,
	}

	return []interface{}{tfMap}
}

func flattenDataPointMenuLabelOption(apiObject *awstypes.DataPointMenuLabelOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"availability_status": apiObject.AvailabilityStatus,
	}

	return []interface{}{tfMap}
}

func flattenDataPointTooltipOption(apiObject *awstypes.DataPointTooltipOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"availability_status": apiObject.AvailabilityStatus,
	}

	return []interface{}{tfMap}
}

func flattenExportToCSVOption(apiObject *awstypes.ExportToCSVOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"availability_status": apiObject.AvailabilityStatus,
	}

	return []interface{}{tfMap}
}

func flattenExportWithHiddenFieldsOption(apiObject *awstypes.ExportWithHiddenFieldsOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"availability_status": apiObject.AvailabilityStatus,
	}

	return []interface{}{tfMap}
}

func flattenSheetControlsOption(apiObject *awstypes.SheetControlsOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"visibility_state": apiObject.VisibilityState,
	}

	return []interface{}{tfMap}
}

func flattenSheetLayoutElementMaximizationOption(apiObject *awstypes.SheetLayoutElementMaximizationOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"availability_status": apiObject.AvailabilityStatus,
	}

	return []interface{}{tfMap}
}

func flattenVisualAxisSortOption(apiObject *awstypes.VisualAxisSortOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"availability_status": apiObject.AvailabilityStatus,
	}

	return []interface{}{tfMap}
}

func flattenVisualMenuOption(apiObject *awstypes.VisualMenuOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"availability_status": apiObject.AvailabilityStatus,
	}

	return []interface{}{tfMap}
}
