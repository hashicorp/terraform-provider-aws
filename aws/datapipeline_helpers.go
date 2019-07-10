package aws

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"

	"github.com/hashicorp/terraform/helper/schema"
)

func expandDataPipelineParameterObject(rawParameterObject map[string]interface{}) (*datapipeline.ParameterObject, error) {
	parameterObject := &datapipeline.ParameterObject{}
	parameterAttributeList := []*datapipeline.ParameterAttribute{}

	if val, ok := rawParameterObject["id"].(string); ok && val != "" {
		parameterObject.Id = aws.String(val)
	}

	if val, ok := rawParameterObject["description"].(string); ok && val != "" {
		p := &datapipeline.ParameterAttribute{
			Key:         aws.String("description"),
			StringValue: aws.String(val),
		}
		parameterAttributeList = append(parameterAttributeList, p)
	}

	if val, ok := rawParameterObject["type"].(string); ok && val != "" {
		p := &datapipeline.ParameterAttribute{
			Key:         aws.String("type"),
			StringValue: aws.String(val),
		}
		parameterAttributeList = append(parameterAttributeList, p)
	}

	if val, ok := rawParameterObject["optional"].(bool); ok {
		p := &datapipeline.ParameterAttribute{
			Key:         aws.String("optional"),
			StringValue: aws.String(strconv.FormatBool(val)),
		}
		parameterAttributeList = append(parameterAttributeList, p)
	}

	if val, ok := rawParameterObject["allowed_values"].([]string); ok && len(val) != 0 {
		for _, v := range val {
			p := &datapipeline.ParameterAttribute{
				Key:         aws.String("allowedValues"),
				StringValue: aws.String(v),
			}
			parameterAttributeList = append(parameterAttributeList, p)
		}
	}

	if val, ok := rawParameterObject["default"].(string); ok && val != "" {
		p := &datapipeline.ParameterAttribute{
			Key:         aws.String("default"),
			StringValue: aws.String(val),
		}
		parameterAttributeList = append(parameterAttributeList, p)
	}

	if val, ok := rawParameterObject["is_array"].(bool); ok {
		p := &datapipeline.ParameterAttribute{
			Key:         aws.String("isArray"),
			StringValue: aws.String(strconv.FormatBool(val)),
		}
		parameterAttributeList = append(parameterAttributeList, p)
	}

	parameterObject.Attributes = parameterAttributeList

	if err := parameterObject.Validate(); err != nil {
		return nil, fmt.Errorf("Error validate parameter object: %s", err)
	}

	return parameterObject, nil
}

func expandDataPipelineParameterValue(rawParameterValue map[string]interface{}) (*datapipeline.ParameterValue, error) {
	parameterValue := &datapipeline.ParameterValue{}

	if val, ok := rawParameterValue["id"].(string); ok && val != "" {
		parameterValue.Id = aws.String(val)
	}

	if val, ok := rawParameterValue["string_value"].(string); ok && val != "" {
		parameterValue.StringValue = aws.String(val)
	}

	if err := parameterValue.Validate(); err != nil {
		return nil, fmt.Errorf("Error validate parameter value: %s", err)
	}

	return parameterValue, nil
}

func expandDataPipelineDefaultPipelineObject(l []interface{}) (*datapipeline.PipelineObject, error) {
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}

	m := l[0].(map[string]interface{})

	pipelineObject := &datapipeline.PipelineObject{
		Id:   aws.String("Default"),
		Name: aws.String("Default"),
	}
	fieldList := []*datapipeline.Field{}

	typeField := &datapipeline.Field{
		Key:         aws.String("type"),
		StringValue: aws.String("Default"),
	}
	fieldList = append(fieldList, typeField)

	if val, ok := m["schedule_type"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("scheduleType"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := m["failure_and_rerun_mode"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("failureAndRerunMode"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := m["pipeline_log_uri"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("pipelineLogUri"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := m["role"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("role"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := m["resource_role"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("resourceRole"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := m["schedule"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:      aws.String("schedule"),
			RefValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	pipelineObject.Fields = fieldList

	err := pipelineObject.Validate()
	if err != nil {
		return nil, fmt.Errorf("Error validate default pipeline object: %s", err)
	}

	return pipelineObject, nil
}

func expandDataPipelinePipelineObject(pipelineObjectType string, rawPipelineObject map[string]interface{}) (*datapipeline.PipelineObject, error) {
	pipelineObject := &datapipeline.PipelineObject{
		Id:   aws.String(rawPipelineObject["id"].(string)),
		Name: aws.String(rawPipelineObject["name"].(string)),
	}
	fieldList := []*datapipeline.Field{}

	typeField := &datapipeline.Field{
		Key:         aws.String("type"),
		StringValue: aws.String(pipelineObjectType),
	}
	fieldList = append(fieldList, typeField)

	if v, ok := rawPipelineObject["period"]; ok && v.(string) != "" {
		f := &datapipeline.Field{
			Key:         aws.String("period"),
			StringValue: aws.String(v.(string)),
		}
		fieldList = append(fieldList, f)
	}

	if v, ok := rawPipelineObject["start_at"]; ok && v.(string) != "" {
		f := &datapipeline.Field{
			Key:         aws.String("startAt"),
			StringValue: aws.String(v.(string)),
		}
		fieldList = append(fieldList, f)
	}

	if v, ok := rawPipelineObject["start_date_time"]; ok && v.(string) != "" {
		f := &datapipeline.Field{
			Key:         aws.String("startDateTime"),
			StringValue: aws.String(v.(string)),
		}
		fieldList = append(fieldList, f)
	}

	if v, ok := rawPipelineObject["end_date_time"]; ok && v.(string) != "" {
		f := &datapipeline.Field{
			Key:         aws.String("endDateTime"),
			StringValue: aws.String(v.(string)),
		}
		fieldList = append(fieldList, f)
	}

	if v, ok := rawPipelineObject["occurrences"]; ok && v.(int) != 0 {
		f := &datapipeline.Field{
			Key:         aws.String("occurrences"),
			StringValue: aws.String(strconv.Itoa(v.(int))),
		}
		fieldList = append(fieldList, f)
	}

	if v, ok := rawPipelineObject["parent"]; ok && v.(string) != "" {
		f := &datapipeline.Field{
			Key:      aws.String("parent"),
			RefValue: aws.String(v.(string)),
		}
		fieldList = append(fieldList, f)
	}

	pipelineObject.Fields = fieldList

	err := pipelineObject.Validate()
	if err != nil {
		return nil, fmt.Errorf("Error validate pipeline object: %s", err)
	}
	return pipelineObject, nil
}

func flattenDataPipelineParameterValues(objects []*datapipeline.ParameterValue) []map[string]interface{} {
	parameterValues := make([]map[string]interface{}, 0, len(objects))

	for _, object := range objects {
		obj := make(map[string]interface{})
		obj["id"] = aws.StringValue(object.Id)
		obj["string_value"] = aws.StringValue(object.StringValue)

		parameterValues = append(parameterValues, obj)
	}
	return parameterValues
}

func flattenDataPipelinePipelineObjects(d *schema.ResourceData, pipelineObjects []*datapipeline.PipelineObject) error {
	var schedules []map[string]interface{}

	for _, pipelineObject := range pipelineObjects {
		for _, field := range pipelineObject.Fields {
			if aws.StringValue(field.Key) == "type" {
				switch aws.StringValue(field.StringValue) {
				case "Default":
					if err := d.Set("default", flattenDataPipelineDefaultPipelineObject(pipelineObject)); err != nil {
						return fmt.Errorf("Error setting default object: %s", err)
					}
				case "Schedule":
					scheduleObject, err := flattenDataPipelinePipelineObject(pipelineObject)
					if err != nil {
						return fmt.Errorf("Error setting schedule object: %s", err)
					}
					schedules = append(schedules, scheduleObject)
				}
			}
		}
	}

	if err := d.Set("schedule", schedules); err != nil {
		return fmt.Errorf("Error setting default object: %s", err)
	}

	return nil
}

func flattenDataPipelineParameterObjects(objects []*datapipeline.ParameterObject) []map[string]interface{} {
	parameterObjects := make([]map[string]interface{}, 0, len(objects))
	for _, object := range objects {
		obj := make(map[string]interface{})

		obj["id"] = aws.StringValue(object.Id)
		for _, attribute := range object.Attributes {
			switch aws.StringValue(attribute.Key) {
			case "type":
				obj["type"] = aws.StringValue(attribute.StringValue)
			case "description":
				obj["description"] = aws.StringValue(attribute.StringValue)
			case "optional":
				b, _ := strconv.ParseBool(aws.StringValue(attribute.StringValue))
				obj["optional"] = b
			case "allowedValues":
				obj["allowed_values"] = aws.StringValue(attribute.StringValue)
			case "default":
				obj["default"] = aws.StringValue(attribute.StringValue)
			case "isArray":
				b, _ := strconv.ParseBool(aws.StringValue(attribute.StringValue))
				obj["is_array"] = b
			}
		}
		parameterObjects = append(parameterObjects, obj)
	}

	return parameterObjects
}

func flattenDataPipelineDefaultPipelineObject(object *datapipeline.PipelineObject) []map[string]interface{} {
	pipelineObject := make(map[string]interface{})
	for _, field := range object.Fields {
		switch aws.StringValue(field.Key) {
		case "scheduleType":
			pipelineObject["schedule_type"] = aws.StringValue(field.StringValue)
		case "failureAndRerunMode":
			pipelineObject["failure_and_rerun_mode"] = aws.StringValue(field.StringValue)
		case "pipelineLogUri":
			pipelineObject["pipeline_log_uri"] = aws.StringValue(field.StringValue)
		case "role":
			pipelineObject["role"] = aws.StringValue(field.StringValue)
		case "resourceRole":
			pipelineObject["resource_role"] = aws.StringValue(field.StringValue)
		case "schedule":
			pipelineObject["schedule"] = aws.StringValue(field.RefValue)
		}
	}

	return []map[string]interface{}{pipelineObject}
}

func flattenDataPipelinePipelineObject(object *datapipeline.PipelineObject) (map[string]interface{}, error) {
	pipelineObject := make(map[string]interface{})

	pipelineObject["id"] = aws.StringValue(object.Id)
	pipelineObject["name"] = aws.StringValue(object.Name)

	for _, field := range object.Fields {
		switch aws.StringValue(field.Key) {
		case "period":
			pipelineObject["period"] = aws.StringValue(field.StringValue)
		case "startAt":
			pipelineObject["start_at"] = aws.StringValue(field.StringValue)
		case "startDateTime":
			pipelineObject["start_date_time"] = aws.StringValue(field.StringValue)
		case "endDateTime":
			pipelineObject["end_date_time"] = aws.StringValue(field.StringValue)
		case "occurrences":
			val, err := strconv.Atoi(aws.StringValue(field.StringValue))
			if err != nil {
				return nil, err
			}
			pipelineObject["occurrences"] = val
		}
	}

	return pipelineObject, nil
}
