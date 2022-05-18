package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccEC2EBSEncryptionByDefault_basic(t *testing.T) {
	resourceName := "aws_ebs_encryption_by_default.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEncryptionByDefaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSEncryptionByDefaultConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsEncryptionByDefault(resourceName, false),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSEncryptionByDefaultConfig(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsEncryptionByDefault(resourceName, true),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
		},
	})
}

func testAccCheckEncryptionByDefaultDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	response, err := conn.GetEbsEncryptionByDefault(&ec2.GetEbsEncryptionByDefaultInput{})
	if err != nil {
		return err
	}

	if aws.BoolValue(response.EbsEncryptionByDefault) != false {
		return fmt.Errorf("EBS encryption by default not disabled on resource removal")
	}

	return nil
}

func testAccCheckEbsEncryptionByDefault(n string, enabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		response, err := conn.GetEbsEncryptionByDefault(&ec2.GetEbsEncryptionByDefaultInput{})
		if err != nil {
			return err
		}

		if aws.BoolValue(response.EbsEncryptionByDefault) != enabled {
			return fmt.Errorf("EBS encryption by default is not in expected state (%t)", enabled)
		}

		return nil
	}
}

func testAccEBSEncryptionByDefaultConfig(enabled bool) string {
	return fmt.Sprintf(`
resource "aws_ebs_encryption_by_default" "test" {
  enabled = %[1]t
}
`, enabled)
}
