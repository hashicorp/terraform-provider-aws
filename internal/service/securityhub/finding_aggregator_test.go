package securityhub_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
)

func testAccFindingAggregator_basic(t *testing.T) {
	resourceName := "aws_securityhub_finding_aggregator.test_aggregator"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFindingAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFindingAggregatorConfig_allRegions(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingAggregatorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "linking_mode", "ALL_REGIONS"),
					resource.TestCheckNoResourceAttr(resourceName, "specified_regions"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFindingAggregatorConfig_specifiedRegions(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingAggregatorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "linking_mode", "SPECIFIED_REGIONS"),
					resource.TestCheckResourceAttr(resourceName, "specified_regions.#", "3"),
				),
			},
			{
				Config: testAccFindingAggregatorConfig_allRegionsExceptSpecified(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingAggregatorExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "linking_mode", "ALL_REGIONS_EXCEPT_SPECIFIED"),
					resource.TestCheckResourceAttr(resourceName, "specified_regions.#", "2"),
				),
			},
		},
	})
}

func testAccFindingAggregator_disappears(t *testing.T) {
	resourceName := "aws_securityhub_finding_aggregator.test_aggregator"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFindingAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFindingAggregatorConfig_allRegions(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFindingAggregatorExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfsecurityhub.ResourceFindingAggregator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFindingAggregatorExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Security Hub finding aggregator ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn

		_, err := conn.GetFindingAggregator(&securityhub.GetFindingAggregatorInput{
			FindingAggregatorArn: &rs.Primary.ID,
		})

		if err != nil {
			return fmt.Errorf("Failed to get finding aggregator: %s", err)
		}

		return nil
	}
}

func testAccCheckFindingAggregatorDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_finding_aggregator" {
			continue
		}

		_, err := conn.GetFindingAggregator(&securityhub.GetFindingAggregatorInput{
			FindingAggregatorArn: &rs.Primary.ID,
		})

		if tfawserr.ErrMessageContains(err, securityhub.ErrCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Security Hub Finding Aggregator %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccFindingAggregatorConfig_allRegions() string {
	return `
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_finding_aggregator" "test_aggregator" {
  linking_mode = "ALL_REGIONS"

  depends_on = [aws_securityhub_account.example]
}
`
}

func testAccFindingAggregatorConfig_specifiedRegions() string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_finding_aggregator" "test_aggregator" {
  linking_mode      = "SPECIFIED_REGIONS"
  specified_regions = ["%s", "%s", "%s"]

  depends_on = [aws_securityhub_account.example]
}
`, endpoints.EuWest1RegionID, endpoints.EuWest2RegionID, endpoints.UsEast1RegionID)
}

func testAccFindingAggregatorConfig_allRegionsExceptSpecified() string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_finding_aggregator" "test_aggregator" {
  linking_mode      = "ALL_REGIONS_EXCEPT_SPECIFIED"
  specified_regions = ["%s", "%s"]

  depends_on = [aws_securityhub_account.example]
}
`, endpoints.EuWest1RegionID, endpoints.EuWest2RegionID)
}
