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

func expandFilterControl(tfMap map[string]any) *awstypes.FilterControl {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FilterControl{}

	if v, ok := tfMap["date_time_picker"].([]any); ok && len(v) > 0 {
		apiObject.DateTimePicker = expandFilterDateTimePickerControl(v)
	}
	if v, ok := tfMap["dropdown"].([]any); ok && len(v) > 0 {
		apiObject.Dropdown = expandFilterDropDownControl(v)
	}
	if v, ok := tfMap["list"].([]any); ok && len(v) > 0 {
		apiObject.List = expandFilterListControl(v)
	}
	if v, ok := tfMap["relative_date_time"].([]any); ok && len(v) > 0 {
		apiObject.RelativeDateTime = expandFilterRelativeDateTimeControl(v)
	}
	if v, ok := tfMap["slider"].([]any); ok && len(v) > 0 {
		apiObject.Slider = expandFilterSliderControl(v)
	}
	if v, ok := tfMap["text_area"].([]any); ok && len(v) > 0 {
		apiObject.TextArea = expandFilterTextAreaControl(v)
	}
	if v, ok := tfMap["text_field"].([]any); ok && len(v) > 0 {
		apiObject.TextField = expandFilterTextFieldControl(v)
	}

	return apiObject
}

func expandFilterDateTimePickerControl(tfList []any) *awstypes.FilterDateTimePickerControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
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
	if v, ok := tfMap["display_options"].([]any); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandDateTimePickerControlDisplayOptions(v)
	}

	return apiObject
}

func expandDateTimePickerControlDisplayOptions(tfList []any) *awstypes.DateTimePickerControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DateTimePickerControlDisplayOptions{}

	if v, ok := tfMap["date_time_format"].(string); ok && v != "" {
		apiObject.DateTimeFormat = aws.String(v)
	}
	if v, ok := tfMap["title_options"].([]any); ok && len(v) > 0 {
		apiObject.TitleOptions = expandLabelOptions(v)
	}

	return apiObject
}

func expandFilterDropDownControl(tfList []any) *awstypes.FilterDropDownControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
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
	if v, ok := tfMap["cascading_control_configuration"].([]any); ok && len(v) > 0 {
		apiObject.CascadingControlConfiguration = expandCascadingControlConfiguration(v)
	}
	if v, ok := tfMap["display_options"].([]any); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandDropDownControlDisplayOptions(v)
	}
	if v, ok := tfMap["selectable_values"].([]any); ok && len(v) > 0 {
		apiObject.SelectableValues = expandFilterSelectableValues(v)
	}

	return apiObject
}

func expandCascadingControlConfiguration(tfList []any) *awstypes.CascadingControlConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.CascadingControlConfiguration{}

	if v, ok := tfMap["source_controls"].([]any); ok && len(v) > 0 {
		apiObject.SourceControls = expandCascadingControlSources(v)
	}

	return apiObject
}

func expandCascadingControlSources(tfList []any) []awstypes.CascadingControlSource {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.CascadingControlSource

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
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

func expandCascadingControlSource(tfMap map[string]any) *awstypes.CascadingControlSource {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CascadingControlSource{}

	if v, ok := tfMap["source_sheet_control_id"].(string); ok && v != "" {
		apiObject.SourceSheetControlId = aws.String(v)
	}
	if v, ok := tfMap["column_to_match"].([]any); ok && len(v) > 0 {
		apiObject.ColumnToMatch = expandColumnIdentifier(v)
	}

	return apiObject
}

func expandDropDownControlDisplayOptions(tfList []any) *awstypes.DropDownControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DropDownControlDisplayOptions{}

	if v, ok := tfMap["select_all_options"].([]any); ok && len(v) > 0 {
		apiObject.SelectAllOptions = expandListControlSelectAllOptions(v)
	}
	if v, ok := tfMap["title_options"].([]any); ok && len(v) > 0 {
		apiObject.TitleOptions = expandLabelOptions(v)
	}

	return apiObject
}

func expandListControlSelectAllOptions(tfList []any) *awstypes.ListControlSelectAllOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ListControlSelectAllOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandFilterSelectableValues(tfList []any) *awstypes.FilterSelectableValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilterSelectableValues{}

	if v, ok := tfMap[names.AttrValues].([]any); ok {
		apiObject.Values = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func expandFilterListControl(tfList []any) *awstypes.FilterListControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
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
	if v, ok := tfMap["cascading_control_configuration"].([]any); ok && len(v) > 0 {
		apiObject.CascadingControlConfiguration = expandCascadingControlConfiguration(v)
	}
	if v, ok := tfMap["display_options"].([]any); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandListControlDisplayOptions(v)
	}
	if v, ok := tfMap["selectable_values"].([]any); ok && len(v) > 0 {
		apiObject.SelectableValues = expandFilterSelectableValues(v)
	}

	return apiObject
}

func expandListControlDisplayOptions(tfList []any) *awstypes.ListControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ListControlDisplayOptions{}

	if v, ok := tfMap["select_all_options"].([]any); ok && len(v) > 0 {
		apiObject.SelectAllOptions = expandListControlSelectAllOptions(v)
	}
	if v, ok := tfMap["title_options"].([]any); ok && len(v) > 0 {
		apiObject.TitleOptions = expandLabelOptions(v)
	}
	if v, ok := tfMap["search_options"].([]any); ok && len(v) > 0 {
		apiObject.SearchOptions = expandListControlSearchOptions(v)
	}

	return apiObject
}

func expandListControlSearchOptions(tfList []any) *awstypes.ListControlSearchOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ListControlSearchOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandFilterRelativeDateTimeControl(tfList []any) *awstypes.FilterRelativeDateTimeControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
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
	if v, ok := tfMap["display_options"].([]any); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandRelativeDateTimeControlDisplayOptions(v)
	}

	return apiObject
}

func expandRelativeDateTimeControlDisplayOptions(tfList []any) *awstypes.RelativeDateTimeControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.RelativeDateTimeControlDisplayOptions{}

	if v, ok := tfMap["date_time_format"].(string); ok && v != "" {
		apiObject.DateTimeFormat = aws.String(v)
	}
	if v, ok := tfMap["title_options"].([]any); ok && len(v) > 0 {
		apiObject.TitleOptions = expandLabelOptions(v)
	}

	return apiObject
}

func expandFilterSliderControl(tfList []any) *awstypes.FilterSliderControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
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
	if v, ok := tfMap["display_options"].([]any); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandSliderControlDisplayOptions(v)
	}

	return apiObject
}

func expandSliderControlDisplayOptions(tfList []any) *awstypes.SliderControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.SliderControlDisplayOptions{}

	if v, ok := tfMap["title_options"].([]any); ok && len(v) > 0 {
		apiObject.TitleOptions = expandLabelOptions(v)
	}

	return apiObject
}

func expandFilterTextAreaControl(tfList []any) *awstypes.FilterTextAreaControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
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
	if v, ok := tfMap["display_options"].([]any); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandTextAreaControlDisplayOptions(v)
	}

	return apiObject
}

func expandTextAreaControlDisplayOptions(tfList []any) *awstypes.TextAreaControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TextAreaControlDisplayOptions{}

	if v, ok := tfMap["placeholder_options"].([]any); ok && len(v) > 0 {
		apiObject.PlaceholderOptions = expandTextControlPlaceholderOptions(v)
	}
	if v, ok := tfMap["title_options"].([]any); ok && len(v) > 0 {
		apiObject.TitleOptions = expandLabelOptions(v)
	}

	return apiObject
}

func expandTextControlPlaceholderOptions(tfList []any) *awstypes.TextControlPlaceholderOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TextControlPlaceholderOptions{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandFilterTextFieldControl(tfList []any) *awstypes.FilterTextFieldControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
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
	if v, ok := tfMap["display_options"].([]any); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandTextFieldControlDisplayOptions(v)
	}

	return apiObject
}

func expandTextFieldControlDisplayOptions(tfList []any) *awstypes.TextFieldControlDisplayOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TextFieldControlDisplayOptions{}

	if v, ok := tfMap["placeholder_options"].([]any); ok && len(v) > 0 {
		apiObject.PlaceholderOptions = expandTextControlPlaceholderOptions(v)
	}
	if v, ok := tfMap["title_options"].([]any); ok && len(v) > 0 {
		apiObject.TitleOptions = expandLabelOptions(v)
	}

	return apiObject
}

func expandParameterControl(tfMap map[string]any) *awstypes.ParameterControl {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ParameterControl{}

	if v, ok := tfMap["date_time_picker"].([]any); ok && len(v) > 0 {
		apiObject.DateTimePicker = expandParameterDateTimePickerControl(v)
	}
	if v, ok := tfMap["dropdown"].([]any); ok && len(v) > 0 {
		apiObject.Dropdown = expandParameterDropDownControl(v)
	}
	if v, ok := tfMap["list"].([]any); ok && len(v) > 0 {
		apiObject.List = expandParameterListControl(v)
	}
	if v, ok := tfMap["slider"].([]any); ok && len(v) > 0 {
		apiObject.Slider = expandParameterSliderControl(v)
	}
	if v, ok := tfMap["text_area"].([]any); ok && len(v) > 0 {
		apiObject.TextArea = expandParameterTextAreaControl(v)
	}
	if v, ok := tfMap["text_field"].([]any); ok && len(v) > 0 {
		apiObject.TextField = expandParameterTextFieldControl(v)
	}

	return apiObject
}

func expandParameterDateTimePickerControl(tfList []any) *awstypes.ParameterDateTimePickerControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
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
	if v, ok := tfMap["display_options"].([]any); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandDateTimePickerControlDisplayOptions(v)
	}

	return apiObject
}

func expandParameterDropDownControl(tfList []any) *awstypes.ParameterDropDownControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
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
	if v, ok := tfMap["cascading_control_configuration"].([]any); ok && len(v) > 0 {
		apiObject.CascadingControlConfiguration = expandCascadingControlConfiguration(v)
	}
	if v, ok := tfMap["display_options"].([]any); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandDropDownControlDisplayOptions(v)
	}
	if v, ok := tfMap["selectable_values"].([]any); ok && len(v) > 0 {
		apiObject.SelectableValues = expandParameterSelectableValues(v)
	}

	return apiObject
}

func expandParameterListControl(tfList []any) *awstypes.ParameterListControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
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
	if v, ok := tfMap["cascading_control_configuration"].([]any); ok && len(v) > 0 {
		apiObject.CascadingControlConfiguration = expandCascadingControlConfiguration(v)
	}
	if v, ok := tfMap["display_options"].([]any); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandListControlDisplayOptions(v)
	}
	if v, ok := tfMap["selectable_values"].([]any); ok && len(v) > 0 {
		apiObject.SelectableValues = expandParameterSelectableValues(v)
	}

	return apiObject
}

func expandParameterSliderControl(tfList []any) *awstypes.ParameterSliderControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
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
	if v, ok := tfMap["display_options"].([]any); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandSliderControlDisplayOptions(v)
	}

	return apiObject
}

func expandParameterTextAreaControl(tfList []any) *awstypes.ParameterTextAreaControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
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
	if v, ok := tfMap["display_options"].([]any); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandTextAreaControlDisplayOptions(v)
	}

	return apiObject
}

func expandParameterTextFieldControl(tfList []any) *awstypes.ParameterTextFieldControl {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
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
	if v, ok := tfMap["display_options"].([]any); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandTextFieldControlDisplayOptions(v)
	}

	return apiObject
}

func flattenFilterControls(apiObjects []awstypes.FilterControl) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

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

func flattenFilterDateTimePickerControl(apiObject *awstypes.FilterDateTimePickerControl) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"filter_control_id": aws.ToString(apiObject.FilterControlId),
		"source_filter_id":  aws.ToString(apiObject.SourceFilterId),
		"title":             aws.ToString(apiObject.Title),
	}

	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenDateTimePickerControlDisplayOptions(apiObject.DisplayOptions)
	}
	tfMap[names.AttrType] = apiObject.Type

	return []any{tfMap}
}

func flattenDateTimePickerControlDisplayOptions(apiObject *awstypes.DateTimePickerControlDisplayOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DateTimeFormat != nil {
		tfMap["date_time_format"] = aws.ToString(apiObject.DateTimeFormat)
	}
	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []any{tfMap}
}

func flattenLabelOptions(apiObject *awstypes.LabelOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CustomLabel != nil {
		tfMap["custom_label"] = aws.ToString(apiObject.CustomLabel)
	}
	if apiObject.FontConfiguration != nil {
		tfMap["font_configuration"] = flattenFontConfiguration(apiObject.FontConfiguration)
	}
	tfMap["visibility"] = apiObject.Visibility

	return []any{tfMap}
}

func flattenFontConfiguration(apiObject *awstypes.FontConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

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

	return []any{tfMap}
}

func flattenFontSize(apiObject *awstypes.FontSize) []any {
	if apiObject == nil || apiObject.Relative == "" {
		return nil
	}

	tfMap := map[string]any{
		"relative": apiObject.Relative,
	}

	return []any{tfMap}
}

func flattenFontWeight(apiObject *awstypes.FontWeight) []any {
	if apiObject == nil || apiObject.Name == "" {
		return nil
	}

	tfMap := map[string]any{
		names.AttrName: apiObject.Name,
	}

	return []any{tfMap}
}

func flattenFilterDropDownControl(apiObject *awstypes.FilterDropDownControl) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
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

	return []any{tfMap}
}

func flattenCascadingControlConfiguration(apiObject *awstypes.CascadingControlConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.SourceControls != nil {
		tfMap["source_controls"] = flattenCascadingControlSource(apiObject.SourceControls)
	}

	return []any{tfMap}
}

func flattenCascadingControlSource(apiObjects []awstypes.CascadingControlSource) []any {
	if apiObjects == nil {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

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

func flattenDropDownControlDisplayOptions(apiObject *awstypes.DropDownControlDisplayOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.SelectAllOptions != nil {
		tfMap["select_all_options"] = flattenListControlSelectAllOptions(apiObject.SelectAllOptions)
	}
	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []any{tfMap}
}

func flattenListControlSelectAllOptions(apiObject *awstypes.ListControlSelectAllOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"visibility": apiObject.Visibility,
	}

	return []any{tfMap}
}

func flattenFilterSelectableValues(apiObject *awstypes.FilterSelectableValues) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Values != nil {
		tfMap[names.AttrValues] = apiObject.Values
	}

	return []any{tfMap}
}

func flattenFilterListControl(apiObject *awstypes.FilterListControl) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
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

	return []any{tfMap}
}

func flattenListControlDisplayOptions(apiObject *awstypes.ListControlDisplayOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.SearchOptions != nil {
		tfMap["search_options"] = flattenListControlSearchOptions(apiObject.SearchOptions)
	}
	if apiObject.SelectAllOptions != nil {
		tfMap["select_all_options"] = flattenListControlSelectAllOptions(apiObject.SelectAllOptions)
	}
	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []any{tfMap}
}

func flattenListControlSearchOptions(apiObject *awstypes.ListControlSearchOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"visibility": apiObject.Visibility,
	}

	return []any{tfMap}
}

func flattenFilterRelativeDateTimeControl(apiObject *awstypes.FilterRelativeDateTimeControl) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"filter_control_id": aws.ToString(apiObject.FilterControlId),
		"source_filter_id":  aws.ToString(apiObject.SourceFilterId),
		"title":             aws.ToString(apiObject.Title),
	}

	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenRelativeDateTimeControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []any{tfMap}
}

func flattenRelativeDateTimeControlDisplayOptions(apiObject *awstypes.RelativeDateTimeControlDisplayOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DateTimeFormat != nil {
		tfMap["date_time_format"] = aws.ToString(apiObject.DateTimeFormat)
	}
	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []any{tfMap}
}

func flattenFilterSliderControl(apiObject *awstypes.FilterSliderControl) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
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

	return []any{tfMap}
}

func flattenSliderControlDisplayOptions(apiObject *awstypes.SliderControlDisplayOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []any{tfMap}
}

func flattenFilterTextAreaControl(apiObject *awstypes.FilterTextAreaControl) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
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

	return []any{tfMap}
}

func flattenTextAreaControlDisplayOptions(apiObject *awstypes.TextAreaControlDisplayOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.PlaceholderOptions != nil {
		tfMap["placeholder_options"] = flattenTextControlPlaceholderOptions(apiObject.PlaceholderOptions)
	}
	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []any{tfMap}
}

func flattenTextControlPlaceholderOptions(apiObject *awstypes.TextControlPlaceholderOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"visibility": apiObject.Visibility,
	}

	return []any{tfMap}
}

func flattenFilterTextFieldControl(apiObject *awstypes.FilterTextFieldControl) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"filter_control_id": aws.ToString(apiObject.FilterControlId),
		"source_filter_id":  aws.ToString(apiObject.SourceFilterId),
		"title":             aws.ToString(apiObject.Title),
	}

	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenTextFieldControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []any{tfMap}
}

func flattenTextFieldControlDisplayOptions(apiObject *awstypes.TextFieldControlDisplayOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.PlaceholderOptions != nil {
		tfMap["placeholder_options"] = flattenTextControlPlaceholderOptions(apiObject.PlaceholderOptions)
	}
	if apiObject.TitleOptions != nil {
		tfMap["title_options"] = flattenLabelOptions(apiObject.TitleOptions)
	}

	return []any{tfMap}
}
