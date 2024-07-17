// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const measureFieldsMaxItems5 = 5
const measureFieldsMaxItems20 = 20
const measureFieldsMaxItems40 = 40
const measureFieldsMaxItems200 = 200
const dimensionsFieldMaxItems10 = 10
const dimensionsFieldMaxItems40 = 40
const dimensionsFieldMaxItems200 = 200

func dimensionFieldSchema(maxItems int) *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: maxItems,
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
							"field_id":             stringSchema(true, validation.StringLenBetween(1, 512)),
							"format_configuration": stringFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_StringFormatConfiguration.html
							"hierarchy_id":         stringSchema(false, validation.StringLenBetween(1, 512)),
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
							"field_id":             stringSchema(true, validation.StringLenBetween(1, 512)),
							"date_granularity":     stringSchema(false, validation.StringInSlice(quicksight.TimeGranularity_Values(), false)),
							"format_configuration": dateTimeFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimeFormatConfiguration.html
							"hierarchy_id":         stringSchema(false, validation.StringLenBetween(1, 512)),
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
							"field_id":             stringSchema(true, validation.StringLenBetween(1, 512)),
							"format_configuration": numberFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumberFormatConfiguration.html
							"hierarchy_id":         stringSchema(false, validation.StringLenBetween(1, 512)),
						},
					},
				},
			},
		},
	}
}

func measureFieldSchema(maxItems int) *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: maxItems,
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
							names.AttrExpression: stringSchema(true, validation.StringLenBetween(1, 4096)),
							"field_id":           stringSchema(true, validation.StringLenBetween(1, 512)),
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
							"field_id":             stringSchema(true, validation.StringLenBetween(1, 512)),
							"aggregation_function": stringSchema(false, validation.StringInSlice(quicksight.CategoricalAggregationFunction_Values(), false)),
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
							"field_id":             stringSchema(true, validation.StringLenBetween(1, 512)),
							"aggregation_function": stringSchema(false, validation.StringInSlice(quicksight.DateAggregationFunction_Values(), false)),
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
							"field_id":             stringSchema(true, validation.StringLenBetween(1, 512)),
							"aggregation_function": numericalAggregationFunctionSchema(false), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericalAggregationFunction.html
							"format_configuration": numberFormatConfigurationSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumberFormatConfiguration.html
						},
					},
				},
			},
		},
	}
}

func expandDimensionFields(tfList []interface{}) []*quicksight.DimensionField {
	if len(tfList) == 0 {
		return nil
	}

	var fields []*quicksight.DimensionField
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		field := expandDimensionInternal(tfMap)
		if field == nil {
			continue
		}

		fields = append(fields, field)
	}

	return fields
}

func expandDimensionInternal(tfMap map[string]interface{}) *quicksight.DimensionField {
	if tfMap == nil {
		return nil
	}

	field := &quicksight.DimensionField{}

	if v, ok := tfMap["categorical_dimension_field"].([]interface{}); ok && len(v) > 0 {
		field.CategoricalDimensionField = expandCategoricalDimensionField(v)
	}
	if v, ok := tfMap["date_dimension_field"].([]interface{}); ok && len(v) > 0 {
		field.DateDimensionField = expandDateDimensionField(v)
	}
	if v, ok := tfMap["numerical_dimension_field"].([]interface{}); ok && len(v) > 0 {
		field.NumericalDimensionField = expandNumericalDimensionField(v)
	}

	return field
}

func expandDimensionField(tfList []interface{}) *quicksight.DimensionField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	if tfMap == nil {
		return nil
	}

	return expandDimensionInternal(tfMap)
}

func expandCategoricalDimensionField(tfList []interface{}) *quicksight.CategoricalDimensionField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	field := &quicksight.CategoricalDimensionField{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		field.FieldId = aws.String(v)
	}
	if v, ok := tfMap["hierarchy_id"].(string); ok && v != "" {
		field.HierarchyId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		field.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["format_configuration"].([]interface{}); ok && len(v) > 0 {
		field.FormatConfiguration = expandStringFormatConfiguration(v)
	}

	return field
}

func expandDateDimensionField(tfList []interface{}) *quicksight.DateDimensionField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	field := &quicksight.DateDimensionField{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		field.FieldId = aws.String(v)
	}
	if v, ok := tfMap["hierarchy_id"].(string); ok && v != "" {
		field.HierarchyId = aws.String(v)
	}
	if v, ok := tfMap["date_granularity"].(string); ok && v != "" {
		field.DateGranularity = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		field.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["format_configuration"].([]interface{}); ok && len(v) > 0 {
		field.FormatConfiguration = expandDateTimeFormatConfiguration(v)
	}

	return field
}

func expandNumericalDimensionField(tfList []interface{}) *quicksight.NumericalDimensionField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	field := &quicksight.NumericalDimensionField{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		field.FieldId = aws.String(v)
	}
	if v, ok := tfMap["hierarchy_id"].(string); ok && v != "" {
		field.HierarchyId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		field.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["format_configuration"].([]interface{}); ok && len(v) > 0 {
		field.FormatConfiguration = expandNumberFormatConfiguration(v)
	}

	return field
}

func expandMeasureFields(tfList []interface{}) []*quicksight.MeasureField {
	if len(tfList) == 0 {
		return nil
	}

	var fields []*quicksight.MeasureField
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		field := expandMeasureFieldInternal(tfMap)
		if field == nil {
			continue
		}

		fields = append(fields, field)
	}

	return fields
}

func expandMeasureFieldInternal(tfMap map[string]interface{}) *quicksight.MeasureField {
	if tfMap == nil {
		return nil
	}

	field := &quicksight.MeasureField{}

	if v, ok := tfMap["calculated_measure_field"].([]interface{}); ok && len(v) > 0 {
		field.CalculatedMeasureField = expandCalculatedMeasureField(v)
	}
	if v, ok := tfMap["categorical_measure_field"].([]interface{}); ok && len(v) > 0 {
		field.CategoricalMeasureField = expandCategoricalMeasureField(v)
	}
	if v, ok := tfMap["date_measure_field"].([]interface{}); ok && len(v) > 0 {
		field.DateMeasureField = expandDateMeasureField(v)
	}
	if v, ok := tfMap["numerical_measure_field"].([]interface{}); ok && len(v) > 0 {
		field.NumericalMeasureField = expandNumericalMeasureField(v)
	}

	return field
}

func expandMeasureField(tfList []interface{}) *quicksight.MeasureField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	return expandMeasureFieldInternal(tfMap)
}

func expandCalculatedMeasureField(tfList []interface{}) *quicksight.CalculatedMeasureField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	field := &quicksight.CalculatedMeasureField{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		field.FieldId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
		field.Expression = aws.String(v)
	}

	return field
}

func expandCategoricalMeasureField(tfList []interface{}) *quicksight.CategoricalMeasureField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	field := &quicksight.CategoricalMeasureField{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		field.FieldId = aws.String(v)
	}
	if v, ok := tfMap["aggregation_function"].(string); ok && v != "" {
		field.AggregationFunction = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		field.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["format_configuration"].([]interface{}); ok && len(v) > 0 {
		field.FormatConfiguration = expandStringFormatConfiguration(v)
	}

	return field
}

func expandDateMeasureField(tfList []interface{}) *quicksight.DateMeasureField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	field := &quicksight.DateMeasureField{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		field.FieldId = aws.String(v)
	}
	if v, ok := tfMap["aggregation_function"].(string); ok && v != "" {
		field.AggregationFunction = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		field.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["format_configuration"].([]interface{}); ok && len(v) > 0 {
		field.FormatConfiguration = expandDateTimeFormatConfiguration(v)
	}

	return field
}

func expandNumericalMeasureField(tfList []interface{}) *quicksight.NumericalMeasureField {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	field := &quicksight.NumericalMeasureField{}

	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		field.FieldId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		field.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["aggregation_function"].([]interface{}); ok && len(v) > 0 {
		field.AggregationFunction = expandNumericalAggregationFunction(v)
	}
	if v, ok := tfMap["format_configuration"].([]interface{}); ok && len(v) > 0 {
		field.FormatConfiguration = expandNumberFormatConfiguration(v)
	}

	return field
}

func flattenDimensionField(apiObject *quicksight.DimensionField) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CategoricalDimensionField != nil {
		tfMap["categorical_dimension_field"] = flattenCategoricalDimensionField(apiObject.CategoricalDimensionField)
	}
	if apiObject.DateDimensionField != nil {
		tfMap["date_dimension_field"] = flattenDateDimensionField(apiObject.DateDimensionField)
	}
	if apiObject.NumericalDimensionField != nil {
		tfMap["numerical_dimension_field"] = flattenNumericalDimensionField(apiObject.NumericalDimensionField)
	}

	return []interface{}{tfMap}
}

func flattenDimensionFields(apiObject []*quicksight.DimensionField) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.CategoricalDimensionField != nil {
			tfMap["categorical_dimension_field"] = flattenCategoricalDimensionField(config.CategoricalDimensionField)
		}
		if config.DateDimensionField != nil {
			tfMap["date_dimension_field"] = flattenDateDimensionField(config.DateDimensionField)
		}
		if config.NumericalDimensionField != nil {
			tfMap["numerical_dimension_field"] = flattenNumericalDimensionField(config.NumericalDimensionField)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenCategoricalDimensionField(apiObject *quicksight.CategoricalDimensionField) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.StringValue(apiObject.FieldId)
	}
	if apiObject.FormatConfiguration != nil {
		tfMap["format_configuration"] = flattenStringFormatConfiguration(apiObject.FormatConfiguration)
	}
	if apiObject.HierarchyId != nil {
		tfMap["hierarchy_id"] = aws.StringValue(apiObject.HierarchyId)
	}

	return []interface{}{tfMap}
}

func flattenDateDimensionField(apiObject *quicksight.DateDimensionField) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.DateGranularity != nil {
		tfMap["date_granularity"] = aws.StringValue(apiObject.DateGranularity)
	}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.StringValue(apiObject.FieldId)
	}
	if apiObject.FormatConfiguration != nil {
		tfMap["format_configuration"] = flattenDateTimeFormatConfiguration(apiObject.FormatConfiguration)
	}
	if apiObject.HierarchyId != nil {
		tfMap["hierarchy_id"] = aws.StringValue(apiObject.HierarchyId)
	}

	return []interface{}{tfMap}
}

func flattenNumericalDimensionField(apiObject *quicksight.NumericalDimensionField) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.StringValue(apiObject.FieldId)
	}
	if apiObject.FormatConfiguration != nil {
		tfMap["format_configuration"] = flattenNumberFormatConfiguration(apiObject.FormatConfiguration)
	}
	if apiObject.HierarchyId != nil {
		tfMap["hierarchy_id"] = aws.StringValue(apiObject.HierarchyId)
	}

	return []interface{}{tfMap}
}

func flattenMeasureField(apiObject *quicksight.MeasureField) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
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

	return []interface{}{tfMap}
}

func flattenMeasureFields(apiObject []*quicksight.MeasureField) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.CalculatedMeasureField != nil {
			tfMap["calculated_measure_field"] = flattenCalculatedMeasureField(config.CalculatedMeasureField)
		}
		if config.CategoricalMeasureField != nil {
			tfMap["categorical_measure_field"] = flattenCategoricalMeasureField(config.CategoricalMeasureField)
		}
		if config.DateMeasureField != nil {
			tfMap["date_measure_field"] = flattenDateMeasureField(config.DateMeasureField)
		}
		if config.NumericalMeasureField != nil {
			tfMap["numerical_measure_field"] = flattenNumericalMeasureField(config.NumericalMeasureField)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenCalculatedMeasureField(apiObject *quicksight.CalculatedMeasureField) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.StringValue(apiObject.FieldId)
	}
	if apiObject.Expression != nil {
		tfMap[names.AttrExpression] = aws.StringValue(apiObject.Expression)
	}

	return []interface{}{tfMap}
}

func flattenCategoricalMeasureField(apiObject *quicksight.CategoricalMeasureField) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AggregationFunction != nil {
		tfMap["aggregation_function"] = aws.StringValue(apiObject.AggregationFunction)
	}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.StringValue(apiObject.FieldId)
	}
	if apiObject.FormatConfiguration != nil {
		tfMap["format_configuration"] = flattenStringFormatConfiguration(apiObject.FormatConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenDateMeasureField(apiObject *quicksight.DateMeasureField) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AggregationFunction != nil {
		tfMap["aggregation_function"] = aws.StringValue(apiObject.AggregationFunction)
	}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.StringValue(apiObject.FieldId)
	}
	if apiObject.FormatConfiguration != nil {
		tfMap["format_configuration"] = flattenDateTimeFormatConfiguration(apiObject.FormatConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenNumericalMeasureField(apiObject *quicksight.NumericalMeasureField) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.AggregationFunction != nil {
		tfMap["aggregation_function"] = flattenNumericalAggregationFunction(apiObject.AggregationFunction)
	}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.FieldId != nil {
		tfMap["field_id"] = aws.StringValue(apiObject.FieldId)
	}
	if apiObject.FormatConfiguration != nil {
		tfMap["format_configuration"] = flattenNumberFormatConfiguration(apiObject.FormatConfiguration)
	}

	return []interface{}{tfMap}
}
