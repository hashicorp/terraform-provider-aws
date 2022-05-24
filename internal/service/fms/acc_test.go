package fms_test

import (
	"context"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

// Firewall Management Service admin APIs are only enabled in specific regions, otherwise:
// InvalidOperationException: This operation is not supported in the 'us-west-2' region.

// testAccAdminRegion is the chosen Firewall Management Service testing region
//
// Cached to prevent issues should multiple regions become available.
var testAccAdminRegion string

// testAccProviderAdmin is the Firewall Management Service provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheckAdmin(t) must be called before using this provider instance.
var testAccProviderAdmin *schema.Provider

// testAccProviderAdminConfigure ensures the provider is only configured once
var testAccProviderAdminConfigure sync.Once

// testAccPreCheckAdmin verifies AWS credentials and that Firewall Management Service is supported
func testAccPreCheckAdmin(t *testing.T) {
	acctest.PreCheckPartitionHasService(fms.EndpointsID, t)

	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderAdminConfigure.Do(func() {
		testAccProviderAdmin = provider.Provider()

		config := map[string]interface{}{
			"region": testAccGetAdminRegion(),
		}

		diags := testAccProviderAdmin.Configure(context.Background(), terraform.NewResourceConfigRaw(config))

		if diags != nil && diags.HasError() {
			for _, d := range diags {
				if d.Severity == diag.Error {
					t.Fatalf("error configuring Firewall Management Service provider: %s", d.Summary)
				}
			}
		}
	})
}

// testAccAdminRegionProviderConfig is the Terraform provider configuration for Firewall Management Service region testing
//
// Testing Firewall Management Service assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccAdminRegionProviderConfig() string {
	return acctest.ConfigRegionalProvider(testAccGetAdminRegion())
}

// testAccGetAdminRegion returns the Firewall Management Service region for testing
func testAccGetAdminRegion() string {
	if testAccAdminRegion != "" {
		return testAccAdminRegion
	}

	testAccAdminRegion = endpoints.UsEast1RegionID

	return testAccAdminRegion
}
