package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSecurityHubStandard_basic(t *testing.T) {
	var standardsSubscription *securityhub.StandardsSubscription

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubStandardConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubStandardExists("aws_securityhub_standard.example", standardsSubscription),
				),
			},
			{
				ResourceName:      "aws_securityhub_standard.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Check Destroy - but only target the specific resource (otherwise Security Hub
				// will be disabled and the destroy check will fail)
				Config: testAccAWSSecurityHubStandardConfig_empty,
				Check:  testAccCheckAWSSecurityHubStandardDestroy,
			},
		},
	})
}

func testAccCheckAWSSecurityHubStandardExists(n string, standardsSubscription *securityhub.StandardsSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).securityhubconn

		resp, err := conn.GetEnabledStandards(&securityhub.GetEnabledStandardsInput{
			StandardsSubscriptionArns: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if len(resp.StandardsSubscriptions) == 0 {
			return fmt.Errorf("Security Hub standard %s not found", rs.Primary.ID)
		}

		standardsSubscription = resp.StandardsSubscriptions[0]

		return nil
	}
}

func testAccCheckAWSSecurityHubStandardDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).securityhubconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_standard" {
			continue
		}

		resp, err := conn.GetEnabledStandards(&securityhub.GetEnabledStandardsInput{
			StandardsSubscriptionArns: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			if isAWSErr(err, securityhub.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}

		if len(resp.StandardsSubscriptions) != 0 {
			return fmt.Errorf("Security Hub standard %s still exists", rs.Primary.ID)
		}

		return nil
	}

	return nil
}

const testAccAWSSecurityHubStandardConfig_empty = `
resource "aws_securityhub_account" "example" {}
`

const testAccAWSSecurityHubStandardConfig_basic = `
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_standard" "example" {
  depends_on    = ["aws_securityhub_account.example"]
  standards_arn = "arn:aws:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
}
`
