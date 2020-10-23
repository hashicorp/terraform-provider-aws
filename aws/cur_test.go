package aws

import (
	"context"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/costandusagereportservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// testAccCurRegion is the chosen Cost and Usage Reporting testing region
//
// Cached to prevent issues should multiple regions become available.
var testAccCurRegion string

// testAccProviderCur is the Cost and Usage Reporting provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheckCur(t) must be called before using this provider instance.
var testAccProviderCur *schema.Provider

// testAccProviderCurConfigure ensures the provider is only configured once
var testAccProviderCurConfigure sync.Once

// testAccPreCheckCur verifies AWS credentials and that Cost and Usage Reporting is supported
func testAccPreCheckCur(t *testing.T) {
	testAccPartitionHasServicePreCheck(costandusagereportservice.ServiceName, t)

	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderCurConfigure.Do(func() {
		testAccProviderCur = Provider()

		config := map[string]interface{}{
			"region": testAccGetCurRegion(),
		}

		err := testAccProviderCur.Configure(context.Background(), terraform.NewResourceConfigRaw(config))

		if err != nil {
			t.Fatal(err)
		}
	})
}

// testAccCurRegionProviderConfig is the Terraform provider configuration for Cost and Usage Reporting region testing
//
// Testing Cost and Usage Reporting assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccCurRegionProviderConfig() string {
	return testAccRegionalProviderConfig(testAccGetCurRegion())
}

// testAccGetCurRegion returns the Cost and Usage Reporting region for testing
func testAccGetCurRegion() string {
	if testAccCurRegion != "" {
		return testAccCurRegion
	}

	if rs, ok := endpoints.RegionsForService(endpoints.DefaultPartitions(), testAccGetPartition(), costandusagereportservice.ServiceName); ok {
		// return available region (random if multiple)
		for regionID := range rs {
			testAccCurRegion = regionID
			return testAccCurRegion
		}
	}

	testAccCurRegion = testAccGetRegion()

	return testAccCurRegion
}

// testAccProviderFactoriesCur initializes providers for Cost and Usage Reporting testing.
//
// Deprecated: This will be replaced with testAccProviderFactories when it only returns the "aws" provider
func testAccProviderFactoriesCur() map[string]func() (*schema.Provider, error) {
	return testAccProviderFactoriesInit(nil, []string{ProviderNameAws})
}
