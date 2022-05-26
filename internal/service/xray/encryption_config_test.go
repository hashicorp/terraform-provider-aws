package xray_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/xray"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccXRayEncryptionConfig_basic(t *testing.T) {
	var EncryptionConfig xray.EncryptionConfig
	resourceName := "aws_xray_encryption_config.test"
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, xray.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptionConfigConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEncryptionConfigExists(resourceName, &EncryptionConfig),
					resource.TestCheckResourceAttr(resourceName, "type", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEncryptionConfigConfig_key(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEncryptionConfigExists(resourceName, &EncryptionConfig),
					resource.TestCheckResourceAttr(resourceName, "type", "KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "key_id", keyResourceName, "arn"),
				),
			},
			{
				Config: testAccEncryptionConfigConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEncryptionConfigExists(resourceName, &EncryptionConfig),
					resource.TestCheckResourceAttr(resourceName, "type", "NONE"),
				),
			},
		},
	})
}

func testAccCheckEncryptionConfigExists(n string, EncryptionConfig *xray.EncryptionConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No XRay Encryption Config ID is set")
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).XRayConn

		config, err := conn.GetEncryptionConfig(&xray.GetEncryptionConfigInput{})

		if err != nil {
			return err
		}

		*EncryptionConfig = *config.EncryptionConfig

		return nil
	}
}

func testAccEncryptionConfigConfig_basic() string {
	return `
resource "aws_xray_encryption_config" "test" {
  type = "NONE"
}
`
}

func testAccEncryptionConfigConfig_key() string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Terraform acc test %s"
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_xray_encryption_config" "test" {
  type   = "KMS"
  key_id = aws_kms_key.test.arn
}
`, sdkacctest.RandString(8))
}
