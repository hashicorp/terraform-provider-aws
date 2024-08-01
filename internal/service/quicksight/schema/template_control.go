// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func filterControlsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterControl.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 200,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"date_time_picker": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterDateTimePickerControl.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"filter_control_id": idSchema(),
							"source_filter_id":  idSchema(),
							"title":             stringSchema(true, validation.StringLenBetween(1, 2048)),
							"display_options":   dateTimePickerControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimePickerControlDisplayOptions.html
							names.AttrType:      stringSchema(false, validation.StringInSlice(quicksight.SheetControlDateTimePickerType_Values(), false)),
						},
					},
				},
				"dropdown": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterDropDownControl.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"filter_control_id":               idSchema(),
							"source_filter_id":                idSchema(),
							"title":                           stringSchema(true, validation.StringLenBetween(1, 2048)),
							"cascading_control_configuration": cascadingControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CascadingControlConfiguration.html
							"display_options":                 dropDownControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DropDownControlDisplayOptions.html
							"selectable_values":               filterSelectableValuesSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterSelectableValues.html
							names.AttrType:                    stringSchema(false, validation.StringInSlice(quicksight.SheetControlListType_Values(), false)),
						},
					},
				},
				"list": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterListControl.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"filter_control_id":               idSchema(),
							"source_filter_id":                idSchema(),
							"title":                           stringSchema(true, validation.StringLenBetween(1, 2048)),
							"cascading_control_configuration": cascadingControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CascadingControlConfiguration.html
							"display_options":                 listControlDisplayOptionsSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ListControlDisplayOptions.html
							"selectable_values":               filterSelectableValuesSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterSelectableValues.html
							names.AttrType:                    stringSchema(false, validation.StringInSlice(quicksight.SheetControlListType_Values(), false)),
						},
					},
				},
				"relative_date_time": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterRelativeDateTimeControl.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"filter_control_id": idSchema(),
							"source_filter_id":  idSchema(),
							"title":             stringSchema(true, validation.StringLenBetween(1, 2048)),
							"display_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RelativeDateTimeControlDisplayOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"date_time_format": stringSchema(false, validation.StringLenBetween(1, 128)),
										"title_options":    labelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LabelOptions.html
									},
								},
							},
						},
					},
				},
				"slider": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterSliderControl.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"filter_control_id": idSchema(),
							"source_filter_id":  idSchema(),
							"title":             stringSchema(true, validation.StringLenBetween(1, 2048)),
							"maximum_value": {
								Type:     schema.TypeFloat,
								Required: true,
							},
							"minimum_value": {
								Type:     schema.TypeFloat,
								Required: true,
							},
							"step_size": {
								Type:     schema.TypeFloat,
								Required: true,
							},
							"display_options": sliderControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SliderControlDisplayOptions.html
							names.AttrType:    stringSchema(false, validation.StringInSlice(quicksight.SheetControlSliderType_Values(), false)),
						},
					},
				},
				"text_area": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterTextAreaControl.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"filter_control_id": idSchema(),
							"source_filter_id":  idSchema(),
							"title":             stringSchema(true, validation.StringLenBetween(1, 2048)),
							"delimiter":         stringSchema(false, validation.StringLenBetween(1, 2048)),
							"display_options":   textAreaControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextAreaControlDisplayOptions.html
						},
					},
				},
				"text_field": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterTextFieldControl.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"filter_control_id": idSchema(),
							"source_filter_id":  idSchema(),
							"title":             stringSchema(true, validation.StringLenBetween(1, 2048)),
							"display_options":   textFieldControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextFieldControlDisplayOptions.html
						},
					},
				},
			},
		},
	}
}

func textFieldControlDisplayOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextFieldControlDisplayOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"placeholder_options": placeholderOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextControlPlaceholderOptions.html
				"title_options":       labelOptionsSchema(),       // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LabelOptions.html
			},
		},
	}
}

func textAreaControlDisplayOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextAreaControlDisplayOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"placeholder_options": placeholderOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextControlPlaceholderOptions.html
				"title_options":       labelOptionsSchema(),       // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LabelOptions.html
			},
		},
	}
}

func sliderControlDisplayOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SliderControlDisplayOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"title_options": labelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LabelOptions.html
			},
		},
	}
}

func dateTimePickerControlDisplayOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimePickerControlDisplayOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"date_time_format": stringSchema(false, validation.StringLenBetween(1, 128)),
				"title_options":    labelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LabelOptions.html
			},
		},
	}
}

func listControlDisplayOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ListControlDisplayOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"search_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ListControlSearchOptions.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
						},
					},
				},
				"select_all_options": selectAllOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ListControlSelectAllOptions.html
				"title_options":      labelOptionsSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LabelOptions.html
			},
		},
	}
}

func cascadingControlConfigurationSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CascadingControlConfiguration.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"source_controls": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CascadingControlSource.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 200,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"column_to_match": columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
							"source_sheet_control_id": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

func selectAllOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ListControlSelectAllOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
			},
		},
	}
}

func dropDownControlDisplayOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DropDownControlDisplayOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"select_all_options": selectAllOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ListControlSelectAllOptions.html
				"title_options":      labelOptionsSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LabelOptions.html
			},
		},
	}
}

func placeholderOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextControlPlaceholderOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
			},
		},
	}
}

func expandFilterControl(tfMap map[string]interface{}) *quicksight.FilterControl {
	if tfMap == nil {
		return nil
	}

	control := &quicksight.FilterControl{}

	if v, ok := tfMap["date_time_picker"].([]interface{}); ok && len(v) > 0 {
		control.DateTimePicker = expandFilterDateTimePickerControl(v)
	}
	if v, ok := tfMap["dropdown"].([]interface{}); ok && len(v) > 0 {
		control.Dropdown = expandFilterDropDownControl(v)
	}
	if v, ok := tfMap["list"].([]interface{}); ok && len(v) > 0 {
		control.List = expandFilterListControl(v)
	}
	if v, ok := tfMap["relative_date_time"].([]interface{}); ok && len(v) > 0 {
		control.RelativeDateTime = expandFilterRelativeDateTimeControl(v)
	}
	if v, ok := tfMap["slider"].([]interface{}); ok && len(v) > 0 {
		control.Slider = expandFilterSliderControl(v)
	}
	if v, ok := tfMap["text_area"].([]interface{}); ok && len(v) > 0 {
		control.TextArea = expandFilterTextAreaControl(v)
	}
	if v, ok := tfMap["text_field"].([]interface{}); ok && len(v) > 0 {
		control.TextField = expandFilterTextFieldControl(v)
	}
	return control
}

func expandFilterDateTimePickerControl(tfList []interface{}) *quicksight.FilterDateTimePickerControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	control := &quicksight.FilterDateTimePickerControl{}

	if v, ok := tfMap["filter_control_id"].(string); ok && v != "" {
		control.FilterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_filter_id"].(string); ok && v != "" {
		control.SourceFilterId = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		control.Title = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		control.Type = aws.String(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		control.DisplayOptions = expandDateTimePickerControlDisplayOptions(v)
	}

	return control
}

func expandDateTimePickerControlDisplayOptions(tfList []interface{}) *quicksight.DateTimePickerControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.DateTimePickerControlDisplayOptions{}

	if v, ok := tfMap["date_time_format"].(string); ok && v != "" {
		options.DateTimeFormat = aws.String(v)
	}
	if v, ok := tfMap["title_options"].([]interface{}); ok && len(v) > 0 {
		options.TitleOptions = expandLabelOptions(v)
	}

	return options
}

func expandFilterDropDownControl(tfList []interface{}) *quicksight.FilterDropDownControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	control := &quicksight.FilterDropDownControl{}

	if v, ok := tfMap["filter_control_id"].(string); ok && v != "" {
		control.FilterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_filter_id"].(string); ok && v != "" {
		control.SourceFilterId = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		control.Title = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		control.Type = aws.String(v)
	}
	if v, ok := tfMap["cascading_control_configuration"].([]interface{}); ok && len(v) > 0 {
		control.CascadingControlConfiguration = expandCascadingControlConfiguration(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		control.DisplayOptions = expandDropDownControlDisplayOptions(v)
	}
	if v, ok := tfMap["selectable_values"].([]interface{}); ok && len(v) > 0 {
		control.SelectableValues = expandFilterSelectableValues(v)
	}

	return control
}

func expandCascadingControlConfiguration(tfList []interface{}) *quicksight.CascadingControlConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.CascadingControlConfiguration{}

	if v, ok := tfMap["source_controls"].([]interface{}); ok && len(v) > 0 {
		config.SourceControls = expandCascadingControlSources(v)
	}

	return config
}

func expandCascadingControlSources(tfList []interface{}) []*quicksight.CascadingControlSource {
	if len(tfList) == 0 {
		return nil
	}

	var sources []*quicksight.CascadingControlSource
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		source := expandCascadingControlSource(tfMap)
		if source == nil {
			continue
		}

		sources = append(sources, source)
	}

	return sources
}

func expandCascadingControlSource(tfMap map[string]interface{}) *quicksight.CascadingControlSource {
	if tfMap == nil {
		return nil
	}

	source := &quicksight.CascadingControlSource{}

	if v, ok := tfMap["source_sheet_control_id"].(string); ok && v != "" {
		source.SourceSheetControlId = aws.String(v)
	}
	if v, ok := tfMap["column_to_match"].([]interface{}); ok && len(v) > 0 {
		source.ColumnToMatch = expandColumnIdentifier(v)
	}

	return source
}

func expandDropDownControlDisplayOptions(tfList []interface{}) *quicksight.DropDownControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.DropDownControlDisplayOptions{}

	if v, ok := tfMap["select_all_options"].([]interface{}); ok && len(v) > 0 {
		options.SelectAllOptions = expandListControlSelectAllOptions(v)
	}
	if v, ok := tfMap["title_options"].([]interface{}); ok && len(v) > 0 {
		options.TitleOptions = expandLabelOptions(v)
	}

	return options
}

func expandListControlSelectAllOptions(tfList []interface{}) *quicksight.ListControlSelectAllOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.ListControlSelectAllOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = aws.String(v)
	}

	return options
}

func expandFilterSelectableValues(tfList []interface{}) *quicksight.FilterSelectableValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	values := &quicksight.FilterSelectableValues{}

	if v, ok := tfMap[names.AttrValues].([]interface{}); ok {
		values.Values = flex.ExpandStringList(v)
	}

	return values
}

func expandFilterListControl(tfList []interface{}) *quicksight.FilterListControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	control := &quicksight.FilterListControl{}

	if v, ok := tfMap["filter_control_id"].(string); ok && v != "" {
		control.FilterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_filter_id"].(string); ok && v != "" {
		control.SourceFilterId = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		control.Title = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		control.Type = aws.String(v)
	}
	if v, ok := tfMap["cascading_control_configuration"].([]interface{}); ok && len(v) > 0 {
		control.CascadingControlConfiguration = expandCascadingControlConfiguration(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		control.DisplayOptions = expandListControlDisplayOptions(v)
	}
	if v, ok := tfMap["selectable_values"].([]interface{}); ok && len(v) > 0 {
		control.SelectableValues = expandFilterSelectableValues(v)
	}

	return control
}

func expandListControlDisplayOptions(tfList []interface{}) *quicksight.ListControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.ListControlDisplayOptions{}

	if v, ok := tfMap["select_all_options"].([]interface{}); ok && len(v) > 0 {
		options.SelectAllOptions = expandListControlSelectAllOptions(v)
	}
	if v, ok := tfMap["title_options"].([]interface{}); ok && len(v) > 0 {
		options.TitleOptions = expandLabelOptions(v)
	}
	if v, ok := tfMap["search_options"].([]interface{}); ok && len(v) > 0 {
		options.SearchOptions = expandListControlSearchOptions(v)
	}

	return options
}

func expandListControlSearchOptions(tfList []interface{}) *quicksight.ListControlSearchOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.ListControlSearchOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = aws.String(v)
	}

	return options
}

func expandFilterRelativeDateTimeControl(tfList []interface{}) *quicksight.FilterRelativeDateTimeControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	control := &quicksight.FilterRelativeDateTimeControl{}

	if v, ok := tfMap["filter_control_id"].(string); ok && v != "" {
		control.FilterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_filter_id"].(string); ok && v != "" {
		control.SourceFilterId = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		control.Title = aws.String(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		control.DisplayOptions = expandRelativeDateTimeControlDisplayOptions(v)
	}

	return control
}

func expandRelativeDateTimeControlDisplayOptions(tfList []interface{}) *quicksight.RelativeDateTimeControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.RelativeDateTimeControlDisplayOptions{}

	if v, ok := tfMap["date_time_format"].(string); ok && v != "" {
		options.DateTimeFormat = aws.String(v)
	}
	if v, ok := tfMap["title_options"].([]interface{}); ok && len(v) > 0 {
		options.TitleOptions = expandLabelOptions(v)
	}

	return options
}

func expandFilterSliderControl(tfList []interface{}) *quicksight.FilterSliderControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	control := &quicksight.FilterSliderControl{}

	if v, ok := tfMap["filter_control_id"].(string); ok && v != "" {
		control.FilterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_filter_id"].(string); ok && v != "" {
		control.SourceFilterId = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		control.Title = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		control.Type = aws.String(v)
	}
	if v, ok := tfMap["maximum_value"].(float64); ok {
		control.MaximumValue = aws.Float64(v)
	}
	if v, ok := tfMap["minimum_value"].(float64); ok {
		control.MinimumValue = aws.Float64(v)
	}
	if v, ok := tfMap["step_size"].(float64); ok {
		control.StepSize = aws.Float64(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		control.DisplayOptions = expandSliderControlDisplayOptions(v)
	}

	return control
}

func expandSliderControlDisplayOptions(tfList []interface{}) *quicksight.SliderControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.SliderControlDisplayOptions{}

	if v, ok := tfMap["title_options"].([]interface{}); ok && len(v) > 0 {
		options.TitleOptions = expandLabelOptions(v)
	}

	return options
}

func expandFilterTextAreaControl(tfList []interface{}) *quicksight.FilterTextAreaControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	control := &quicksight.FilterTextAreaControl{}

	if v, ok := tfMap["filter_control_id"].(string); ok && v != "" {
		control.FilterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_filter_id"].(string); ok && v != "" {
		control.SourceFilterId = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		control.Title = aws.String(v)
	}
	if v, ok := tfMap["delimiter"].(string); ok && v != "" {
		control.Delimiter = aws.String(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		control.DisplayOptions = expandTextAreaControlDisplayOptions(v)
	}

	return control
}

func expandTextAreaControlDisplayOptions(tfList []interface{}) *quicksight.TextAreaControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TextAreaControlDisplayOptions{}

	if v, ok := tfMap["placeholder_options"].([]interface{}); ok && len(v) > 0 {
		options.PlaceholderOptions = expandTextControlPlaceholderOptions(v)
	}
	if v, ok := tfMap["title_options"].([]interface{}); ok && len(v) > 0 {
		options.TitleOptions = expandLabelOptions(v)
	}

	return options
}

func expandTextControlPlaceholderOptions(tfList []interface{}) *quicksight.TextControlPlaceholderOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TextControlPlaceholderOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		options.Visibility = aws.String(v)
	}

	return options
}

func expandFilterTextFieldControl(tfList []interface{}) *quicksight.FilterTextFieldControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	control := &quicksight.FilterTextFieldControl{}

	if v, ok := tfMap["filter_control_id"].(string); ok && v != "" {
		control.FilterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_filter_id"].(string); ok && v != "" {
		control.SourceFilterId = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		control.Title = aws.String(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		control.DisplayOptions = expandTextFieldControlDisplayOptions(v)
	}

	return control
}

func expandTextFieldControlDisplayOptions(tfList []interface{}) *quicksight.TextFieldControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.TextFieldControlDisplayOptions{}

	if v, ok := tfMap["placeholder_options"].([]interface{}); ok && len(v) > 0 {
		options.PlaceholderOptions = expandTextControlPlaceholderOptions(v)
	}
	if v, ok := tfMap["title_options"].([]interface{}); ok && len(v) > 0 {
		options.TitleOptions = expandLabelOptions(v)
	}

	return options
}

func expandParameterControl(tfMap map[string]interface{}) *quicksight.ParameterControl {
	if tfMap == nil {
		return nil
	}

	control := &quicksight.ParameterControl{}

	if v, ok := tfMap["date_time_picker"].([]interface{}); ok && len(v) > 0 {
		control.DateTimePicker = expandParameterDateTimePickerControl(v)
	}
	if v, ok := tfMap["dropdown"].([]interface{}); ok && len(v) > 0 {
		control.Dropdown = expandParameterDropDownControl(v)
	}
	if v, ok := tfMap["list"].([]interface{}); ok && len(v) > 0 {
		control.List = expandParameterListControl(v)
	}
	if v, ok := tfMap["slider"].([]interface{}); ok && len(v) > 0 {
		control.Slider = expandParameterSliderControl(v)
	}
	if v, ok := tfMap["text_area"].([]interface{}); ok && len(v) > 0 {
		control.TextArea = expandParameterTextAreaControl(v)
	}
	if v, ok := tfMap["text_field"].([]interface{}); ok && len(v) > 0 {
		control.TextField = expandParameterTextFieldControl(v)
	}

	return control
}

func expandParameterDateTimePickerControl(tfList []interface{}) *quicksight.ParameterDateTimePickerControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	control := &quicksight.ParameterDateTimePickerControl{}

	if v, ok := tfMap["parameter_control_id"].(string); ok && v != "" {
		control.ParameterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_parameter_name"].(string); ok && v != "" {
		control.SourceParameterName = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		control.Title = aws.String(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		control.DisplayOptions = expandDateTimePickerControlDisplayOptions(v)
	}

	return control
}

func expandParameterDropDownControl(tfList []interface{}) *quicksight.ParameterDropDownControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	control := &quicksight.ParameterDropDownControl{}

	if v, ok := tfMap["parameter_control_id"].(string); ok && v != "" {
		control.ParameterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_parameter_name"].(string); ok && v != "" {
		control.SourceParameterName = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		control.Title = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		control.Type = aws.String(v)
	}
	if v, ok := tfMap["cascading_control_configuration"].([]interface{}); ok && len(v) > 0 {
		control.CascadingControlConfiguration = expandCascadingControlConfiguration(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		control.DisplayOptions = expandDropDownControlDisplayOptions(v)
	}
	if v, ok := tfMap["selectable_values"].([]interface{}); ok && len(v) > 0 {
		control.SelectableValues = expandParameterSelectableValues(v)
	}

	return control
}

func expandParameterListControl(tfList []interface{}) *quicksight.ParameterListControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	control := &quicksight.ParameterListControl{}

	if v, ok := tfMap["parameter_control_id"].(string); ok && v != "" {
		control.ParameterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_parameter_name"].(string); ok && v != "" {
		control.SourceParameterName = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		control.Title = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		control.Type = aws.String(v)
	}
	if v, ok := tfMap["cascading_control_configuration"].([]interface{}); ok && len(v) > 0 {
		control.CascadingControlConfiguration = expandCascadingControlConfiguration(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		control.DisplayOptions = expandListControlDisplayOptions(v)
	}
	if v, ok := tfMap["selectable_values"].([]interface{}); ok && len(v) > 0 {
		control.SelectableValues = expandParameterSelectableValues(v)
	}

	return control
}

func expandParameterSliderControl(tfList []interface{}) *quicksight.ParameterSliderControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	control := &quicksight.ParameterSliderControl{}

	if v, ok := tfMap["parameter_control_id"].(string); ok && v != "" {
		control.ParameterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_parameter_name"].(string); ok && v != "" {
		control.SourceParameterName = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		control.Title = aws.String(v)
	}
	if v, ok := tfMap["maximum_value"].(float64); ok {
		control.MaximumValue = aws.Float64(v)
	}
	if v, ok := tfMap["minimum_value"].(float64); ok {
		control.MinimumValue = aws.Float64(v)
	}
	if v, ok := tfMap["step_size"].(float64); ok {
		control.StepSize = aws.Float64(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		control.DisplayOptions = expandSliderControlDisplayOptions(v)
	}

	return control
}

func expandParameterTextAreaControl(tfList []interface{}) *quicksight.ParameterTextAreaControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	control := &quicksight.ParameterTextAreaControl{}

	if v, ok := tfMap["parameter_control_id"].(string); ok && v != "" {
		control.ParameterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_parameter_name"].(string); ok && v != "" {
		control.SourceParameterName = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		control.Title = aws.String(v)
	}
	if v, ok := tfMap["delimiter"].(string); ok && v != "" {
		control.Delimiter = aws.String(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		control.DisplayOptions = expandTextAreaControlDisplayOptions(v)
	}

	return control
}

func expandParameterTextFieldControl(tfList []interface{}) *quicksight.ParameterTextFieldControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	control := &quicksight.ParameterTextFieldControl{}

	if v, ok := tfMap["parameter_control_id"].(string); ok && v != "" {
		control.ParameterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_parameter_name"].(string); ok && v != "" {
		control.SourceParameterName = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		control.Title = aws.String(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		control.DisplayOptions = expandTextFieldControlDisplayOptions(v)
	}

	return control
}

func flattenFilterControls(apiObject []*quicksight.FilterControl) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.DateTimePicker != nil {
			tfMap["date_time_picker"] = flattenFilterDateTimePickerControl(config.DateTimePicker)
		}
		if config.Dropdown != nil {
			tfMap["dropdown"] = flattenFilterDropDownControl(config.Dropdown)
		}
		if config.List != nil {
			tfMap["list"] = flattenFilterListControl(config.List)
		}
		if config.RelativeDateTime != nil {
			tfMap["relative_date_time"] = flattenFilterRelativeDateTimeControl(config.RelativeDateTime)
		}
		if config.Slider != nil {
			tfMap["slider"] = flattenFilterSliderControl(config.Slider)
		}
		if config.TextArea != nil {
			tfMap["text_area"] = flattenFilterTextAreaControl(config.TextArea)
		}
		if config.TextField != nil {
			tfMap["text_field"] = flattenFilterTextFieldControl(config.TextField)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenFilterDateTimePickerControl(apiObject *quicksight.FilterDateTimePickerControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"filter_control_id": aws.StringValue(apiObject.FilterControlId),
		"source_filter_id":  aws.StringValue(apiObject.SourceFilterId),
		"title":             aws.StringValue(apiObject.Title),
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenDateTimePickerControlDisplayOptions(apiObject.DisplayOptions)
	}
	if apiObject.Type != nil {
		tfMap[names.AttrType] = aws.StringValue(apiObject.Type)
	}

	return []interface{}{tfMap}
}

func flattenDateTimePickerControlDisplayOptions(apiObject *quicksight.DateTimePickerControlDisplayOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DateTimeFormat != nil {
		tfMap["date_time_format"] = aws.StringValue(apiObject.DateTimeFormat)
	}
	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []interface{}{tfMap}
}

func flattenLabelOptions(apiObject *quicksight.LabelOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CustomLabel != nil {
		tfMap["custom_label"] = aws.StringValue(apiObject.CustomLabel)
	}
	if apiObject.FontConfiguration != nil {
		tfMap["font_configuration"] = flattenFontConfiguration(apiObject.FontConfiguration)
	}
	if apiObject.Visibility != nil {
		tfMap["visibility"] = aws.StringValue(apiObject.Visibility)
	}

	return []interface{}{tfMap}
}

func flattenFontConfiguration(apiObject *quicksight.FontConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FontColor != nil {
		tfMap["font_color"] = aws.StringValue(apiObject.FontColor)
	}
	if apiObject.FontDecoration != nil {
		tfMap["font_decoration"] = aws.StringValue(apiObject.FontDecoration)
	}
	if apiObject.FontSize != nil {
		tfMap["font_size"] = flattenFontSize(apiObject.FontSize)
	}
	if apiObject.FontStyle != nil {
		tfMap["font_style"] = aws.StringValue(apiObject.FontStyle)
	}
	if apiObject.FontWeight != nil {
		tfMap["font_weight"] = flattenFontWeight(apiObject.FontWeight)
	}

	return []interface{}{tfMap}
}

func flattenFontSize(apiObject *quicksight.FontSize) []interface{} {
	if apiObject == nil || apiObject.Relative == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Relative != nil {
		tfMap["relative"] = aws.StringValue(apiObject.Relative)
	}

	return []interface{}{tfMap}
}

func flattenFontWeight(apiObject *quicksight.FontWeight) []interface{} {
	if apiObject == nil || apiObject.Name == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.StringValue(apiObject.Name)
	}

	return []interface{}{tfMap}
}

func flattenFilterDropDownControl(apiObject *quicksight.FilterDropDownControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"filter_control_id": aws.StringValue(apiObject.FilterControlId),
		"source_filter_id":  aws.StringValue(apiObject.SourceFilterId),
		"title":             aws.StringValue(apiObject.Title),
	}
	if apiObject.CascadingControlConfiguration != nil {
		tfMap["cascading_control_configuration"] = flattenCascadingControlConfiguration(apiObject.CascadingControlConfiguration)
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenDropDownControlDisplayOptions(apiObject.DisplayOptions)
	}
	if apiObject.SelectableValues != nil {
		tfMap["selectable_values"] = flattenFilterSelectableValues(apiObject.SelectableValues)
	}
	if apiObject.Type != nil {
		tfMap[names.AttrType] = aws.StringValue(apiObject.Type)
	}

	return []interface{}{tfMap}
}

func flattenCascadingControlConfiguration(apiObject *quicksight.CascadingControlConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.SourceControls != nil {
		tfMap["source_controls"] = flattenCascadingControlSource(apiObject.SourceControls)
	}

	return []interface{}{tfMap}
}

func flattenCascadingControlSource(apiObject []*quicksight.CascadingControlSource) []interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.ColumnToMatch != nil {
			tfMap["column_to_match"] = flattenColumnIdentifier(config.ColumnToMatch)
		}
		if config.SourceSheetControlId != nil {
			tfMap["source_sheet_control_id"] = aws.StringValue(config.SourceSheetControlId)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDropDownControlDisplayOptions(apiObject *quicksight.DropDownControlDisplayOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.SelectAllOptions != nil {
		tfMap["select_all_options"] = flattenListControlSelectAllOptions(apiObject.SelectAllOptions)
	}
	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []interface{}{tfMap}
}

func flattenListControlSelectAllOptions(apiObject *quicksight.ListControlSelectAllOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Visibility != nil {
		tfMap["visibility"] = aws.StringValue(apiObject.Visibility)
	}

	return []interface{}{tfMap}
}

func flattenFilterSelectableValues(apiObject *quicksight.FilterSelectableValues) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flex.FlattenStringList(apiObject.Values)
	}

	return []interface{}{tfMap}
}

func flattenFilterListControl(apiObject *quicksight.FilterListControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"filter_control_id": aws.StringValue(apiObject.FilterControlId),
		"source_filter_id":  aws.StringValue(apiObject.SourceFilterId),
		"title":             aws.StringValue(apiObject.Title),
	}
	if apiObject.CascadingControlConfiguration != nil {
		tfMap["cacading_control_configuration"] = flattenCascadingControlConfiguration(apiObject.CascadingControlConfiguration)
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenListControlDisplayOptions(apiObject.DisplayOptions)
	}
	if apiObject.SelectableValues != nil {
		tfMap["selectable_values"] = flattenFilterSelectableValues(apiObject.SelectableValues)
	}
	if apiObject.Type != nil {
		tfMap[names.AttrType] = aws.StringValue(apiObject.Type)
	}

	return []interface{}{tfMap}
}

func flattenListControlDisplayOptions(apiObject *quicksight.ListControlDisplayOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.SearchOptions != nil {
		tfMap["search_options"] = flattenListControlSearchOptions(apiObject.SearchOptions)
	}
	if apiObject.SelectAllOptions != nil {
		tfMap["select_all_options"] = flattenListControlSelectAllOptions(apiObject.SelectAllOptions)
	}
	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []interface{}{tfMap}
}

func flattenListControlSearchOptions(apiObject *quicksight.ListControlSearchOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Visibility != nil {
		tfMap["visibility"] = aws.StringValue(apiObject.Visibility)
	}

	return []interface{}{tfMap}
}

func flattenFilterRelativeDateTimeControl(apiObject *quicksight.FilterRelativeDateTimeControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"filter_control_id": aws.StringValue(apiObject.FilterControlId),
		"source_filter_id":  aws.StringValue(apiObject.SourceFilterId),
		"title":             aws.StringValue(apiObject.Title),
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenRelativeDateTimeControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []interface{}{tfMap}
}

func flattenRelativeDateTimeControlDisplayOptions(apiObject *quicksight.RelativeDateTimeControlDisplayOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DateTimeFormat != nil {
		tfMap["date_time_format"] = aws.StringValue(apiObject.DateTimeFormat)
	}
	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []interface{}{tfMap}
}

func flattenFilterSliderControl(apiObject *quicksight.FilterSliderControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"filter_control_id": aws.StringValue(apiObject.FilterControlId),
		"source_filter_id":  aws.StringValue(apiObject.SourceFilterId),
		"title":             aws.StringValue(apiObject.Title),
		"maximum_value":     aws.Float64Value(apiObject.MaximumValue),
		"minimum_value":     aws.Float64Value(apiObject.MinimumValue),
		"step_size":         aws.Float64Value(apiObject.StepSize),
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenSliderControlDisplayOptions(apiObject.DisplayOptions)
	}
	if apiObject.Type != nil {
		tfMap[names.AttrType] = aws.StringValue(apiObject.Type)
	}

	return []interface{}{tfMap}
}

func flattenSliderControlDisplayOptions(apiObject *quicksight.SliderControlDisplayOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []interface{}{tfMap}
}

func flattenFilterTextAreaControl(apiObject *quicksight.FilterTextAreaControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"filter_control_id": aws.StringValue(apiObject.FilterControlId),
		"source_filter_id":  aws.StringValue(apiObject.SourceFilterId),
		"title":             aws.StringValue(apiObject.Title),
	}
	if apiObject.Delimiter != nil {
		tfMap["delimiter"] = aws.StringValue(apiObject.Delimiter)
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenTextAreaControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []interface{}{tfMap}
}

func flattenTextAreaControlDisplayOptions(apiObject *quicksight.TextAreaControlDisplayOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.PlaceholderOptions != nil {
		tfMap["placeholder_options"] = flattenTextControlPlaceholderOptions(apiObject.PlaceholderOptions)
	}
	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []interface{}{tfMap}
}

func flattenTextControlPlaceholderOptions(apiObject *quicksight.TextControlPlaceholderOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Visibility != nil {
		tfMap["visibility"] = aws.StringValue(apiObject.Visibility)
	}

	return []interface{}{tfMap}
}

func flattenFilterTextFieldControl(apiObject *quicksight.FilterTextFieldControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"filter_control_id": aws.StringValue(apiObject.FilterControlId),
		"source_filter_id":  aws.StringValue(apiObject.SourceFilterId),
		"title":             aws.StringValue(apiObject.Title),
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenTextFieldControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []interface{}{tfMap}
}

func flattenTextFieldControlDisplayOptions(apiObject *quicksight.TextFieldControlDisplayOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.PlaceholderOptions != nil {
		tfMap["placeholder_options"] = flattenTextControlPlaceholderOptions(apiObject.PlaceholderOptions)
	}
	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []interface{}{tfMap}
}
