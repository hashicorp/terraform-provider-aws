package aws

import (
	"context"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// FMS Admin Account Configurations can only be enabled with destinations in specific regions,

// testAccFmsAdminAccountConfigurationRegion is the chosen FMS Admin Account Configurations testing region
//
// Cached to prevent issues should multiple regions become available.
var testAccFmsAdminAccountConfigurationRegion string

// testAccProviderFmsAdminAccountConfiguration is the FMS Admin Account Configurations provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheckFmsAdminAccountConfiguration(t) must be called before using this provider instance.
var testAccProviderFmsAdminAccountConfiguration *schema.Provider

// testAccProviderFmsAdminAccountConfigurationConfigure ensures the provider is only configured once
var testAccProviderFmsAdminAccountConfigurationConfigure sync.Once

// testAccPreCheckFmsAdminAccountConfiguration verifies AWS credentials and that FMS Admin Account Configurations is supported
func testAccPreCheckFmsAdminAccountConfiguration(t *testing.T) {
	testAccPartitionHasServicePreCheck(fms.EndpointsID, t)

	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderFmsAdminAccountConfigurationConfigure.Do(func() {
		testAccProviderFmsAdminAccountConfiguration = Provider()

		region := testAccGetFmsAdminAccountConfigurationRegion()

		if region == "" {
			t.Skip("FMS Admin Account Configuration not available in this AWS Partition")
		}

		config := map[string]interface{}{
			"region": region,
		}

		diags := testAccProviderFmsAdminAccountConfiguration.Configure(context.Background(), terraform.NewResourceConfigRaw(config))

		if diags != nil && diags.HasError() {
			for _, d := range diags {
				if d.Severity == diag.Error {
					t.Fatalf("error configuring FMS Admin Account Configurations provider: %s", d.Summary)
				}
			}
		}
	})
}

// testAccFmsAdminAccountConfigurationRegionProviderConfig is the Terraform provider configuration for FMS Admin Account Configurations region testing
//
// Testing FMS Admin Account Configurations assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccFmsAdminAccountConfigurationRegionProviderConfig() string {
	return testAccRegionalProviderConfig(testAccGetFmsAdminAccountConfigurationRegion())
}

// testAccGetFmsAdminAccountConfigurationRegion returns the FMS Admin Account Configurations region for testing
func testAccGetFmsAdminAccountConfigurationRegion() string {
	if testAccFmsAdminAccountConfigurationRegion != "" {
		return testAccFmsAdminAccountConfigurationRegion
	}

	switch testAccGetPartition() {
	case endpoints.AwsPartitionID:
		testAccFmsAdminAccountConfigurationRegion = endpoints.UsEast1RegionID
	}

	return testAccFmsAdminAccountConfigurationRegion
}
