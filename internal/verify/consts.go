package verify

import (
	"github.com/aws/aws-sdk-go/aws/endpoints"
)

// CloudFrontDistributionHostedZoneIDForPartition returns for the Route 53 hosted zone ID
// for Amazon CloudFront distributions in the specified AWS partition.
func CloudFrontDistributionHostedZoneIDForPartition(partition string) string {
	if partition == endpoints.AwsCnPartitionID {
		return "Z3RFFRIM2A3IF5" // See https://docs.amazonaws.cn/en_us/aws/latest/userguide/route53.html
	}
	return "Z2FDTNDATAQYW2" // See https://docs.aws.amazon.com/Route53/latest/APIReference/API_AliasTarget.html#Route53-Type-AliasTarget-HostedZoneId
}
