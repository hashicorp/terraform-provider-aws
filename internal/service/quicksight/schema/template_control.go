// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var filterControlsSchema = sync.OnceValue(func() *schema.Schema {
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
							"title":             stringLenBetweenSchema(attrRequired, 1, 2048),
							"display_options":   dateTimePickerControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimePickerControlDisplayOptions.html
							names.AttrType:      stringEnumSchema[awstypes.SheetControlDateTimePickerType](attrOptional),
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
							"title":                           stringLenBetweenSchema(attrRequired, 1, 2048),
							"cascading_control_configuration": cascadingControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CascadingControlConfiguration.html
							"display_options":                 dropDownControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DropDownControlDisplayOptions.html
							"selectable_values":               filterSelectableValuesSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterSelectableValues.html
							names.AttrType:                    stringEnumSchema[awstypes.SheetControlListType](attrOptional),
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
							"title":                           stringLenBetweenSchema(attrRequired, 1, 2048),
							"cascading_control_configuration": cascadingControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CascadingControlConfiguration.html
							"display_options":                 listControlDisplayOptionsSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ListControlDisplayOptions.html
							"selectable_values":               filterSelectableValuesSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterSelectableValues.html
							names.AttrType:                    stringEnumSchema[awstypes.SheetControlListType](attrOptional),
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
							"title":             stringLenBetweenSchema(attrRequired, 1, 2048),
							"display_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RelativeDateTimeControlDisplayOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"date_time_format": stringLenBetweenSchema(attrOptional, 1, 128),
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
							"title":             stringLenBetweenSchema(attrRequired, 1, 2048),
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
							names.AttrType:    stringEnumSchema[awstypes.SheetControlSliderType](attrOptional),
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
							"title":             stringLenBetweenSchema(attrRequired, 1, 2048),
							"delimiter":         stringLenBetweenSchema(attrOptional, 1, 2048),
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
							"title":             stringLenBetweenSchema(attrRequired, 1, 2048),
							"display_options":   textFieldControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextFieldControlDisplayOptions.html
						},
					},
				},
			},
		},
	}
})

var textFieldControlDisplayOptionsSchema = sync.OnceValue(func() *schema.Schema {
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
})

var textAreaControlDisplayOptionsSchema = sync.OnceValue(func() *schema.Schema {
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
})

var sliderControlDisplayOptionsSchema = sync.OnceValue(func() *schema.Schema {
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
})

var dateTimePickerControlDisplayOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimePickerControlDisplayOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"date_time_format": stringLenBetweenSchema(attrOptional, 1, 128),
				"title_options":    labelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LabelOptions.html
			},
		},
	}
})

var listControlDisplayOptionsSchema = sync.OnceValue(func() *schema.Schema {
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
							"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
						},
					},
				},
				"select_all_options": selectAllOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ListControlSelectAllOptions.html
				"title_options":      labelOptionsSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LabelOptions.html
			},
		},
	}
})

var cascadingControlConfigurationSchema = sync.OnceValue(func() *schema.Schema {
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
})

var selectAllOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ListControlSelectAllOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
			},
		},
	}
})

var dropDownControlDisplayOptionsSchema = sync.OnceValue(func() *schema.Schema {
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
})

var placeholderOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextControlPlaceholderOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
			},
		},
	}
})

func expandFilterControl(tfMap map[string]interface{}) *awstypes.FilterControl {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FilterControl{}

	if v, ok := tfMap["date_time_picker"].([]interface{}); ok && len(v) > 0 {
		apiObject.DateTimePicker = expandFilterDateTimePickerControl(v)
	}
	if v, ok := tfMap["dropdown"].([]interface{}); ok && len(v) > 0 {
		apiObject.Dropdown = expandFilterDropDownControl(v)
	}
	if v, ok := tfMap["list"].([]interface{}); ok && len(v) > 0 {
		apiObject.List = expandFilterListControl(v)
	}
	if v, ok := tfMap["relative_date_time"].([]interface{}); ok && len(v) > 0 {
		apiObject.RelativeDateTime = expandFilterRelativeDateTimeControl(v)
	}
	if v, ok := tfMap["slider"].([]interface{}); ok && len(v) > 0 {
		apiObject.Slider = expandFilterSliderControl(v)
	}
	if v, ok := tfMap["text_area"].([]interface{}); ok && len(v) > 0 {
		apiObject.TextArea = expandFilterTextAreaControl(v)
	}
	if v, ok := tfMap["text_field"].([]interface{}); ok && len(v) > 0 {
		apiObject.TextField = expandFilterTextFieldControl(v)
	}

	return apiObject
}

func expandFilterDateTimePickerControl(tfList []interface{}) *awstypes.FilterDateTimePickerControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilterDateTimePickerControl{}

	if v, ok := tfMap["filter_control_id"].(string); ok && v != "" {
		apiObject.FilterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_filter_id"].(string); ok && v != "" {
		apiObject.SourceFilterId = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		apiObject.Title = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.SheetControlDateTimePickerType(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandDateTimePickerControlDisplayOptions(v)
	}

	return apiObject
}

func expandDateTimePickerControlDisplayOptions(tfList []interface{}) *awstypes.DateTimePickerControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DateTimePickerControlDisplayOptions{}

	if v, ok := tfMap["date_time_format"].(string); ok && v != "" {
		apiObject.DateTimeFormat = aws.String(v)
	}
	if v, ok := tfMap["title_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.TitleOptions = expandLabelOptions(v)
	}

	return apiObject
}

func expandFilterDropDownControl(tfList []interface{}) *awstypes.FilterDropDownControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilterDropDownControl{}

	if v, ok := tfMap["filter_control_id"].(string); ok && v != "" {
		apiObject.FilterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_filter_id"].(string); ok && v != "" {
		apiObject.SourceFilterId = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		apiObject.Title = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.SheetControlListType(v)
	}
	if v, ok := tfMap["cascading_control_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.CascadingControlConfiguration = expandCascadingControlConfiguration(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandDropDownControlDisplayOptions(v)
	}
	if v, ok := tfMap["selectable_values"].([]interface{}); ok && len(v) > 0 {
		apiObject.SelectableValues = expandFilterSelectableValues(v)
	}

	return apiObject
}

func expandCascadingControlConfiguration(tfList []interface{}) *awstypes.CascadingControlConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.CascadingControlConfiguration{}

	if v, ok := tfMap["source_controls"].([]interface{}); ok && len(v) > 0 {
		apiObject.SourceControls = expandCascadingControlSources(v)
	}

	return apiObject
}

func expandCascadingControlSources(tfList []interface{}) []awstypes.CascadingControlSource {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.CascadingControlSource

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandCascadingControlSource(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandCascadingControlSource(tfMap map[string]interface{}) *awstypes.CascadingControlSource {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CascadingControlSource{}

	if v, ok := tfMap["source_sheet_control_id"].(string); ok && v != "" {
		apiObject.SourceSheetControlId = aws.String(v)
	}
	if v, ok := tfMap["column_to_match"].([]interface{}); ok && len(v) > 0 {
		apiObject.ColumnToMatch = expandColumnIdentifier(v)
	}

	return apiObject
}

func expandDropDownControlDisplayOptions(tfList []interface{}) *awstypes.DropDownControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DropDownControlDisplayOptions{}

	if v, ok := tfMap["select_all_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.SelectAllOptions = expandListControlSelectAllOptions(v)
	}
	if v, ok := tfMap["title_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.TitleOptions = expandLabelOptions(v)
	}

	return apiObject
}

func expandListControlSelectAllOptions(tfList []interface{}) *awstypes.ListControlSelectAllOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ListControlSelectAllOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandFilterSelectableValues(tfList []interface{}) *awstypes.FilterSelectableValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilterSelectableValues{}

	if v, ok := tfMap[names.AttrValues].([]interface{}); ok {
		apiObject.Values = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func expandFilterListControl(tfList []interface{}) *awstypes.FilterListControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilterListControl{}

	if v, ok := tfMap["filter_control_id"].(string); ok && v != "" {
		apiObject.FilterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_filter_id"].(string); ok && v != "" {
		apiObject.SourceFilterId = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		apiObject.Title = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.SheetControlListType(v)
	}
	if v, ok := tfMap["cascading_control_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.CascadingControlConfiguration = expandCascadingControlConfiguration(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandListControlDisplayOptions(v)
	}
	if v, ok := tfMap["selectable_values"].([]interface{}); ok && len(v) > 0 {
		apiObject.SelectableValues = expandFilterSelectableValues(v)
	}

	return apiObject
}

func expandListControlDisplayOptions(tfList []interface{}) *awstypes.ListControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ListControlDisplayOptions{}

	if v, ok := tfMap["select_all_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.SelectAllOptions = expandListControlSelectAllOptions(v)
	}
	if v, ok := tfMap["title_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.TitleOptions = expandLabelOptions(v)
	}
	if v, ok := tfMap["search_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.SearchOptions = expandListControlSearchOptions(v)
	}

	return apiObject
}

func expandListControlSearchOptions(tfList []interface{}) *awstypes.ListControlSearchOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ListControlSearchOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandFilterRelativeDateTimeControl(tfList []interface{}) *awstypes.FilterRelativeDateTimeControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilterRelativeDateTimeControl{}

	if v, ok := tfMap["filter_control_id"].(string); ok && v != "" {
		apiObject.FilterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_filter_id"].(string); ok && v != "" {
		apiObject.SourceFilterId = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		apiObject.Title = aws.String(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandRelativeDateTimeControlDisplayOptions(v)
	}

	return apiObject
}

func expandRelativeDateTimeControlDisplayOptions(tfList []interface{}) *awstypes.RelativeDateTimeControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.RelativeDateTimeControlDisplayOptions{}

	if v, ok := tfMap["date_time_format"].(string); ok && v != "" {
		apiObject.DateTimeFormat = aws.String(v)
	}
	if v, ok := tfMap["title_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.TitleOptions = expandLabelOptions(v)
	}

	return apiObject
}

func expandFilterSliderControl(tfList []interface{}) *awstypes.FilterSliderControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilterSliderControl{}

	if v, ok := tfMap["filter_control_id"].(string); ok && v != "" {
		apiObject.FilterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_filter_id"].(string); ok && v != "" {
		apiObject.SourceFilterId = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		apiObject.Title = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.SheetControlSliderType(v)
	}
	if v, ok := tfMap["maximum_value"].(float64); ok {
		apiObject.MaximumValue = v
	}
	if v, ok := tfMap["minimum_value"].(float64); ok {
		apiObject.MinimumValue = v
	}
	if v, ok := tfMap["step_size"].(float64); ok {
		apiObject.StepSize = v
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandSliderControlDisplayOptions(v)
	}

	return apiObject
}

func expandSliderControlDisplayOptions(tfList []interface{}) *awstypes.SliderControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SliderControlDisplayOptions{}

	if v, ok := tfMap["title_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.TitleOptions = expandLabelOptions(v)
	}

	return apiObject
}

func expandFilterTextAreaControl(tfList []interface{}) *awstypes.FilterTextAreaControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilterTextAreaControl{}

	if v, ok := tfMap["filter_control_id"].(string); ok && v != "" {
		apiObject.FilterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_filter_id"].(string); ok && v != "" {
		apiObject.SourceFilterId = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		apiObject.Title = aws.String(v)
	}
	if v, ok := tfMap["delimiter"].(string); ok && v != "" {
		apiObject.Delimiter = aws.String(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandTextAreaControlDisplayOptions(v)
	}

	return apiObject
}

func expandTextAreaControlDisplayOptions(tfList []interface{}) *awstypes.TextAreaControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.TextAreaControlDisplayOptions{}

	if v, ok := tfMap["placeholder_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.PlaceholderOptions = expandTextControlPlaceholderOptions(v)
	}
	if v, ok := tfMap["title_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.TitleOptions = expandLabelOptions(v)
	}

	return apiObject
}

func expandTextControlPlaceholderOptions(tfList []interface{}) *awstypes.TextControlPlaceholderOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.TextControlPlaceholderOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandFilterTextFieldControl(tfList []interface{}) *awstypes.FilterTextFieldControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilterTextFieldControl{}

	if v, ok := tfMap["filter_control_id"].(string); ok && v != "" {
		apiObject.FilterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_filter_id"].(string); ok && v != "" {
		apiObject.SourceFilterId = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		apiObject.Title = aws.String(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandTextFieldControlDisplayOptions(v)
	}

	return apiObject
}

func expandTextFieldControlDisplayOptions(tfList []interface{}) *awstypes.TextFieldControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.TextFieldControlDisplayOptions{}

	if v, ok := tfMap["placeholder_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.PlaceholderOptions = expandTextControlPlaceholderOptions(v)
	}
	if v, ok := tfMap["title_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.TitleOptions = expandLabelOptions(v)
	}

	return apiObject
}

func expandParameterControl(tfMap map[string]interface{}) *awstypes.ParameterControl {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ParameterControl{}

	if v, ok := tfMap["date_time_picker"].([]interface{}); ok && len(v) > 0 {
		apiObject.DateTimePicker = expandParameterDateTimePickerControl(v)
	}
	if v, ok := tfMap["dropdown"].([]interface{}); ok && len(v) > 0 {
		apiObject.Dropdown = expandParameterDropDownControl(v)
	}
	if v, ok := tfMap["list"].([]interface{}); ok && len(v) > 0 {
		apiObject.List = expandParameterListControl(v)
	}
	if v, ok := tfMap["slider"].([]interface{}); ok && len(v) > 0 {
		apiObject.Slider = expandParameterSliderControl(v)
	}
	if v, ok := tfMap["text_area"].([]interface{}); ok && len(v) > 0 {
		apiObject.TextArea = expandParameterTextAreaControl(v)
	}
	if v, ok := tfMap["text_field"].([]interface{}); ok && len(v) > 0 {
		apiObject.TextField = expandParameterTextFieldControl(v)
	}

	return apiObject
}

func expandParameterDateTimePickerControl(tfList []interface{}) *awstypes.ParameterDateTimePickerControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ParameterDateTimePickerControl{}

	if v, ok := tfMap["parameter_control_id"].(string); ok && v != "" {
		apiObject.ParameterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_parameter_name"].(string); ok && v != "" {
		apiObject.SourceParameterName = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		apiObject.Title = aws.String(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandDateTimePickerControlDisplayOptions(v)
	}

	return apiObject
}

func expandParameterDropDownControl(tfList []interface{}) *awstypes.ParameterDropDownControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ParameterDropDownControl{}

	if v, ok := tfMap["parameter_control_id"].(string); ok && v != "" {
		apiObject.ParameterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_parameter_name"].(string); ok && v != "" {
		apiObject.SourceParameterName = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		apiObject.Title = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.SheetControlListType(v)
	}
	if v, ok := tfMap["cascading_control_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.CascadingControlConfiguration = expandCascadingControlConfiguration(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandDropDownControlDisplayOptions(v)
	}
	if v, ok := tfMap["selectable_values"].([]interface{}); ok && len(v) > 0 {
		apiObject.SelectableValues = expandParameterSelectableValues(v)
	}

	return apiObject
}

func expandParameterListControl(tfList []interface{}) *awstypes.ParameterListControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ParameterListControl{}

	if v, ok := tfMap["parameter_control_id"].(string); ok && v != "" {
		apiObject.ParameterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_parameter_name"].(string); ok && v != "" {
		apiObject.SourceParameterName = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		apiObject.Title = aws.String(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.SheetControlListType(v)
	}
	if v, ok := tfMap["cascading_control_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.CascadingControlConfiguration = expandCascadingControlConfiguration(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandListControlDisplayOptions(v)
	}
	if v, ok := tfMap["selectable_values"].([]interface{}); ok && len(v) > 0 {
		apiObject.SelectableValues = expandParameterSelectableValues(v)
	}

	return apiObject
}

func expandParameterSliderControl(tfList []interface{}) *awstypes.ParameterSliderControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ParameterSliderControl{}

	if v, ok := tfMap["parameter_control_id"].(string); ok && v != "" {
		apiObject.ParameterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_parameter_name"].(string); ok && v != "" {
		apiObject.SourceParameterName = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		apiObject.Title = aws.String(v)
	}
	if v, ok := tfMap["maximum_value"].(float64); ok {
		apiObject.MaximumValue = v
	}
	if v, ok := tfMap["minimum_value"].(float64); ok {
		apiObject.MinimumValue = v
	}
	if v, ok := tfMap["step_size"].(float64); ok {
		apiObject.StepSize = v
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandSliderControlDisplayOptions(v)
	}

	return apiObject
}

func expandParameterTextAreaControl(tfList []interface{}) *awstypes.ParameterTextAreaControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ParameterTextAreaControl{}

	if v, ok := tfMap["parameter_control_id"].(string); ok && v != "" {
		apiObject.ParameterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_parameter_name"].(string); ok && v != "" {
		apiObject.SourceParameterName = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		apiObject.Title = aws.String(v)
	}
	if v, ok := tfMap["delimiter"].(string); ok && v != "" {
		apiObject.Delimiter = aws.String(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandTextAreaControlDisplayOptions(v)
	}

	return apiObject
}

func expandParameterTextFieldControl(tfList []interface{}) *awstypes.ParameterTextFieldControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ParameterTextFieldControl{}

	if v, ok := tfMap["parameter_control_id"].(string); ok && v != "" {
		apiObject.ParameterControlId = aws.String(v)
	}
	if v, ok := tfMap["source_parameter_name"].(string); ok && v != "" {
		apiObject.SourceParameterName = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		apiObject.Title = aws.String(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandTextFieldControlDisplayOptions(v)
	}

	return apiObject
}

func flattenFilterControls(apiObjects []awstypes.FilterControl) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.DateTimePicker != nil {
			tfMap["date_time_picker"] = flattenFilterDateTimePickerControl(apiObject.DateTimePicker)
		}
		if apiObject.Dropdown != nil {
			tfMap["dropdown"] = flattenFilterDropDownControl(apiObject.Dropdown)
		}
		if apiObject.List != nil {
			tfMap["list"] = flattenFilterListControl(apiObject.List)
		}
		if apiObject.RelativeDateTime != nil {
			tfMap["relative_date_time"] = flattenFilterRelativeDateTimeControl(apiObject.RelativeDateTime)
		}
		if apiObject.Slider != nil {
			tfMap["slider"] = flattenFilterSliderControl(apiObject.Slider)
		}
		if apiObject.TextArea != nil {
			tfMap["text_area"] = flattenFilterTextAreaControl(apiObject.TextArea)
		}
		if apiObject.TextField != nil {
			tfMap["text_field"] = flattenFilterTextFieldControl(apiObject.TextField)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenFilterDateTimePickerControl(apiObject *awstypes.FilterDateTimePickerControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"filter_control_id": aws.ToString(apiObject.FilterControlId),
		"source_filter_id":  aws.ToString(apiObject.SourceFilterId),
		"title":             aws.ToString(apiObject.Title),
	}

	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenDateTimePickerControlDisplayOptions(apiObject.DisplayOptions)
	}
	tfMap[names.AttrType] = apiObject.Type

	return []interface{}{tfMap}
}

func flattenDateTimePickerControlDisplayOptions(apiObject *awstypes.DateTimePickerControlDisplayOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.DateTimeFormat != nil {
		tfMap["date_time_format"] = aws.ToString(apiObject.DateTimeFormat)
	}
	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []interface{}{tfMap}
}

func flattenLabelOptions(apiObject *awstypes.LabelOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.CustomLabel != nil {
		tfMap["custom_label"] = aws.ToString(apiObject.CustomLabel)
	}
	if apiObject.FontConfiguration != nil {
		tfMap["font_configuration"] = flattenFontConfiguration(apiObject.FontConfiguration)
	}
	tfMap["visibility"] = apiObject.Visibility

	return []interface{}{tfMap}
}

func flattenFontConfiguration(apiObject *awstypes.FontConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.FontColor != nil {
		tfMap["font_color"] = aws.ToString(apiObject.FontColor)
	}
	tfMap["font_decoration"] = apiObject.FontDecoration
	if apiObject.FontSize != nil {
		tfMap["font_size"] = flattenFontSize(apiObject.FontSize)
	}
	tfMap["font_style"] = apiObject.FontStyle
	if apiObject.FontWeight != nil {
		tfMap["font_weight"] = flattenFontWeight(apiObject.FontWeight)
	}

	return []interface{}{tfMap}
}

func flattenFontSize(apiObject *awstypes.FontSize) []interface{} {
	if apiObject == nil || apiObject.Relative == "" {
		return nil
	}

	tfMap := map[string]interface{}{
		"relative": apiObject.Relative,
	}

	return []interface{}{tfMap}
}

func flattenFontWeight(apiObject *awstypes.FontWeight) []interface{} {
	if apiObject == nil || apiObject.Name == "" {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrName: apiObject.Name,
	}

	return []interface{}{tfMap}
}

func flattenFilterDropDownControl(apiObject *awstypes.FilterDropDownControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"filter_control_id": aws.ToString(apiObject.FilterControlId),
		"source_filter_id":  aws.ToString(apiObject.SourceFilterId),
		"title":             aws.ToString(apiObject.Title),
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
	tfMap[names.AttrType] = apiObject.Type

	return []interface{}{tfMap}
}

func flattenCascadingControlConfiguration(apiObject *awstypes.CascadingControlConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.SourceControls != nil {
		tfMap["source_controls"] = flattenCascadingControlSource(apiObject.SourceControls)
	}

	return []interface{}{tfMap}
}

func flattenCascadingControlSource(apiObjects []awstypes.CascadingControlSource) []interface{} {
	if apiObjects == nil {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.ColumnToMatch != nil {
			tfMap["column_to_match"] = flattenColumnIdentifier(apiObject.ColumnToMatch)
		}
		if apiObject.SourceSheetControlId != nil {
			tfMap["source_sheet_control_id"] = aws.ToString(apiObject.SourceSheetControlId)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDropDownControlDisplayOptions(apiObject *awstypes.DropDownControlDisplayOptions) []interface{} {
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

func flattenListControlSelectAllOptions(apiObject *awstypes.ListControlSelectAllOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"visibility": apiObject.Visibility,
	}

	return []interface{}{tfMap}
}

func flattenFilterSelectableValues(apiObject *awstypes.FilterSelectableValues) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Values != nil {
		tfMap[names.AttrValues] = apiObject.Values
	}

	return []interface{}{tfMap}
}

func flattenFilterListControl(apiObject *awstypes.FilterListControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"filter_control_id": aws.ToString(apiObject.FilterControlId),
		"source_filter_id":  aws.ToString(apiObject.SourceFilterId),
		"title":             aws.ToString(apiObject.Title),
	}

	if apiObject.CascadingControlConfiguration != nil {
		tfMap["cascading_control_configuration"] = flattenCascadingControlConfiguration(apiObject.CascadingControlConfiguration)
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenListControlDisplayOptions(apiObject.DisplayOptions)
	}
	if apiObject.SelectableValues != nil {
		tfMap["selectable_values"] = flattenFilterSelectableValues(apiObject.SelectableValues)
	}
	tfMap[names.AttrType] = apiObject.Type

	return []interface{}{tfMap}
}

func flattenListControlDisplayOptions(apiObject *awstypes.ListControlDisplayOptions) []interface{} {
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

func flattenListControlSearchOptions(apiObject *awstypes.ListControlSearchOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"visibility": apiObject.Visibility,
	}

	return []interface{}{tfMap}
}

func flattenFilterRelativeDateTimeControl(apiObject *awstypes.FilterRelativeDateTimeControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"filter_control_id": aws.ToString(apiObject.FilterControlId),
		"source_filter_id":  aws.ToString(apiObject.SourceFilterId),
		"title":             aws.ToString(apiObject.Title),
	}

	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenRelativeDateTimeControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []interface{}{tfMap}
}

func flattenRelativeDateTimeControlDisplayOptions(apiObject *awstypes.RelativeDateTimeControlDisplayOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.DateTimeFormat != nil {
		tfMap["date_time_format"] = aws.ToString(apiObject.DateTimeFormat)
	}
	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []interface{}{tfMap}
}

func flattenFilterSliderControl(apiObject *awstypes.FilterSliderControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"filter_control_id": aws.ToString(apiObject.FilterControlId),
		"source_filter_id":  aws.ToString(apiObject.SourceFilterId),
		"title":             aws.ToString(apiObject.Title),
		"maximum_value":     apiObject.MaximumValue,
		"minimum_value":     apiObject.MinimumValue,
		"step_size":         apiObject.StepSize,
	}

	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenSliderControlDisplayOptions(apiObject.DisplayOptions)
	}
	tfMap[names.AttrType] = apiObject.Type

	return []interface{}{tfMap}
}

func flattenSliderControlDisplayOptions(apiObject *awstypes.SliderControlDisplayOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []interface{}{tfMap}
}

func flattenFilterTextAreaControl(apiObject *awstypes.FilterTextAreaControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"filter_control_id": aws.ToString(apiObject.FilterControlId),
		"source_filter_id":  aws.ToString(apiObject.SourceFilterId),
		"title":             aws.ToString(apiObject.Title),
	}

	if apiObject.Delimiter != nil {
		tfMap["delimiter"] = aws.ToString(apiObject.Delimiter)
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenTextAreaControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []interface{}{tfMap}
}

func flattenTextAreaControlDisplayOptions(apiObject *awstypes.TextAreaControlDisplayOptions) []interface{} {
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

func flattenTextControlPlaceholderOptions(apiObject *awstypes.TextControlPlaceholderOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"visibility": apiObject.Visibility,
	}

	return []interface{}{tfMap}
}

func flattenFilterTextFieldControl(apiObject *awstypes.FilterTextFieldControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"filter_control_id": aws.ToString(apiObject.FilterControlId),
		"source_filter_id":  aws.ToString(apiObject.SourceFilterId),
		"title":             aws.ToString(apiObject.Title),
	}

	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenTextFieldControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []interface{}{tfMap}
}

func flattenTextFieldControlDisplayOptions(apiObject *awstypes.TextFieldControlDisplayOptions) []interface{} {
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
