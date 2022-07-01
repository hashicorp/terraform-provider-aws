package cur_test

import (
	"context"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/costandusagereportservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
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
// testAccPreCheck(t) must be called before using this provider instance.
var testAccProviderCur *schema.Provider

// testAccProviderCurConfigure ensures the provider is only configured once
var testAccProviderCurConfigure sync.Once

// testAccPreCheck verifies AWS credentials and that Cost and Usage Reporting is supported
func testAccPreCheck(t *testing.T) {
	acctest.PreCheckPartitionHasService(costandusagereportservice.ServiceName, t)

	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderCurConfigure.Do(func() {
		testAccProviderCur = provider.Provider()

		config := map[string]interface{}{
			"region": testAccGetRegion(),
		}

		diags := testAccProviderCur.Configure(context.Background(), terraform.NewResourceConfigRaw(config))

		if diags != nil && diags.HasError() {
			for _, d := range diags {
				if d.Severity == diag.Error {
					t.Fatalf("error configuring CUR provider: %s", d.Summary)
				}
			}
		}
	})

	conn := testAccProviderCur.Meta().(*conns.AWSClient).CURConn

	input := &costandusagereportservice.DescribeReportDefinitionsInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeReportDefinitions(input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrMessageContains(err, "AccessDeniedException", "linked account is not allowed to modify report preference") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// testAccRegionProviderConfig is the Terraform provider configuration for Cost and Usage Reporting region testing
//
// Testing Cost and Usage Reporting assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccRegionProviderConfig() string {
	return acctest.ConfigRegionalProvider(testAccGetRegion())
}

// testAccGetRegion returns the Cost and Usage Reporting region for testing
func testAccGetRegion() string {
	if testAccCurRegion != "" {
		return testAccCurRegion
	}

	if rs, ok := endpoints.RegionsForService(endpoints.DefaultPartitions(), acctest.Partition(), costandusagereportservice.ServiceName); ok {
		// return available region (random if multiple)
		for regionID := range rs {
			testAccCurRegion = regionID
			return testAccCurRegion
		}
	}

	testAccCurRegion = acctest.Region()

	return testAccCurRegion
}
