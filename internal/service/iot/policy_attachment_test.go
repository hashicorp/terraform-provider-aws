// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTPolicyAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	policyName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_iot_policy_attachment.test1"
	resource2Name := "aws_iot_policy_attachment.test2"
	resource3Name := "aws_iot_policy_attachment.test3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyAttchmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyAttachmentConfig_basic(policyName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists(ctx, resource1Name),
				),
			},
			{
				Config: testAccPolicyAttachmentConfig_update1(policyName1, policyName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists(ctx, resource1Name),
					testAccCheckPolicyAttachmentExists(ctx, resource2Name),
				),
			},
			{
				Config: testAccPolicyAttachmentConfig_update2(policyName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists(ctx, resource2Name),
				),
			},
			{
				Config: testAccPolicyAttachmentConfig_update3(policyName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists(ctx, resource2Name),
					testAccCheckPolicyAttachmentExists(ctx, resource3Name),
				),
			},
		},
	})
}

func testAccCheckPolicyAttchmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_policy_attachment" {
				continue
			}

			_, err := tfiot.FindAttachedPolicyByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrPolicy], rs.Primary.Attributes[names.AttrTarget])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IoT Policy Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPolicyAttachmentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		_, err := tfiot.FindAttachedPolicyByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrPolicy], rs.Primary.Attributes[names.AttrTarget])

		return err
	}
}

func testAccPolicyAttachmentConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "test1" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_policy" "test1" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iot:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF

}

resource "aws_iot_policy_attachment" "test1" {
  policy = aws_iot_policy.test1.name
  target = aws_iot_certificate.test1.arn
}
`, rName)
}

func testAccPolicyAttachmentConfig_update1(policyName1, policyName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "test1" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_policy" "test1" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iot:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_iot_policy" "test2" {
  name = %[2]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iot:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF

}

resource "aws_iot_policy_attachment" "test1" {
  policy = aws_iot_policy.test1.name
  target = aws_iot_certificate.test1.arn
}

resource "aws_iot_policy_attachment" "test2" {
  policy = aws_iot_policy.test2.name
  target = aws_iot_certificate.test1.arn
}
`, policyName1, policyName2)
}

func testAccPolicyAttachmentConfig_update2(policyName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "test1" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_policy" "test2" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iot:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_iot_policy_attachment" "test2" {
  policy = aws_iot_policy.test2.name
  target = aws_iot_certificate.test1.arn
}
`, policyName2)
}

func testAccPolicyAttachmentConfig_update3(policyName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_certificate" "test1" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_certificate" "test2" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}

resource "aws_iot_policy" "test2" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iot:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_iot_policy_attachment" "test2" {
  policy = aws_iot_policy.test2.name
  target = aws_iot_certificate.test1.arn
}

resource "aws_iot_policy_attachment" "test3" {
  policy = aws_iot_policy.test2.name
  target = aws_iot_certificate.test2.arn
}
`, policyName2)
}
