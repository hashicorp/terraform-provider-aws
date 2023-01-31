package apigateway_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

// API Gateway Edge-Optimized Domain Name can only be created with ACM Certificates in specific regions.

// testAccEdgeDomainNameRegion is the chosen API Gateway Domain Name testing region
//
// Cached to prevent issues should multiple regions become available.
var testAccEdgeDomainNameRegion string

// testAccProviderEdgeDomainName is the API Gateway Domain Name provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheckEdgeDomainName(t) must be called before using this provider instance.
var testAccProviderEdgeDomainName *schema.Provider

// testAccProviderEdgeDomainNameConfigure ensures the provider is only configured once
var testAccProviderEdgeDomainNameConfigure sync.Once

// testAccPreCheckEdgeDomainName verifies AWS credentials and that API Gateway Domain Name is supported
func testAccPreCheckEdgeDomainName(ctx context.Context, t *testing.T) {
	acctest.PreCheckPartitionHasService(apigateway.EndpointsID, t)

	region := testAccGetEdgeDomainNameRegion()

	if region == "" {
		t.Skip("API Gateway Domain Name not available in this AWS Partition")
	}

	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderEdgeDomainNameConfigure.Do(func() {
		var err error
		testAccProviderEdgeDomainName, err = provider.New(ctx)

		if err != nil {
			t.Fatal(err)
		}

		config := map[string]interface{}{
			"region": region,
		}

		diags := testAccProviderEdgeDomainName.Configure(ctx, terraform.NewResourceConfigRaw(config))

		if diags != nil && diags.HasError() {
			for _, d := range diags {
				if d.Severity == diag.Error {
					t.Fatalf("error configuring API Gateway Domain Name provider: %s", d.Summary)
				}
			}
		}
	})
}

// testAccEdgeDomainNameRegionProviderConfig is the Terraform provider configuration for API Gateway Domain Name region testing
//
// Testing API Gateway Domain Name assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccEdgeDomainNameRegionProviderConfig() string {
	return acctest.ConfigRegionalProvider(testAccGetEdgeDomainNameRegion())
}

// testAccEdgeDomainNameRegion returns the API Gateway Domain Name region for testing
func testAccGetEdgeDomainNameRegion() string {
	if testAccEdgeDomainNameRegion != "" {
		return testAccEdgeDomainNameRegion
	}

	// AWS Commercial: https://docs.aws.amazon.com/apigateway/latest/developerguide/how-to-custom-domains.html
	// AWS GovCloud (US) - edge custom domain names not supported: https://docs.aws.amazon.com/govcloud-us/latest/UserGuide/govcloud-abp.html
	// AWS China - edge custom domain names not supported: https://docs.amazonaws.cn/en_us/aws/latest/userguide/api-gateway.html
	switch acctest.Partition() {
	case endpoints.AwsPartitionID:
		testAccEdgeDomainNameRegion = endpoints.UsEast1RegionID
	}

	return testAccEdgeDomainNameRegion
}

// testAccCheckResourceAttrRegionalARNEdgeDomainName ensures the Terraform state exactly matches the expected API Gateway Edge Domain Name format
func testAccCheckResourceAttrRegionalARNEdgeDomainName(resourceName, attributeName, arnService string, domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			Partition: acctest.Partition(),
			Region:    testAccGetEdgeDomainNameRegion(),
			Resource:  fmt.Sprintf("/domainnames/%s", domain),
			Service:   arnService,
		}.String()

		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}

// testAccCheckResourceAttrRegionalARNRegionalDomainName ensures the Terraform state exactly matches the expected API Gateway Regional Domain Name format
func testAccCheckResourceAttrRegionalARNRegionalDomainName(resourceName, attributeName, arnService string, domain string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := arn.ARN{
			Partition: acctest.Partition(),
			Region:    acctest.Region(),
			Resource:  fmt.Sprintf("/domainnames/%s", domain),
			Service:   arnService,
		}.String()

		return resource.TestCheckResourceAttr(resourceName, attributeName, attributeValue)(s)
	}
}
