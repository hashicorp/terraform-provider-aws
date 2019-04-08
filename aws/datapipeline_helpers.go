package aws

import (
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"
)

func buildDefaultPipelineObject(rawDefault map[string]interface{}) (*datapipeline.PipelineObject, error) {
	pipelineObject := &datapipeline.PipelineObject{}
	fieldList := []*datapipeline.Field{}

	pipelineObject.Id = aws.String("Default")
	pipelineObject.Name = aws.String("Default")

	typeField := &datapipeline.Field{
		Key:         aws.String("type"),
		StringValue: aws.String("Default"),
	}
	fieldList = append(fieldList, typeField)

	if val, ok := rawDefault["schedule_type"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("scheduleType"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawDefault["failure_and_rerun_mode"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("failureAndRerunMode"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawDefault["pipeline_log_uri"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("pipelineLogUri"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawDefault["role"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("role"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawDefault["resource_role"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("resourceRole"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawDefault["schedule"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:      aws.String("schedule"),
			RefValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	pipelineObject.Fields = fieldList

	err := pipelineObject.Validate()
	if err != nil {
		return nil, err
	}

	return pipelineObject, nil
}

func buildCommonPipelineObject(pipelineObjectType string, rawPipelineObject map[string]interface{}) (*datapipeline.PipelineObject, error) {
	pipelineObject := &datapipeline.PipelineObject{}
	fieldList := []*datapipeline.Field{}

	if val, ok := rawPipelineObject["id"].(string); ok && val != "" {
		pipelineObject.Id = aws.String(val)
	}

	if val, ok := rawPipelineObject["name"].(string); ok && val != "" {
		pipelineObject.Name = aws.String(val)
	}

	typeField := &datapipeline.Field{
		Key:         aws.String("type"),
		StringValue: aws.String(pipelineObjectType),
	}
	fieldList = append(fieldList, typeField)

	if val, ok := rawPipelineObject["associate_public_ip_address"].(bool); ok && val {
		f := &datapipeline.Field{
			Key:         aws.String("associatePublicIpAddress"),
			StringValue: aws.String(strconv.FormatBool(val)),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["username"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("username"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["password"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("*password"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["rds_instance_id"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("rdsInstanceId"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["database_name"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("databaseName"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["jdbc_driver_jar_uri"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("jdbcDriverJarUri"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["jdbc_properties"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("jdbcProperties"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["schedule"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:      aws.String("schedule"),
			RefValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["runs_on"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:      aws.String("runsOn"),
			RefValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["worker_group"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("workerGroup"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["attempt_status"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("attemptStatus"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["attempt_timeout"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("attemptTimeout"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["create_table_sql"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("createTableSql"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["database"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:      aws.String("database"),
			RefValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["compression"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("compression"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["data_format"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:      aws.String("dataFormat"),
			RefValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["depends_on"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:      aws.String("dependsOn"),
			RefValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["directory_path"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("directoryPath"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["failure_and_rerun_mode"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("failureAndRerunMode"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["insert_query"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("insertQuery"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["availability_zone"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("availabilityZone"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["http_proxy"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:      aws.String("httpProxy"),
			RefValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["image_id"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("imageId"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["instance_type"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("instanceType"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["key_pair"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("keyPair"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["input"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:      aws.String("input"),
			RefValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["file_path"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("filePath"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["late_after_timeout"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("lateAfterTimeout"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["manifest_file_path"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("manifestFilePath"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["max_active_instances"].(int); ok && val != 0 {
		f := &datapipeline.Field{
			Key:         aws.String("maxActiveInstances"),
			StringValue: aws.String(strconv.Itoa(val)),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["maximum_retries"].(int); ok && val != 0 {
		f := &datapipeline.Field{
			Key:         aws.String("maximumRetries"),
			StringValue: aws.String(strconv.Itoa(val)),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["on_fail"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:      aws.String("onFail"),
			RefValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["on_late_action"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:      aws.String("onLateAction"),
			RefValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["on_success"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:      aws.String("onSuccess"),
			RefValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["output"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:      aws.String("output"),
			RefValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["period"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("period"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["start_at"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("startAt"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["start_date_time"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("startDateTime"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["end_date_time"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("endDateTime"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["occurrences"].(int); ok && val != 0 {
		f := &datapipeline.Field{
			Key:         aws.String("occurrences"),
			StringValue: aws.String(strconv.Itoa(val)),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["parent"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:      aws.String("parent"),
			RefValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["pipeline_log_uri"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("pipelineLogUri"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["precondition"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:      aws.String("precondition"),
			RefValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["report_progress_timeout"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("reportProgressTimeout"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["retry_delay"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("retryDelay"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["region"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("region"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["s3_encryption_type"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("s3EncryptionType"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["schedule_type"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("scheduleType"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["table"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("table"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["schema_name"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("schemaName"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["select_query"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("selectQuery"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["security_group_ids"].([]string); ok && len(val) != 0 {
		for _, v := range val {
			f := &datapipeline.Field{
				Key:         aws.String("securityGroupIds"),
				StringValue: aws.String(v),
			}
			fieldList = append(fieldList, f)
		}
	}

	if val, ok := rawPipelineObject["security_groups"].([]string); ok && len(val) != 0 {
		for _, v := range val {
			f := &datapipeline.Field{
				Key:         aws.String("securityGroups"),
				StringValue: aws.String(v),
			}
			fieldList = append(fieldList, f)
		}
	}

	if val, ok := rawPipelineObject["spot_bid_price"].(float64); ok && val > 0 && val < 20 {
		f := &datapipeline.Field{
			Key:         aws.String("spotBidPrice"),
			StringValue: aws.String(strconv.FormatFloat(val, 'f', -1, 64)),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["subnet_id"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("subnetId"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["terminate_after"].(string); ok && val != "" {
		f := &datapipeline.Field{
			Key:         aws.String("terminateAfter"),
			StringValue: aws.String(val),
		}
		fieldList = append(fieldList, f)
	}

	if val, ok := rawPipelineObject["use_on_demand_on_last_attempt"].(bool); ok {
		f := &datapipeline.Field{
			Key:         aws.String("useOnDemandOnLastAttempt"),
			StringValue: aws.String(strconv.FormatBool(val)),
		}
		fieldList = append(fieldList, f)
	}

	pipelineObject.Fields = fieldList
	err := pipelineObject.Validate()
	if err != nil {
		return nil, err
	}
	return pipelineObject, nil
}

func buildParameterObject(rawParameterObject map[string]interface{}) (*datapipeline.ParameterObject, error) {
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
	err := parameterObject.Validate()
	if err != nil {
		return nil, err
	}

	return parameterObject, nil
}

func flattenDefaultPipelineObject(object *datapipeline.PipelineObject) (map[string]interface{}, error) {
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
	return pipelineObject, nil
}

func flattenCommonPipelineObject(object *datapipeline.PipelineObject) (map[string]interface{}, error) {
	pipelineObject := make(map[string]interface{})
	var securityGroupIDObjects []string
	var securityGroupObjects []string

	pipelineObject["id"] = *object.Id
	pipelineObject["name"] = *object.Name

	for _, field := range object.Fields {
		switch aws.StringValue(field.Key) {
		case "attemptStatus":
			pipelineObject["attempt_status"] = aws.StringValue(field.StringValue)
		case "attemptTimeout":
			pipelineObject["attempt_timeout"] = aws.StringValue(field.StringValue)
		case "associatePublicIpAddress":
			val, _ := strconv.ParseBool(aws.StringValue(field.StringValue))
			pipelineObject["associate_public_ip_address"] = val
		case "username":
			pipelineObject["username"] = aws.StringValue(field.StringValue)
		case "*password":
			pipelineObject["password"] = aws.StringValue(field.StringValue)
		case "rdsInstanceId":
			pipelineObject["rds_instance_id"] = aws.StringValue(field.StringValue)
		case "databaseName":
			pipelineObject["database_name"] = aws.StringValue(field.StringValue)
		case "jdbcProperties":
			pipelineObject["jdbc_properties"] = aws.StringValue(field.StringValue)
		case "jdbcDriverJarUri":
			pipelineObject["jdbc_driver_jar_uri"] = aws.StringValue(field.StringValue)
		case "createTableSql":
			pipelineObject["create_table_sql"] = aws.StringValue(field.StringValue)
		case "database":
			pipelineObject["database"] = aws.StringValue(field.RefValue)
		case "insertQuery":
			pipelineObject["insert_query"] = aws.StringValue(field.StringValue)
		case "schemaName":
			pipelineObject["schema_name"] = aws.StringValue(field.StringValue)
		case "selectQuery":
			pipelineObject["select_query"] = aws.StringValue(field.StringValue)
		case "availabilityZone":
			pipelineObject["availability_zone"] = aws.StringValue(field.StringValue)
		case "httpProxy":
			pipelineObject["http_proxy"] = aws.StringValue(field.RefValue)
		case "imageId":
			pipelineObject["image_id"] = aws.StringValue(field.StringValue)
		case "instanceType":
			pipelineObject["instance_type"] = aws.StringValue(field.StringValue)
		case "keyPair":
			pipelineObject["key_pair"] = aws.StringValue(field.StringValue)
		case "maxActiveInstances":
			val, _ := strconv.Atoi(aws.StringValue(field.StringValue))
			pipelineObject["max_active_instances"] = val
		case "maximumRetries":
			val, _ := strconv.Atoi(aws.StringValue(field.StringValue))
			pipelineObject["maximum_retries"] = val
		case "onSuccess":
			pipelineObject["on_success"] = aws.StringValue(field.RefValue)
		case "onFail":
			pipelineObject["on_fail"] = aws.StringValue(field.RefValue)
		case "onLateAction":
			pipelineObject["on_late_action"] = aws.StringValue(field.RefValue)
		case "pipelineLogUri":
			pipelineObject["pipeline_log_uri"] = aws.StringValue(field.StringValue)
		case "region":
			pipelineObject["region"] = aws.StringValue(field.StringValue)
		case "scheduleType":
			pipelineObject["schedule_type"] = aws.StringValue(field.StringValue)
		case "schedule":
			pipelineObject["schedule"] = aws.StringValue(field.RefValue)
		case "securityGroupIds":
			securityGroupIDObjects = append(securityGroupIDObjects, aws.StringValue(field.StringValue))
		case "securityGroups":
			securityGroupObjects = append(securityGroupObjects, aws.StringValue(field.StringValue))
		case "subnetId":
			pipelineObject["subnet_id"] = aws.StringValue(field.StringValue)
		case "table":
			pipelineObject["table"] = aws.StringValue(field.StringValue)
		case "terminateAfter":
			pipelineObject["terminate_after"] = aws.StringValue(field.StringValue)
		case "useOnDemandOnLastAttempt":
			val, _ := strconv.ParseBool(aws.StringValue(field.StringValue))
			pipelineObject["use_on_demand_on_last_attempt"] = val
		case "compression":
			pipelineObject["compression"] = aws.StringValue(field.StringValue)
		case "dataFormat":
			pipelineObject["data_format"] = aws.StringValue(field.RefValue)
		case "dependsOn":
			pipelineObject["depends_on"] = aws.StringValue(field.RefValue)
		case "directoryPath":
			pipelineObject["directory_path"] = aws.StringValue(field.StringValue)
		case "failureAndRerunMode":
			pipelineObject["failure_and_rerun_mode"] = aws.StringValue(field.StringValue)
		case "input":
			pipelineObject["input"] = aws.StringValue(field.RefValue)
		case "filePath":
			pipelineObject["file_path"] = aws.StringValue(field.StringValue)
		case "output":
			pipelineObject["output"] = aws.StringValue(field.RefValue)
		case "parent":
			pipelineObject["parent"] = aws.StringValue(field.RefValue)
		case "lateAfterTimeout":
			pipelineObject["late_after_timeout"] = aws.StringValue(field.StringValue)
		case "manifestFilePath":
			pipelineObject["manifest_file_path"] = aws.StringValue(field.StringValue)
		case "precondition":
			pipelineObject["precondition"] = aws.StringValue(field.RefValue)
		case "reportProgressTimeout":
			pipelineObject["report_progress_timeout"] = aws.StringValue(field.StringValue)
		case "retryDelay":
			pipelineObject["retry_delay"] = aws.StringValue(field.StringValue)
		case "runsOn":
			pipelineObject["runs_on"] = aws.StringValue(field.RefValue)
		case "s3EncryptionType":
			pipelineObject["s3_encryption_type"] = aws.StringValue(field.StringValue)
		case "workerGroup":
			pipelineObject["worker_group"] = aws.StringValue(field.StringValue)
		case "period":
			pipelineObject["period"] = aws.StringValue(field.StringValue)
		case "spotBidPrice":
			val, _ := strconv.ParseFloat(aws.StringValue(field.StringValue), 64)
			pipelineObject["spot_bid_price"] = val
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
	if len(securityGroupIDObjects) != 0 {
		pipelineObject["security_group_ids"] = make([]string, 0, len(securityGroupIDObjects))
		pipelineObject["security_group_ids"] = securityGroupIDObjects
	}
	if len(securityGroupObjects) != 0 {
		pipelineObject["security_groups"] = make([]string, 0, len(securityGroupObjects))
		pipelineObject["security_groups"] = securityGroupObjects
	}

	return pipelineObject, nil
}
