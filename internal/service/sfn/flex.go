// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sfn/types"
)

func expandEncryptionConfiguration(tfMap map[string]interface{}) *awstypes.EncryptionConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.EncryptionConfiguration{}

	if v, ok := tfMap["kms_key_id"].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = awstypes.EncryptionType(v)
	}

	if v, ok := tfMap["kms_data_key_reuse_period_seconds"].(int); ok && v != 0 {
		apiObject.KmsDataKeyReusePeriodSeconds = aws.Int32(int32(v))
	}

	return apiObject
}

func flattenEncryptionConfiguration(apiObject *awstypes.EncryptionConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"kms_key_id": apiObject.KmsKeyId,
		"type":       apiObject.Type,
	}

	if v := apiObject.KmsDataKeyReusePeriodSeconds; v != nil {
		tfMap["kms_data_key_reuse_period_seconds"] = *v
	}

	return tfMap
}
