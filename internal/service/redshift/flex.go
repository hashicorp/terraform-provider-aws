package redshift

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func flattenLogging(ls *redshift.LoggingStatus) []interface{} {
	if ls == nil {
		return []interface{}{}
	}

	cfg := make(map[string]interface{})
	cfg["enable"] = aws.BoolValue(ls.LoggingEnabled)
	if ls.BucketName != nil {
		cfg["bucket_name"] = aws.StringValue(ls.BucketName)
	}
	if ls.LogDestinationType != nil {
		cfg["log_destination_type"] = aws.StringValue(ls.LogDestinationType)
	}
	if ls.LogExports != nil {
		cfg["log_exports"] = flex.FlattenStringSet(ls.LogExports)
	}
	if ls.S3KeyPrefix != nil {
		cfg["s3_key_prefix"] = aws.StringValue(ls.S3KeyPrefix)
	}
	return []interface{}{cfg}
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
