package aws

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccAWSSESIdentityFeedbackForwardingEnabled_basic(t *testing.T) {
	domain := fmt.Sprintf(
		"%s.terraformtesting.com",
		acctest.RandString(10))
	//resourceName := "aws_ses_identity_feedback_forwarding_enabled.test"
	forwardingEnabled := true

	resource.ParallelTest(t, resource.TestCase{
		//PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSES(t) },
		Providers: testAccProviders,
		//CheckDestroy: testAccCheckSESDomainMailFromDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESIdentityFeedbackForwardingEnabledConfig(domain, forwardingEnabled),
				Check:  resource.ComposeTestCheckFunc(
				//testAccCheckAwsSESDomainMailFromExists(resourceName),
				//resource.TestCheckResourceAttr(resourceName, "behavior_on_mx_failure", ses.BehaviorOnMXFailureUseDefaultValue),
				//resource.TestCheckResourceAttr(resourceName, "domain", domain),
				//resource.TestCheckResourceAttr(resourceName, "mail_from_domain", mailFromDomain1),
				),
			},
		},
	})
}

func testAccAwsSESIdentityFeedbackForwardingEnabledConfig(domain string, fowardingEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}

resource "aws_ses_identity_feedback_forwarding_enabled" "test" {
  domain           = aws_ses_domain_identity.test.domain
  enabled = %v
}
`, domain, fowardingEnabled)
}
