// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
								Type:         schema.TypeString,
								Optional:     true,
								Default:      quicksight.StatusEnabled,
								ValidateFunc: validation.StringInSlice(quicksight.Status_Values(), false),
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
								Type:         schema.TypeString,
								Optional:     true,
								Default:      quicksight.StatusEnabled,
								ValidateFunc: validation.StringInSlice(quicksight.Status_Values(), false),
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
								Type:         schema.TypeString,
								Optional:     true,
								Default:      quicksight.StatusEnabled,
								ValidateFunc: validation.StringInSlice(quicksight.Status_Values(), false),
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
								Type:         schema.TypeString,
								Optional:     true,
								Default:      quicksight.StatusEnabled,
								ValidateFunc: validation.StringInSlice(quicksight.Status_Values(), false),
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
								Type:         schema.TypeString,
								Optional:     true,
								Default:      quicksight.StatusEnabled,
								ValidateFunc: validation.StringInSlice(quicksight.Status_Values(), false),
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
								Type:         schema.TypeString,
								Optional:     true,
								Default:      quicksight.StatusDisabled,
								ValidateFunc: validation.StringInSlice(quicksight.Status_Values(), false),
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
								Type:         schema.TypeString,
								Optional:     true,
								Default:      quicksight.DashboardUIStateCollapsed,
								ValidateFunc: validation.StringInSlice(quicksight.DashboardUIState_Values(), false),
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
								Type:         schema.TypeString,
								Optional:     true,
								Default:      quicksight.StatusEnabled,
								ValidateFunc: validation.StringInSlice(quicksight.Status_Values(), false),
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
								Type:         schema.TypeString,
								Optional:     true,
								Default:      quicksight.StatusEnabled,
								ValidateFunc: validation.StringInSlice(quicksight.Status_Values(), false),
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
								Type:         schema.TypeString,
								Optional:     true,
								Default:      quicksight.StatusEnabled,
								ValidateFunc: validation.StringInSlice(quicksight.Status_Values(), false),
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
							names.AttrARN: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: verify.ValidARN,
							},
							"data_set_references": dataSetReferencesSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataSetReference.html
						},
					},
				},
			},
		},
	}
}

func ExpandDashboardSourceEntity(tfList []interface{}) *quicksight.DashboardSourceEntity {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	sourceEntity := &quicksight.DashboardSourceEntity{}

	if v, ok := tfMap["source_template"].([]interface{}); ok && len(v) > 0 {
		sourceEntity.SourceTemplate = expandDashboardSourceTemplate(v[0].(map[string]interface{}))
	}

	return sourceEntity
}

func expandDashboardSourceTemplate(tfMap map[string]interface{}) *quicksight.DashboardSourceTemplate {
	if tfMap == nil {
		return nil
	}

	sourceTemplate := &quicksight.DashboardSourceTemplate{}
	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		sourceTemplate.Arn = aws.String(v)
	}
	if v, ok := tfMap["data_set_references"].([]interface{}); ok && len(v) > 0 {
		sourceTemplate.DataSetReferences = expandDataSetReferences(v)
	}

	return sourceTemplate
}

func ExpandDashboardDefinition(tfList []interface{}) *quicksight.DashboardVersionDefinition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	definition := &quicksight.DashboardVersionDefinition{}

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

func ExpandDashboardPublishOptions(tfList []interface{}) *quicksight.DashboardPublishOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.DashboardPublishOptions{}

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

func expandAdHocFilteringOption(tfMap map[string]interface{}) *quicksight.AdHocFilteringOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.AdHocFilteringOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = aws.String(v)
	}

	return options
}

func expandDataPointDrillUpDownOption(tfMap map[string]interface{}) *quicksight.DataPointDrillUpDownOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.DataPointDrillUpDownOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = aws.String(v)
	}

	return options
}

func expandDataPointMenuLabelOption(tfMap map[string]interface{}) *quicksight.DataPointMenuLabelOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.DataPointMenuLabelOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = aws.String(v)
	}

	return options
}

func expandDataPointTooltipOption(tfMap map[string]interface{}) *quicksight.DataPointTooltipOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.DataPointTooltipOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = aws.String(v)
	}

	return options
}

func expandExportToCSVOption(tfMap map[string]interface{}) *quicksight.ExportToCSVOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.ExportToCSVOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = aws.String(v)
	}

	return options
}

func expandExportWithHiddenFieldsOption(tfMap map[string]interface{}) *quicksight.ExportWithHiddenFieldsOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.ExportWithHiddenFieldsOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = aws.String(v)
	}

	return options
}

func expandSheetLayoutElementMaximizationOption(tfMap map[string]interface{}) *quicksight.SheetLayoutElementMaximizationOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.SheetLayoutElementMaximizationOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = aws.String(v)
	}

	return options
}

func expandSheetControlsOption(tfMap map[string]interface{}) *quicksight.SheetControlsOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.SheetControlsOption{}
	if v, ok := tfMap["visibility_state"].(string); ok && v != "" {
		options.VisibilityState = aws.String(v)
	}

	return options
}

func expandVisualAxisSortOption(tfMap map[string]interface{}) *quicksight.VisualAxisSortOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.VisualAxisSortOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = aws.String(v)
	}

	return options
}

func expandVisualMenuOption(tfMap map[string]interface{}) *quicksight.VisualMenuOption {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.VisualMenuOption{}
	if v, ok := tfMap["availability_status"].(string); ok && v != "" {
		options.AvailabilityStatus = aws.String(v)
	}

	return options
}

func FlattenDashboardDefinition(apiObject *quicksight.DashboardVersionDefinition) []interface{} {
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

func FlattenDashboardPublishOptions(apiObject *quicksight.DashboardPublishOptions) []interface{} {
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

func flattenAdHocFilteringOption(apiObject *quicksight.AdHocFilteringOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AvailabilityStatus != nil {
		tfMap["availability_status"] = aws.StringValue(apiObject.AvailabilityStatus)
	}

	return []interface{}{tfMap}
}

func flattenDataPointDrillUpDownOption(apiObject *quicksight.DataPointDrillUpDownOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AvailabilityStatus != nil {
		tfMap["availability_status"] = aws.StringValue(apiObject.AvailabilityStatus)
	}

	return []interface{}{tfMap}
}

func flattenDataPointMenuLabelOption(apiObject *quicksight.DataPointMenuLabelOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AvailabilityStatus != nil {
		tfMap["availability_status"] = aws.StringValue(apiObject.AvailabilityStatus)
	}

	return []interface{}{tfMap}
}

func flattenDataPointTooltipOption(apiObject *quicksight.DataPointTooltipOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AvailabilityStatus != nil {
		tfMap["availability_status"] = aws.StringValue(apiObject.AvailabilityStatus)
	}

	return []interface{}{tfMap}
}

func flattenExportToCSVOption(apiObject *quicksight.ExportToCSVOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AvailabilityStatus != nil {
		tfMap["availability_status"] = aws.StringValue(apiObject.AvailabilityStatus)
	}

	return []interface{}{tfMap}
}

func flattenExportWithHiddenFieldsOption(apiObject *quicksight.ExportWithHiddenFieldsOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AvailabilityStatus != nil {
		tfMap["availability_status"] = aws.StringValue(apiObject.AvailabilityStatus)
	}

	return []interface{}{tfMap}
}

func flattenSheetControlsOption(apiObject *quicksight.SheetControlsOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.VisibilityState != nil {
		tfMap["visibility_state"] = aws.StringValue(apiObject.VisibilityState)
	}

	return []interface{}{tfMap}
}

func flattenSheetLayoutElementMaximizationOption(apiObject *quicksight.SheetLayoutElementMaximizationOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AvailabilityStatus != nil {
		tfMap["availability_status"] = aws.StringValue(apiObject.AvailabilityStatus)
	}

	return []interface{}{tfMap}
}

func flattenVisualAxisSortOption(apiObject *quicksight.VisualAxisSortOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AvailabilityStatus != nil {
		tfMap["availability_status"] = aws.StringValue(apiObject.AvailabilityStatus)
	}

	return []interface{}{tfMap}
}

func flattenVisualMenuOption(apiObject *quicksight.VisualMenuOption) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AvailabilityStatus != nil {
		tfMap["availability_status"] = aws.StringValue(apiObject.AvailabilityStatus)
	}

	return []interface{}{tfMap}
}
