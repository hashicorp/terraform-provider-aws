// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package xray_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/xray/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfxray "github.com/hashicorp/terraform-provider-aws/internal/service/xray"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccXRayEncryptionConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.EncryptionConfig
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_xray_encryption_config.test"
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.XRayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptionConfigConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEncryptionConfigExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEncryptionConfigConfig_key(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEncryptionConfigExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "KMS"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKeyID, keyResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccEncryptionConfigConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEncryptionConfigExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "NONE"),
				),
			},
		},
	})
}

func testAccCheckEncryptionConfigExists(ctx context.Context, n string, v *types.EncryptionConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No XRay Encryption Config ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).XRayClient(ctx)

		output, err := tfxray.FindEncryptionConfig(ctx, conn)

		if err != nil {
			return err
		}

		*v = *output

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

func testAccEncryptionConfigConfig_key(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
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
`, rName)
}
