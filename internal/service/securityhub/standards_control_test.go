package securityhub_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
)

func testAccStandardsControl_basic(t *testing.T) {
	var standardsControl securityhub.StandardsControl
	resourceName := "aws_securityhub_standards_control.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil, //lintignore:AT001
		Steps: []resource.TestStep{
			{
				Config: testAccStandardsControlConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStandardsControlExists(resourceName, &standardsControl),
					resource.TestCheckResourceAttr(resourceName, "control_id", "CIS.1.10"),
					resource.TestCheckResourceAttr(resourceName, "control_status", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "control_status_updated_at"),
					resource.TestCheckResourceAttr(resourceName, "description", "IAM password policies can prevent the reuse of a given password by the same user. It is recommended that the password policy prevent the reuse of passwords."),
					resource.TestCheckResourceAttr(resourceName, "disabled_reason", ""),
					resource.TestCheckResourceAttr(resourceName, "related_requirements.0", "CIS AWS Foundations 1.10"),
					resource.TestCheckResourceAttrSet(resourceName, "remediation_url"),
					resource.TestCheckResourceAttr(resourceName, "severity_rating", "LOW"),
					resource.TestCheckResourceAttr(resourceName, "title", "Ensure IAM password policy prevents password reuse"),
				),
			},
		},
	})
}

func testAccStandardsControl_disabledControlStatus(t *testing.T) {
	var standardsControl securityhub.StandardsControl
	resourceName := "aws_securityhub_standards_control.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil, //lintignore:AT001
		Steps: []resource.TestStep{
			{
				Config: testAccStandardsControlConfig_disabledStatus(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStandardsControlExists(resourceName, &standardsControl),
					resource.TestCheckResourceAttr(resourceName, "control_status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "disabled_reason", "We handle password policies within Okta"),
				),
			},
		},
	})
}

func testAccStandardsControl_enabledControlStatusAndDisabledReason(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil, //lintignore:AT001
		Steps: []resource.TestStep{
			{
				Config:      testAccStandardsControlConfig_enabledStatus(),
				ExpectError: regexp.MustCompile("InvalidInputException: DisabledReason should not be given for action other than disabling control"),
			},
		},
	})
}

func testAccCheckStandardsControlExists(n string, control *securityhub.StandardsControl) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Security Hub Standards Control ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn

		standardsSubscriptionARN, err := tfsecurityhub.StandardsControlARNToStandardsSubscriptionARN(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := tfsecurityhub.FindStandardsControlByStandardsSubscriptionARNAndStandardsControlARN(context.TODO(), conn, standardsSubscriptionARN, rs.Primary.ID)

		if err != nil {
			return err
		}

		*control = *output

		return nil
	}
}

func testAccStandardsControlConfig_basic() string {
	return acctest.ConfigCompose(
		testAccStandardsSubscriptionConfig_basic,
		`
resource aws_securityhub_standards_control test {
  standards_control_arn = format("%s/1.10", replace(aws_securityhub_standards_subscription.test.id, "subscription", "control"))
  control_status        = "ENABLED"
}
`)
}

func testAccStandardsControlConfig_disabledStatus() string {
	return acctest.ConfigCompose(
		testAccStandardsSubscriptionConfig_basic,
		`
resource aws_securityhub_standards_control test {
  standards_control_arn = format("%s/1.11", replace(aws_securityhub_standards_subscription.test.id, "subscription", "control"))
  control_status        = "DISABLED"
  disabled_reason       = "We handle password policies within Okta"
}
`)
}

func testAccStandardsControlConfig_enabledStatus() string {
	return acctest.ConfigCompose(
		testAccStandardsSubscriptionConfig_basic,
		`
resource aws_securityhub_standards_control test {
  standards_control_arn = format("%s/1.12", replace(aws_securityhub_standards_subscription.test.id, "subscription", "control"))
  control_status        = "ENABLED"
  disabled_reason       = "We handle password policies within Okta"
}
`)
}
