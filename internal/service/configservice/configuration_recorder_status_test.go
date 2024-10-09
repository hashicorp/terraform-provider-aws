// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconfig "github.com/hashicorp/terraform-provider-aws/internal/service/configservice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccConfigurationRecorderStatus_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var crs types.ConfigurationRecorderStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_recorder_status.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationRecorderStatusDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationRecorderStatusConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationRecorderStatusExists(ctx, resourceName, &crs),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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

func testAccConfigurationRecorderStatus_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var crs types.ConfigurationRecorderStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_recorder_status.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationRecorderStatusDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationRecorderStatusConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationRecorderStatusExists(ctx, resourceName, &crs),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconfig.ResourceConfigurationRecorder(), "aws_config_configuration_recorder.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccConfigurationRecorderStatus_startEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var crs types.ConfigurationRecorderStatus
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_config_configuration_recorder_status.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConfigServiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationRecorderStatusDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationRecorderStatusConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationRecorderStatusExists(ctx, resourceName, &crs),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationRecorderStatusConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationRecorderStatusExists(ctx, resourceName, &crs),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccConfigurationRecorderStatusConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationRecorderStatusExists(ctx, resourceName, &crs),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckConfigurationRecorderStatusExists(ctx context.Context, n string, v *types.ConfigurationRecorderStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		output, err := tfconfig.FindConfigurationRecorderStatusByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckConfigurationRecorderStatusDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_config_configuration_recorder_status" {
				continue
			}

			_, err := tfconfig.FindConfigurationRecorderStatusByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ConfigService Configuration Recorder Status %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccConfigurationRecorderStatusConfig_basic(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_config_configuration_recorder" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_config_delivery_channel" "test" {
  name           = %[1]q
  s3_bucket_name = aws_s3_bucket.test.bucket
}

resource "aws_config_configuration_recorder_status" "test" {
  name       = aws_config_configuration_recorder.test.name
  is_enabled = %[2]t
  depends_on = [aws_config_delivery_channel.test]
}
`, rName, enabled)
}
