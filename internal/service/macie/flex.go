package macie

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandClassificationType(d *schema.ResourceData) *macie.ClassificationType {
	continuous := macie.S3ContinuousClassificationTypeFull
	oneTime := macie.S3OneTimeClassificationTypeNone
	if v := d.Get("classification_type").([]interface{}); len(v) > 0 {
		m := v[0].(map[string]interface{})
		continuous = m["continuous"].(string)
		oneTime = m["one_time"].(string)
	}

	return &macie.ClassificationType{
		Continuous: aws.String(continuous),
		OneTime:    aws.String(oneTime),
	}
}

func expandClassificationTypeUpdate(d *schema.ResourceData) *macie.ClassificationTypeUpdate {
	continuous := macie.S3ContinuousClassificationTypeFull
	oneTime := macie.S3OneTimeClassificationTypeNone
	if v := d.Get("classification_type").([]interface{}); len(v) > 0 {
		m := v[0].(map[string]interface{})
		continuous = m["continuous"].(string)
		oneTime = m["one_time"].(string)
	}

	return &macie.ClassificationTypeUpdate{
		Continuous: aws.String(continuous),
		OneTime:    aws.String(oneTime),
	}
}

func flattenClassificationType(classificationType *macie.ClassificationType) []map[string]interface{} {
	if classificationType == nil {
		return []map[string]interface{}{}
	}
	m := map[string]interface{}{
		"continuous": aws.StringValue(classificationType.Continuous),
		"one_time":   aws.StringValue(classificationType.OneTime),
	}
	return []map[string]interface{}{m}
}
