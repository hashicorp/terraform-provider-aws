package account_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/account"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccAccountAlternateContact_basic(t *testing.T) {
	resourceName := "aws_account_alternate_contact.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, account.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccountAlternateContactDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccountAlternateContactConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountAlternateContactExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudfront_zone_id", "Z2FDTNDATAQYW2"),
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

func testAccountAlternateContactDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AccountConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_account_alternate_contact" {
			continue
		}

		contactType := rs.Primary.Attributes["type"]

		input := &account.GetAlternateContactInput{AlternateContactType: aws.String(contactType)}

		resp, err := conn.GetAlternateContact(input)

		if tfawserr.ErrCodeEquals(err, account.ErrCodeResourceNotFoundException) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error reading Account Alternate Contact (%s): %w", rs.Primary.Attributes["type"], err)
		}

		if resp == nil {
			return fmt.Errorf("error reading Account Alternate Contact (%s): empty response", rs.Primary.Attributes["type"])
		}
	}

	return nil

}

func testAccCheckAccountAlternateContactExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AccountConn
		contactType := rs.Primary.Attributes["type"]

		input := &account.GetAlternateContactInput{AlternateContactType: aws.String(contactType)}

		_, err := conn.GetAlternateContact(input)
		if err != nil {
			return fmt.Errorf("error reading Account Alternate Contact (%s): %w", rs.Primary.Attributes["type"], err)
		}

		return nil
	}
}

func testAccountAlternateContactConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_account_alternate_contact" "test" {
  type          = "SECURITY"
  name          = %[1]q
  title         = %[1]q
  email_address = "test@test.test"
  phone_number  = "+1234567890"
}
`, rName)
}
