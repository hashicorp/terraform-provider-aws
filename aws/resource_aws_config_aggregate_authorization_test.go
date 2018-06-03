package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_config_aggregate_authorization", &resource.Sweeper{
		Name: "aws_config_aggregate_authorization",
		F:    testSweepConfigAggregateAuthorizations,
	})
}

func testSweepConfigAggregateAuthorizations(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).configconn

	resp, err := conn.DescribeAggregationAuthorizations(&configservice.DescribeAggregationAuthorizationsInput{})
	if err != nil {
		return fmt.Errorf("Error retrieving config aggregate authorizations: %s", err)
	}

	if len(resp.AggregationAuthorizations) == 0 {
		log.Print("[DEBUG] No config aggregate authorizations to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d config aggregate authorizations", len(resp.AggregationAuthorizations))

	for _, auth := range resp.AggregationAuthorizations {
		log.Printf("[INFO] Deleting config authorization %s", *auth.AggregationAuthorizationArn)
		_, err := conn.DeleteAggregationAuthorization(&configservice.DeleteAggregationAuthorizationInput{
			AuthorizedAccountId: auth.AuthorizedAccountId,
			AuthorizedAwsRegion: auth.AuthorizedAwsRegion,
		})
		if err != nil {
			return fmt.Errorf("Error deleting config aggregate authorization %s: %s", *auth.AggregationAuthorizationArn, err)
		}
	}

	return nil
}

func TestAccAWSConfigAggregateAuthorization_basic(t *testing.T) {
	rString := acctest.RandStringFromCharSet(12, "0123456789")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSConfigAggregateAuthorizationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSConfigAggregateAuthorizationConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_config_aggregate_authorization.example", "account_id", rString),
					resource.TestCheckResourceAttr("aws_config_aggregate_authorization.example", "region", "eu-west-1"),
					resource.TestMatchResourceAttr("aws_config_aggregate_authorization.example", "arn", regexp.MustCompile("^arn:aws:config:[\\w-]+:\\d{12}:aggregation-authorization/\\d{12}/[\\w-]+$")),
				),
			},
		},
	})
}

func testAccCheckAWSConfigAggregateAuthorizationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).configconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_aggregate_authorization" {
			continue
		}

		resp, err := conn.DescribeAggregationAuthorizations(&configservice.DescribeAggregationAuthorizationsInput{})

		if err == nil {
			if len(resp.AggregationAuthorizations) != 0 &&
				*resp.AggregationAuthorizations[0].AuthorizedAccountId == rs.Primary.Attributes["account_id"] {
				return fmt.Errorf("Config aggregate authorization still exists: %s", rs.Primary.Attributes["account_id"])
			}
		}
	}

	return nil
}

func testAccAWSConfigAggregateAuthorizationConfig_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_config_aggregate_authorization" "example" {
  account_id = "%s" # Required
  region = "eu-west-1"    # Required
}`, rString)
}
