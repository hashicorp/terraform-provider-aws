package aws

import (
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/terraform-providers/terraform-provider-aws/atest"
)

// testAccCloudfrontRegionProviderConfig is the Terraform provider configuration for CloudFront region testing
//
// Testing CloudFront assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccCloudfrontRegionProviderConfig() string {
	switch atest.Partition() {
	case endpoints.AwsPartitionID:
		return atest.ConfigProviderRegional(endpoints.UsEast1RegionID)
	case endpoints.AwsCnPartitionID:
		return atest.ConfigProviderRegional(endpoints.CnNorthwest1RegionID)
	default:
		return atest.ConfigProviderRegional(atest.Region())
	}
}
