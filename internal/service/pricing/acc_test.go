package pricing_test

import (
	"context"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

// testAccPricingRegion is the chosen Pricing testing region
//
// Cached to prevent issues should multiple regions become available.
var testAccPricingRegion string

// testAccProviderPricing is the Pricing provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheck(t) must be called before using this provider instance.
var testAccProviderPricing *schema.Provider

// testAccProviderPricingConfigure ensures the provider is only configured once
var testAccProviderPricingConfigure sync.Once

// testAccPreCheck verifies AWS credentials and that Pricing is supported
func testAccPreCheck(t *testing.T) {
	acctest.PreCheckPartitionHasService(pricing.EndpointsID, t)

	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderPricingConfigure.Do(func() {
		testAccProviderPricing = provider.Provider()

		config := map[string]interface{}{
			"region": testAccGetRegion(),
		}

		diags := testAccProviderPricing.Configure(context.Background(), terraform.NewResourceConfigRaw(config))

		if diags != nil && diags.HasError() {
			for _, d := range diags {
				if d.Severity == diag.Error {
					t.Fatalf("error configuring Pricing provider: %s", d.Summary)
				}
			}
		}
	})
}

// testAccRegionProviderConfig is the Terraform provider configuration for Pricing region testing
//
// Testing Pricing assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccRegionProviderConfig() string {
	return acctest.ConfigRegionalProvider(testAccGetRegion())
}

// testAccGetRegion returns the Pricing region for testing
func testAccGetRegion() string {
	if testAccPricingRegion != "" {
		return testAccPricingRegion
	}

	if rs, ok := endpoints.RegionsForService(endpoints.DefaultPartitions(), acctest.Partition(), pricing.ServiceName); ok {
		// return available region (random if multiple)
		for regionID := range rs {
			testAccPricingRegion = regionID
			return testAccPricingRegion
		}
	}

	testAccPricingRegion = acctest.Region()

	return testAccPricingRegion
}
