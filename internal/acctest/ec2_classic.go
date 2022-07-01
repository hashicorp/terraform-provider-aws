package acctest

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

const (
	// EC2-Classic region testing environment variable name
	ec2ClassicRegionEnvVar = "AWS_EC2_CLASSIC_REGION"
)

// ProviderEC2Classic is the EC2-Classic provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// PreCheckEC2Classic(t) must be called before using this provider instance.
var ProviderEC2Classic *schema.Provider

// testAccProviderEc2ClassicConfigure ensures the provider is only configured once
var testAccProviderEc2ClassicConfigure sync.Once

// PreCheckEC2Classic verifies AWS credentials and that EC2-Classic is supported
func PreCheckEC2Classic(t *testing.T) {
	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderEc2ClassicConfigure.Do(func() {
		ProviderEC2Classic = provider.Provider()

		config := map[string]interface{}{
			"region": EC2ClassicRegion(),
		}

		err := ProviderEC2Classic.Configure(context.Background(), terraform.NewResourceConfigRaw(config))

		if err != nil {
			t.Fatal(err)
		}
	})

	client := ProviderEC2Classic.Meta().(*conns.AWSClient)
	platforms := client.SupportedPlatforms
	region := client.Region
	if !conns.HasEC2Classic(platforms) {
		t.Skipf("this test can only run in EC2-Classic, platforms available in %s: %q", region, platforms)
	}
}

// ConfigEC2ClassicRegionProvider is the Terraform provider configuration for EC2-Classic region testing
//
// Testing EC2-Classic assumes no other provider configurations are necessary
// and overwrites the "aws" provider configuration.
func ConfigEC2ClassicRegionProvider() string {
	return ConfigRegionalProvider(EC2ClassicRegion())
}

// EC2ClassicRegion returns the EC2-Classic region for testing
func EC2ClassicRegion() string {
	v := os.Getenv(ec2ClassicRegionEnvVar)

	if v != "" {
		return v
	}

	if Partition() == endpoints.AwsPartitionID {
		return endpoints.UsEast1RegionID
	}

	return Region()
}

// CheckResourceAttrRegionalARNEC2Classic ensures the Terraform state exactly matches a formatted ARN with EC2-Classic region
func CheckResourceAttrRegionalARNEC2Classic(resourceName, attributeName, arnService, arnResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			AccountID: AccountID(),
			Partition: Partition(),
			Region:    EC2ClassicRegion(),
			Resource:  arnResource,
			Service:   arnService,
		}.String()
		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}
