package aws

import (
	"fmt"
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
					testAccCheckAwsSESIdentityFeedbackForwardingEnabledEnabled(resourceName),
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

func testAccCheckAwsSESIdentityFeedbackForwardingEnabledEnabled(n string) resource.TestCheckFunc {
	fmt.Println("testAccCheckAwsSESIdentityFeedbackForwardingEnabledEnabled")
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		fmt.Println("rs:", rs)
		//expectedNum := 3
		//expectedFormat := regexp.MustCompile("[a-z0-9]{32}")

		//tokenNum, _ := strconv.Atoi(rs.Primary.Attributes["dkim_tokens.#"])
		//if expectedNum != tokenNum {
		//	return fmt.Errorf("Incorrect number of DKIM tokens, expected: %d, got: %d", expectedNum, tokenNum)
		//}
		//for i := 0; i < expectedNum; i++ {
		//	key := fmt.Sprintf("dkim_tokens.%d", i)
		//	token := rs.Primary.Attributes[key]
		//	if !expectedFormat.MatchString(token) {
		//		return fmt.Errorf("Incorrect format of DKIM token: %v", token)
		//	}
		//}

		return nil
	}
}
