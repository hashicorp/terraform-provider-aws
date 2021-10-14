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
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSXrayEncryptionConfig_basic(t *testing.T) {
	var EncryptionConfig xray.EncryptionConfig
	resourceName := "aws_xray_encryption_config.test"
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, xray.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSXrayEncryptionConfigBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXrayEncryptionConfigExists(resourceName, &EncryptionConfig),
					resource.TestCheckResourceAttr(resourceName, "type", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSXrayEncryptionConfigWithKeyConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXrayEncryptionConfigExists(resourceName, &EncryptionConfig),
					resource.TestCheckResourceAttr(resourceName, "type", "KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "key_id", keyResourceName, "arn"),
				),
			},
			{
				Config: testAccAWSXrayEncryptionConfigBasicConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXrayEncryptionConfigExists(resourceName, &EncryptionConfig),
					resource.TestCheckResourceAttr(resourceName, "type", "NONE"),
				),
			},
		},
	})
}

func testAccCheckXrayEncryptionConfigExists(n string, EncryptionConfig *xray.EncryptionConfig) resource.TestCheckFunc {
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

func testAccAWSXrayEncryptionConfigBasicConfig() string {
	return `
resource "aws_xray_encryption_config" "test" {
  type = "NONE"
}
`
}

func testAccAWSXrayEncryptionConfigWithKeyConfig() string {
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
