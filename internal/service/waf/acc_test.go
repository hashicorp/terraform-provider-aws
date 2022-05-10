package waf_test

import (
	"context"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

// WAF Logging Configurations can only be enabled with destinations in specific regions,

// testAccWafLoggingConfigurationRegion is the chosen WAF Logging Configurations testing region
//
// Cached to prevent issues should multiple regions become available.
var testAccWafLoggingConfigurationRegion string

// testAccProviderWafLoggingConfiguration is the WAF Logging Configurations provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheckLoggingConfiguration(t) must be called before using this provider instance.
var testAccProviderWafLoggingConfiguration *schema.Provider

// testAccProviderWafLoggingConfigurationConfigure ensures the provider is only configured once
var testAccProviderWafLoggingConfigurationConfigure sync.Once

// testAccPreCheckLoggingConfiguration verifies AWS credentials and that WAF Logging Configurations is supported
func testAccPreCheckLoggingConfiguration(t *testing.T) {
	acctest.PreCheckPartitionHasService(waf.EndpointsID, t)

	region := testAccGetLoggingConfigurationRegion()

	if region == "" {
		t.Skip("WAF Logging Configuration not available in this AWS Partition")
	}

	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderWafLoggingConfigurationConfigure.Do(func() {
		testAccProviderWafLoggingConfiguration = provider.Provider()

		config := map[string]interface{}{
			"region": region,
		}

		diags := testAccProviderWafLoggingConfiguration.Configure(context.Background(), terraform.NewResourceConfigRaw(config))

		if diags != nil && diags.HasError() {
			for _, d := range diags {
				if d.Severity == diag.Error {
					t.Fatalf("error configuring WAF Logging Configurations provider: %s", d.Summary)
				}
			}
		}
	})
}

// testAccLoggingConfigurationRegionProviderConfig is the Terraform provider configuration for WAF Logging Configurations region testing
//
// Testing WAF Logging Configurations assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccLoggingConfigurationRegionProviderConfig() string {
	return acctest.ConfigRegionalProvider(testAccGetLoggingConfigurationRegion())
}

// testAccGetLoggingConfigurationRegion returns the WAF Logging Configurations region for testing
func testAccGetLoggingConfigurationRegion() string {
	if testAccWafLoggingConfigurationRegion != "" {
		return testAccWafLoggingConfigurationRegion
	}

	// AWS Commercial: https://docs.aws.amazon.com/waf/latest/developerguide/classic-logging.html
	// AWS GovCloud (US) - not available yet: https://docs.aws.amazon.com/govcloud-us/latest/UserGuide/govcloud-waf.html
	// AWS China - not available yet
	switch acctest.Partition() {
	case endpoints.AwsPartitionID:
		testAccWafLoggingConfigurationRegion = endpoints.UsEast1RegionID
	}

	return testAccWafLoggingConfigurationRegion
}
