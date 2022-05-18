package redshift

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
)

func ExpandParameters(configured []interface{}) []*redshift.Parameter {
	var parameters []*redshift.Parameter

	// Loop over our configured parameters and create
	// an array of aws-sdk-go compatible objects
	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})

		if data["name"].(string) == "" {
			continue
		}

		p := &redshift.Parameter{
			ParameterName:  aws.String(data["name"].(string)),
			ParameterValue: aws.String(data["value"].(string)),
		}

		parameters = append(parameters, p)
	}

	return parameters
}

func flattenLogging(ls *redshift.LoggingStatus) []interface{} {
	if ls == nil {
		return []interface{}{}
	}

	cfg := make(map[string]interface{})
	cfg["enable"] = aws.BoolValue(ls.LoggingEnabled)
	if ls.BucketName != nil {
		cfg["bucket_name"] = aws.StringValue(ls.BucketName)
	}
	if ls.S3KeyPrefix != nil {
		cfg["s3_key_prefix"] = aws.StringValue(ls.S3KeyPrefix)
	}
	return []interface{}{cfg}
}

// Flattens an array of Redshift Parameters into a []map[string]interface{}
func FlattenParameters(list []*redshift.Parameter) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		result = append(result, map[string]interface{}{
			"name":  aws.StringValue(i.ParameterName),
			"value": aws.StringValue(i.ParameterValue),
		})
	}
	return result
}

func flattenSnapshotCopy(scs *redshift.ClusterSnapshotCopyStatus) []interface{} {
	if scs == nil {
		return []interface{}{}
	}

	cfg := make(map[string]interface{})
	if scs.DestinationRegion != nil {
		cfg["destination_region"] = aws.StringValue(scs.DestinationRegion)
	}
	if scs.RetentionPeriod != nil {
		cfg["retention_period"] = aws.Int64Value(scs.RetentionPeriod)
	}
	if scs.SnapshotCopyGrantName != nil {
		cfg["grant_name"] = aws.StringValue(scs.SnapshotCopyGrantName)
	}

	return []interface{}{cfg}
}
