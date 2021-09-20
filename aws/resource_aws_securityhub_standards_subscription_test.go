package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/securityhub/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

func testAccAWSSecurityHubStandardsSubscription_basic(t *testing.T) {
	var standardsSubscription securityhub.StandardsSubscription
	resourceName := "aws_securityhub_standards_subscription.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, securityhub.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSecurityHubStandardsSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubStandardsSubscriptionConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubStandardsSubscriptionExists(resourceName, &standardsSubscription),
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

func testAccAWSSecurityHubStandardsSubscription_disappears(t *testing.T) {
	var standardsSubscription securityhub.StandardsSubscription
	resourceName := "aws_securityhub_standards_subscription.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, securityhub.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSecurityHubStandardsSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubStandardsSubscriptionConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubStandardsSubscriptionExists(resourceName, &standardsSubscription),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceStandardsSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSecurityHubStandardsSubscriptionExists(n string, standardsSubscription *securityhub.StandardsSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Security Hub Standards Subscription ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn

		output, err := finder.StandardsSubscriptionByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*standardsSubscription = *output

		return nil
	}
}

func testAccCheckAWSSecurityHubStandardsSubscriptionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_standards_subscription" {
			continue
		}

		output, err := finder.StandardsSubscriptionByARN(conn, rs.Primary.ID)

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

const testAccAWSSecurityHubStandardsSubscriptionConfig_basic = `
resource "aws_securityhub_account" "test" {}

data "aws_partition" "current" {}

resource "aws_securityhub_standards_subscription" "test" {
  standards_arn = "arn:${data.aws_partition.current.partition}:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
  depends_on    = [aws_securityhub_account.test]
}
`
