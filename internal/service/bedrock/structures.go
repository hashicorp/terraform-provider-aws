package bedrock

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/bedrock"
)

func flattenTrainingMetrics(metrics *bedrock.TrainingMetrics) []map[string]interface{} {
	if metrics == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"training_loss": aws.Float64Value(metrics.TrainingLoss),
	}

	return []map[string]interface{}{m}
}

func flattenValidationDataConfig(config *bedrock.ValidationDataConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	l := make([]map[string]interface{}, 0, len(config.Validators))

	for _, validator := range config.Validators {
		m := map[string]interface{}{
			"validator": aws.StringValue(validator.S3Uri),
		}
		l = append(l, m)
	}

	return l
}

func flattenValidationMetrics(metrics []*bedrock.ValidatorMetric) []map[string]interface{} {
	if metrics == nil {
		return []map[string]interface{}{}
	}

	l := make([]map[string]interface{}, 0, len(metrics))

	for _, metric := range metrics {
		m := map[string]interface{}{
			"validation_loss": aws.Float64Value(metric.ValidationLoss),
		}
		l = append(l, m)
	}

	return l
}

func flattenCustomModelSummaries(models []*bedrock.CustomModelSummary) []map[string]interface{} {
	if len(models) == 0 {
		return []map[string]interface{}{}
	}

	l := make([]map[string]interface{}, 0, len(models))

	for _, model := range models {
		m := map[string]interface{}{
			"base_model_arn":  aws.StringValue(model.BaseModelArn),
			"base_model_name": aws.StringValue(model.BaseModelName),
			"model_arn":       aws.StringValue(model.ModelArn),
			"model_name":      aws.StringValue(model.ModelName),
			"creation_time":   aws.TimeValue(model.CreationTime).Format(time.RFC3339),
		}
		l = append(l, m)
	}

	return l
}

func flattenFoundationModelSummaries(models []*bedrock.FoundationModelSummary) []map[string]interface{} {
	if len(models) == 0 {
		return []map[string]interface{}{}
	}

	l := make([]map[string]interface{}, 0, len(models))

	for _, model := range models {
		m := map[string]interface{}{
			"model_arn":                    aws.StringValue(model.ModelArn),
			"model_id":                     aws.StringValue(model.ModelId),
			"model_name":                   aws.StringValue(model.ModelName),
			"provider_name":                aws.StringValue(model.ProviderName),
			"customizations_supported":     aws.StringValueSlice(model.CustomizationsSupported),
			"inference_types_supported":    aws.StringValueSlice(model.InferenceTypesSupported),
			"input_modalities":             aws.StringValueSlice(model.InputModalities),
			"output_modalities":            aws.StringValueSlice(model.OutputModalities),
			"response_streaming_supported": aws.BoolValue(model.ResponseStreamingSupported),
		}

		l = append(l, m)
	}

	return l
}

func flattenLoggingConfig(apiObject *bedrock.LoggingConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CloudWatchConfig; v != nil {
		tfMap["cloud_watch_config"] = flattenCloudWatchConfig(v)
	}

	tfMap["embedding_data_delivery_enabled"] = aws.Bool(*apiObject.EmbeddingDataDeliveryEnabled)
	tfMap["image_data_delivery_enabled"] = aws.Bool(*apiObject.ImageDataDeliveryEnabled)
	tfMap["text_data_delivery_enabled"] = aws.Bool(*apiObject.TextDataDeliveryEnabled)

	if v := apiObject.S3Config; v != nil {
		tfMap["s3_config"] = flattenS3Config(v)
	}

	return tfMap
}

func flattenCloudWatchConfig(apiObject *bedrock.CloudWatchConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LargeDataDeliveryS3Config; v != nil {
		tfMap["large_data_delivery_s3_config"] = flattenS3Config(v)
	}
	if v := apiObject.LogGroupName; v != nil {
		tfMap["log_group_name"] = aws.StringValue(v)
	}
	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenS3Config(apiObject *bedrock.S3Config) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BucketName; v != nil {
		tfMap["bucket_name"] = aws.StringValue(v)
	}
	if v := apiObject.KeyPrefix; v != nil {
		tfMap["key_prefix"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func expandLoggingConfig(tfMap map[string]interface{}) *bedrock.LoggingConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &bedrock.LoggingConfig{}

	if v, ok := tfMap["cloud_watch_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.CloudWatchConfig = expandCloudWatchConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["embedding_data_delivery_enabled"].(bool); ok {
		apiObject.EmbeddingDataDeliveryEnabled = aws.Bool(v)
	}

	if v, ok := tfMap["image_data_delivery_enabled"].(bool); ok {
		apiObject.ImageDataDeliveryEnabled = aws.Bool(v)
	}

	if v, ok := tfMap["text_data_delivery_enabled"].(bool); ok {
		apiObject.TextDataDeliveryEnabled = aws.Bool(v)
	}

	if v, ok := tfMap["s3_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.S3Config = expandS3Config(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandCloudWatchConfig(tfMap map[string]interface{}) *bedrock.CloudWatchConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &bedrock.CloudWatchConfig{}

	if v, ok := tfMap["large_data_delivery_s3_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.LargeDataDeliveryS3Config = expandS3Config(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["log_group_name"].(string); ok && v != "" {
		apiObject.LogGroupName = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandS3Config(tfMap map[string]interface{}) *bedrock.S3Config {
	if tfMap == nil {
		return nil
	}

	apiObject := &bedrock.S3Config{}

	if v, ok := tfMap["bucket_name"].(string); ok && v != "" {
		apiObject.BucketName = aws.String(v)
	}
	if v, ok := tfMap["key_prefix"].(string); ok && v != "" {
		apiObject.KeyPrefix = aws.String(v)
	}

	return apiObject
}
