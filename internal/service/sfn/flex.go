// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandEncryptionConfiguration(tfMap map[string]any) *awstypes.EncryptionConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.EncryptionConfiguration{}

	if v, ok := tfMap["kms_data_key_reuse_period_seconds"].(int); ok && v != 0 {
		apiObject.KmsDataKeyReusePeriodSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrKMSKeyID].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.EncryptionType(v)
	}

	return apiObject
}

func flattenEncryptionConfiguration(apiObject *awstypes.EncryptionConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrKMSKeyID: aws.ToString(apiObject.KmsKeyId),
		names.AttrType:     apiObject.Type,
	}

	if v := apiObject.KmsDataKeyReusePeriodSeconds; v != nil {
		tfMap["kms_data_key_reuse_period_seconds"] = aws.ToInt32(v)
	}

	return tfMap
}
