package sesv2_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2EmailIdentityFeedbackAttributes_basic(t *testing.T) {
	rName := acctest.RandomEmailAddress(acctest.RandomDomainName())
	resourceName := "aws_sesv2_email_identity_feedback_attributes.test"
	emailIdentityName := "aws_sesv2_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityFeedbackAttributesConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(emailIdentityName),
					resource.TestCheckResourceAttrPair(resourceName, "email_identity", emailIdentityName, "email_identity"),
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

func TestAccSESV2EmailIdentityFeedbackAttributes_emailForwardingEnabled(t *testing.T) {
	rName := acctest.RandomEmailAddress(acctest.RandomDomainName())
	resourceName := "aws_sesv2_email_identity_feedback_attributes.test"
	emailIdentityName := "aws_sesv2_email_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2EndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEmailIdentityFeedbackAttributesConfig_emailForwardingEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(emailIdentityName),
					resource.TestCheckResourceAttr(resourceName, "email_forwarding_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEmailIdentityFeedbackAttributesConfig_emailForwardingEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailIdentityExists(emailIdentityName),
					resource.TestCheckResourceAttr(resourceName, "email_forwarding_enabled", "false"),
				),
			},
		},
	})
}

func testAccEmailIdentityFeedbackAttributesConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_email_identity" "test" {
  email_identity = %[1]q
}

resource "aws_sesv2_email_identity_feedback_attributes" "test" {
  email_identity = aws_sesv2_email_identity.test.email_identity
}
`, rName)
}

func testAccEmailIdentityFeedbackAttributesConfig_emailForwardingEnabled(rName string, emailForwardingEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_sesv2_email_identity" "test" {
  email_identity = %[1]q
}

resource "aws_sesv2_email_identity_feedback_attributes" "test" {
  email_identity           = aws_sesv2_email_identity.test.email_identity
  email_forwarding_enabled = %[2]t
}
`, rName, emailForwardingEnabled)
}
