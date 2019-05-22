package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
)

func TestAccAwsRamResourceShareAccepter_basic(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_ram_resource_share_accepter.test"
	shareName := fmt.Sprintf("tf-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckAwsRamResourceShareAccepterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRamResourceShareAccepterBasic(shareName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsRamResourceShareAccepterExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "share_arn", regexp.MustCompile(`^arn\:aws\:ram\:.*resource-share/.+$`)),
					resource.TestMatchResourceAttr(resourceName, "invitation_arn", regexp.MustCompile(`^arn\:aws\:ram\:.*resource-share-invitation/.+$`)),
					resource.TestMatchResourceAttr(resourceName, "share_id", regexp.MustCompile(`^rs-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "status", ram.ResourceShareInvitationStatusAccepted),
					resource.TestMatchResourceAttr(resourceName, "receiver_account_id", regexp.MustCompile(`\d{12}`)),
					resource.TestMatchResourceAttr(resourceName, "sender_account_id", regexp.MustCompile(`\d{12}`)),
					resource.TestCheckResourceAttr(resourceName, "share_name", shareName),
					resource.TestCheckResourceAttr(resourceName, "resources.%", "0"),
				),
			},
			{
				Config:            testAccAwsRamResourceShareAccepterBasic(shareName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsRamResourceShareAccepterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ramconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ram_resource_share_accepter" {
			continue
		}

		request := &ram.GetResourceSharesInput{
			ResourceShareArns: []*string{aws.String(rs.Primary.Attributes["share_arn"])},
			ResourceOwner:     aws.String(ram.ResourceOwnerOtherAccounts),
		}

		resp, err := conn.GetResourceShares(request)
		if err != nil {
			if awserr, ok := err.(awserr.Error); ok {
				switch awserr.Code() {
				case ram.ErrCodeUnknownResourceException:
					// good - it's gone
					return nil
				}
			}
			return fmt.Errorf("Error deleting RAM resource share: %s", err)
		}

		if len(resp.ResourceShares) == 0 {
			return nil
		}
	}
	return fmt.Errorf("RAM resource share association found, should be destroyed")
}

func testAccCheckAwsRamResourceShareAccepterExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAwsRamResourceShareAccepterBasic(shareName string) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  provider = "aws.alternate"

  name                      = %q
  allow_external_principals = true

  tags = {
	Name = %q
  }
}

resource "aws_ram_principal_association" "test" {
  provider = "aws.alternate"

  principal          = "${data.aws_caller_identity.receiver.account_id}"
  resource_share_arn = "${aws_ram_resource_share.test.arn}"

  depends_on = ["data.aws_caller_identity.receiver"]
}

data "aws_caller_identity" "receiver" {}

resource "aws_ram_resource_share_accepter" "test" {
  share_arn = "${aws_ram_resource_share.test.arn}"

  depends_on = ["aws_ram_resource_share.test", "aws_ram_principal_association.test"]
}
`, shareName, shareName)
}
