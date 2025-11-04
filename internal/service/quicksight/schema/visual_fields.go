// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type measureFieldsSize int

const (
	measureFieldsMaxItems5   measureFieldsSize = 5
	measureFieldsMaxItems20  measureFieldsSize = 20
	measureFieldsMaxItems40  measureFieldsSize = 40
	measureFieldsMaxItems200 measureFieldsSize = 200
)

type dimensionFieldSize int

const (
	dimensionsFieldMaxItems10  dimensionFieldSize = 10
	dimensionsFieldMaxItems40  dimensionFieldSize = 40
	dimensionsFieldMaxItems200 dimensionFieldSize = 200
)

type dimensionFieldSchemaIdentity dimensionFieldSize

var dimensionFieldSchemaCache syncMap[dimensionFieldSchemaIdentity, *schema.Schema]

func dimensionFieldSchema(maxItems dimensionFieldSize) *schema.Schema {
	id := dimensionFieldSchemaIdentity(maxItems)

	s, ok := dimensionFieldSchemaCache.Load(id)
	if ok {
		return s
	}

	// Use a separate `LoadOrStore` to avoid allocation if item is already in the cache
	// Use `LoadOrStore` instead of `Store` in case there is a race
	s, _ = dimensionFieldSchemaCache.LoadOrStore(
		id,
		&schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
			Type:     schema.TypeList,
			MinItems: 1,
			MaxItems: int(maxItems),
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"categorical_dimension_field": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CategoricalDimensionField.html
						Type:     schema.TypeList,
						MinItems: 1,
						MaxItems: 1,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"column":               columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
								"field_id":             stringLenBetweenSchema(attrRequired, 1, 512),
								"format_configuration": stringFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_StringFormatConfiguration.html
								"hierarchy_id":         stringLenBetweenSchema(attrOptional, 1, 512),
							},
						},
					},
					"date_dimension_field": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateDimensionField.html
						Type:     schema.TypeList,
						MinItems: 1,
						MaxItems: 1,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"column":               columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
								"field_id":             stringLenBetweenSchema(attrRequired, 1, 512),
								"date_granularity":     stringEnumSchema[awstypes.TimeGranularity](attrOptional),
								"format_configuration": dateTimeFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimeFormatConfiguration.html
								"hierarchy_id":         stringLenBetweenSchema(attrOptional, 1, 512),
							},
						},
					},
					"numerical_dimension_field": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericalDimensionField.html
						Type:     schema.TypeList,
						MinItems: 1,
						MaxItems: 1,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"column":               columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
								"field_id":             stringLenBetweenSchema(attrRequired, 1, 512),
								"format_configuration": numberFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumberFormatConfiguration.html
								"hierarchy_id":         stringLenBetweenSchema(attrOptional, 1, 512),
							},
						},
					},
				},
			},
		},
	)
	return s
}

type meaureFieldSchemaIdentity measureFieldsSize

var measureFieldSchemaCache syncMap[meaureFieldSchemaIdentity, *schema.Schema]

func measureFieldSchema(maxItems measureFieldsSize) *schema.Schema {
	id := meaureFieldSchemaIdentity(maxItems)

	s, ok := measureFieldSchemaCache.Load(id)
	if ok {
		return s
	}

	// Use a separate `LoadOrStore` to avoid allocation if item is already in the cache
	// Use `LoadOrStore` instead of `Store` in case there is a race
	s, _ = measureFieldSchemaCache.LoadOrStore(
		id,
		&schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
			Type:     schema.TypeList,
			MinItems: 1,
			MaxItems: int(maxItems),
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"calculated_measure_field": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CalculatedMeasureField.html
						Type:     schema.TypeList,
						MinItems: 1,
						MaxItems: 1,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								names.AttrExpression: stringLenBetweenSchema(attrRequired, 1, 4096),
								"field_id":           stringLenBetweenSchema(attrRequired, 1, 512),
							},
						},
					},
					"categorical_measure_field": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CategoricalMeasureField.html
						Type:     schema.TypeList,
						MinItems: 1,
						MaxItems: 1,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"column":               columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
								"field_id":             stringLenBetweenSchema(attrRequired, 1, 512),
								"aggregation_function": stringEnumSchema[awstypes.CategoricalAggregationFunction](attrOptional),
								"format_configuration": stringFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_StringFormatConfiguration.html
							},
						},
					},
					"date_measure_field": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateMeasureField.html
						Type:     schema.TypeList,
						MinItems: 1,
						MaxItems: 1,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"column":               columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
								"field_id":             stringLenBetweenSchema(attrRequired, 1, 512),
								"aggregation_function": stringEnumSchema[awstypes.DateAggregationFunction](attrOptional),
								"format_configuration": dateTimeFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimeFormatConfiguration.html
							},
						},
					},
					"numerical_measure_field": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericalMeasureField.html
						Type:     schema.TypeList,
						MinItems: 1,
						MaxItems: 1,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"column":               columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
								"field_id":             stringLenBetweenSchema(attrRequired, 1, 512),
								"aggregation_function": numericalAggregationFunctionSchema(false), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericalAggregationFunction.html
								"format_configuration": numberFormatConfigurationSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumberFormatConfiguration.html
							},
						},
					},
				},
			},
		},
	)
	return s
}

func expandDimensionFields(tfList []any) []awstypes.DimensionField {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.DimensionField

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandDimensionInternal(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandDimensionInternal(tfMap map[string]any) *awstypes.DimensionField {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DimensionField{}

	if v, ok := tfMap["categorical_dimension_field"].([]any); ok && len(v) > 0 {
		apiObject.CategoricalDimensionField = expandCategoricalDimensionField(v)
	}
	if v, ok := tfMap["date_dimension_field"].([]any); ok && len(v) > 0 {
		apiObject.DateDimensionField = expandDateDimensionField(v)
	}
	if v, ok := tfMap["numerical_dimension_field"].([]any); ok && len(v) > 0 {
		apiObject.NumericalDimensionField = expandNumericalDimensionField(v)
	}

	return apiObject
}

func expandDimensionField(tfList []any) *awstypes.DimensionField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	if tfMap == nil {
		return nil
	}

	return expandDimensionInternal(tfMap)
}

func expandCategoricalDimensionField(tfList []any) *awstypes.CategoricalDimensionField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.CategoricalDimensionField{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["hierarchy_id"].(string); ok && v != "" {
		apiObject.HierarchyId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]any); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.FormatConfiguration = expandStringFormatConfiguration(v)
	}

	return apiObject
}

func expandDateDimensionField(tfList []any) *awstypes.DateDimensionField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DateDimensionField{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["hierarchy_id"].(string); ok && v != "" {
		apiObject.HierarchyId = aws.String(v)
	}
	if v, ok := tfMap["date_granularity"].(string); ok && v != "" {
		apiObject.DateGranularity = awstypes.TimeGranularity(v)
	}
	if v, ok := tfMap["column"].([]any); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.FormatConfiguration = expandDateTimeFormatConfiguration(v)
	}

	return apiObject
}

func expandNumericalDimensionField(tfList []any) *awstypes.NumericalDimensionField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.NumericalDimensionField{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["hierarchy_id"].(string); ok && v != "" {
		apiObject.HierarchyId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]any); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.FormatConfiguration = expandNumberFormatConfiguration(v)
	}

	return apiObject
}

func expandMeasureFields(tfList []any) []awstypes.MeasureField {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.MeasureField

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandMeasureFieldInternal(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandMeasureFieldInternal(tfMap map[string]any) *awstypes.MeasureField {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.MeasureField{}

	if v, ok := tfMap["calculated_measure_field"].([]any); ok && len(v) > 0 {
		apiObject.CalculatedMeasureField = expandCalculatedMeasureField(v)
	}
	if v, ok := tfMap["categorical_measure_field"].([]any); ok && len(v) > 0 {
		apiObject.CategoricalMeasureField = expandCategoricalMeasureField(v)
	}
	if v, ok := tfMap["date_measure_field"].([]any); ok && len(v) > 0 {
		apiObject.DateMeasureField = expandDateMeasureField(v)
	}
	if v, ok := tfMap["numerical_measure_field"].([]any); ok && len(v) > 0 {
		apiObject.NumericalMeasureField = expandNumericalMeasureField(v)
	}

	return apiObject
}

func expandMeasureField(tfList []any) *awstypes.MeasureField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	return expandMeasureFieldInternal(tfMap)
}

func expandCalculatedMeasureField(tfList []any) *awstypes.CalculatedMeasureField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.CalculatedMeasureField{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
		apiObject.Expression = aws.String(v)
	}

	return apiObject
}

func expandCategoricalMeasureField(tfList []any) *awstypes.CategoricalMeasureField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.CategoricalMeasureField{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["aggregation_function"].(string); ok && v != "" {
		apiObject.AggregationFunction = awstypes.CategoricalAggregationFunction(v)
	}
	if v, ok := tfMap["column"].([]any); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.FormatConfiguration = expandStringFormatConfiguration(v)
	}

	return apiObject
}

func expandDateMeasureField(tfList []any) *awstypes.DateMeasureField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DateMeasureField{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["aggregation_function"].(string); ok && v != "" {
		apiObject.AggregationFunction = awstypes.DateAggregationFunction(v)
	}
	if v, ok := tfMap["column"].([]any); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.FormatConfiguration = expandDateTimeFormatConfiguration(v)
	}

	return apiObject
}

func expandNumericalMeasureField(tfList []any) *awstypes.NumericalMeasureField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.NumericalMeasureField{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]any); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["aggregation_function"].([]any); ok && len(v) > 0 {
		apiObject.AggregationFunction = expandNumericalAggregationFunction(v)
	}
	if v, ok := tfMap["format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.FormatConfiguration = expandNumberFormatConfiguration(v)
	}

	return apiObject
}

func flattenDimensionField(apiObject *awstypes.DimensionField) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CategoricalDimensionField != nil {
		tfMap["categorical_dimension_field"] = flattenCategoricalDimensionField(apiObject.CategoricalDimensionField)
	}
	if apiObject.DateDimensionField != nil {
		tfMap["date_dimension_field"] = flattenDateDimensionField(apiObject.DateDimensionField)
	}
	if apiObject.NumericalDimensionField != nil {
		tfMap["numerical_dimension_field"] = flattenNumericalDimensionField(apiObject.NumericalDimensionField)
	}

	return []any{tfMap}
}

func flattenDimensionFields(apiObjects []awstypes.DimensionField) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.CategoricalDimensionField != nil {
			tfMap["categorical_dimension_field"] = flattenCategoricalDimensionField(apiObject.CategoricalDimensionField)
		}
		if apiObject.DateDimensionField != nil {
			tfMap["date_dimension_field"] = flattenDateDimensionField(apiObject.DateDimensionField)
		}
		if apiObject.NumericalDimensionField != nil {
			tfMap["numerical_dimension_field"] = flattenNumericalDimensionField(apiObject.NumericalDimensionField)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenCategoricalDimensionField(apiObject *awstypes.CategoricalDimensionField) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.FormatConfiguration != nil {
		tfMap["format_configuration"] = flattenStringFormatConfiguration(apiObject.FormatConfiguration)
	}
	if apiObject.HierarchyId != nil {
		tfMap["hierarchy_id"] = aws.ToString(apiObject.HierarchyId)
	}

	return []any{tfMap}
}

func flattenDateDimensionField(apiObject *awstypes.DateDimensionField) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	tfMap["date_granularity"] = apiObject.DateGranularity
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.FormatConfiguration != nil {
		tfMap["format_configuration"] = flattenDateTimeFormatConfiguration(apiObject.FormatConfiguration)
	}
	if apiObject.HierarchyId != nil {
		tfMap["hierarchy_id"] = aws.ToString(apiObject.HierarchyId)
	}

	return []any{tfMap}
}

func flattenNumericalDimensionField(apiObject *awstypes.NumericalDimensionField) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.FormatConfiguration != nil {
		tfMap["format_configuration"] = flattenNumberFormatConfiguration(apiObject.FormatConfiguration)
	}
	if apiObject.HierarchyId != nil {
		tfMap["hierarchy_id"] = aws.ToString(apiObject.HierarchyId)
	}

	return []any{tfMap}
}

func flattenMeasureField(apiObject *awstypes.MeasureField) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CalculatedMeasureField != nil {
		tfMap["calculated_measure_field"] = flattenCalculatedMeasureField(apiObject.CalculatedMeasureField)
	}
	if apiObject.CategoricalMeasureField != nil {
		tfMap["categorical_measure_field"] = flattenCategoricalMeasureField(apiObject.CategoricalMeasureField)
	}
	if apiObject.DateMeasureField != nil {
		tfMap["date_measure_field"] = flattenDateMeasureField(apiObject.DateMeasureField)
	}
	if apiObject.NumericalMeasureField != nil {
		tfMap["numerical_measure_field"] = flattenNumericalMeasureField(apiObject.NumericalMeasureField)
	}

	return []any{tfMap}
}

func flattenMeasureFields(apiObjects []awstypes.MeasureField) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.CalculatedMeasureField != nil {
			tfMap["calculated_measure_field"] = flattenCalculatedMeasureField(apiObject.CalculatedMeasureField)
		}
		if apiObject.CategoricalMeasureField != nil {
			tfMap["categorical_measure_field"] = flattenCategoricalMeasureField(apiObject.CategoricalMeasureField)
		}
		if apiObject.DateMeasureField != nil {
			tfMap["date_measure_field"] = flattenDateMeasureField(apiObject.DateMeasureField)
		}
		if apiObject.NumericalMeasureField != nil {
			tfMap["numerical_measure_field"] = flattenNumericalMeasureField(apiObject.NumericalMeasureField)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenCalculatedMeasureField(apiObject *awstypes.CalculatedMeasureField) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.Expression != nil {
		tfMap[names.AttrExpression] = aws.ToString(apiObject.Expression)
	}

	return []any{tfMap}
}

func flattenCategoricalMeasureField(apiObject *awstypes.CategoricalMeasureField) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["aggregation_function"] = apiObject.AggregationFunction
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.FormatConfiguration != nil {
		tfMap["format_configuration"] = flattenStringFormatConfiguration(apiObject.FormatConfiguration)
	}

	return []any{tfMap}
}

func flattenDateMeasureField(apiObject *awstypes.DateMeasureField) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap["aggregation_function"] = apiObject.AggregationFunction
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.FormatConfiguration != nil {
		tfMap["format_configuration"] = flattenDateTimeFormatConfiguration(apiObject.FormatConfiguration)
	}

	return []any{tfMap}
}

func flattenNumericalMeasureField(apiObject *awstypes.NumericalMeasureField) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.AggregationFunction != nil {
		tfMap["aggregation_function"] = flattenNumericalAggregationFunction(apiObject.AggregationFunction)
	}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.ToString(apiObject.FieldId)
	}
	if apiObject.FormatConfiguration != nil {
		tfMap["format_configuration"] = flattenNumberFormatConfiguration(apiObject.FormatConfiguration)
	}

	return []any{tfMap}
}
