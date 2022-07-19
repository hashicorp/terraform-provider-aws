package securityhub_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccStandardsSubscription_basic(t *testing.T) {
	var standardsSubscription securityhub.StandardsSubscription
	resourceName := "aws_securityhub_standards_subscription.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStandardsSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStandardsSubscriptionConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStandardsSubscriptionExists(resourceName, &standardsSubscription),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccStandardsSubscription_disappears(t *testing.T) {
	var standardsSubscription securityhub.StandardsSubscription
	resourceName := "aws_securityhub_standards_subscription.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckStandardsSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStandardsSubscriptionConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStandardsSubscriptionExists(resourceName, &standardsSubscription),
					acctest.CheckResourceDisappears(acctest.Provider, tfsecurityhub.ResourceStandardsSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStandardsSubscriptionExists(n string, standardsSubscription *securityhub.StandardsSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Security Hub Standards Subscription ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn

		output, err := tfsecurityhub.FindStandardsSubscriptionByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*standardsSubscription = *output

		return nil
	}
}

func testAccCheckStandardsSubscriptionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_standards_subscription" {
			continue
		}

		output, err := tfsecurityhub.FindStandardsSubscriptionByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		// INCOMPLETE subscription status => deleted.
		if aws.StringValue(output.StandardsStatus) == securityhub.StandardsStatusIncomplete {
			continue
		}

		return fmt.Errorf("Security Hub Standards Subscription %s still exists", rs.Primary.ID)
	}

	return nil
}

const testAccStandardsSubscriptionConfig_basic = `
resource "aws_securityhub_account" "test" {}

data "aws_partition" "current" {}

resource "aws_securityhub_standards_subscription" "test" {
  standards_arn = "arn:${data.aws_partition.current.partition}:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
  depends_on    = [aws_securityhub_account.test]
}
`
