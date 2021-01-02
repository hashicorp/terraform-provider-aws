package aws

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSESIdentityFeedbackForwardingEnabled_basic(t *testing.T) {
	domain := fmt.Sprintf(
		"%s.terraformtesting.com",
		acctest.RandString(10))
	resourceName := "aws_ses_identity_feedback_forwarding_enabled.test"
	forwardingEnabled := false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSES(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSESIdentityFeedbackForwardingEnabledDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSESIdentityFeedbackForwardingEnabledConfig(domain, forwardingEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSESEmailIdentityExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity", domain),
					resource.TestCheckResourceAttr(resourceName, "enabled", strconv.FormatBool(forwardingEnabled)),
				),
			},
		},
	})
}

func testAccAwsSESIdentityFeedbackForwardingEnabledConfig(domain string, fowardingEnabled bool) string {
	config := fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = "%v"
}

resource "aws_ses_identity_feedback_forwarding_enabled" "test" {
  identity = aws_ses_domain_identity.test.domain
  enabled  = "%v"
}

`, domain, fowardingEnabled)
	return config

}

func testAccCheckAwsSESIdentityFeedbackForwardingEnabledDestroy(s *terraform.State) error {
	fmt.Println("testAccCheckAwsSESIdentityFeedbackForwardingEnabledDestroy")

	// List registered E-mail identities
	conn := testAccProvider.Meta().(*AWSClient).sesv2conn
	list, err := conn.ListEmailIdentities(&sesv2.ListEmailIdentitiesInput{})
	if err != nil {
		return err
	}

	// Loop on resources in terraform
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ses_identity_feedback_forwarding_enabled" {
			continue
		}

		// If resources in terraform still exists, it fails
		identity := rs.Primary.ID
		for _, item := range list.EmailIdentities {
			if identity == *item.IdentityName {
				return fmt.Errorf("SES Email identity still exists")
			}
		}
	}

	return nil
}
