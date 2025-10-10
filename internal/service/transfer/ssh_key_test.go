// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccSSHKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.SshPublicKey
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_ssh_key.test"
	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSSHKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSSHKeyConfig_basic(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSSHKeyExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "body", publicKey),
					resource.TestCheckResourceAttrPair(resourceName, "server_id", "aws_transfer_server.test", names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "ssh_key_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, "aws_transfer_user.test", names.AttrUserName),
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

func testAccSSHKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.SshPublicKey
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_ssh_key.test"
	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSSHKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSSHKeyConfig_basic(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSSHKeyExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceSSHKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSSHKeyExists(ctx context.Context, n string, v *awstypes.SshPublicKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferClient(ctx)

		_, output, err := tftransfer.FindUserSSHKeyByThreePartKey(ctx, conn, rs.Primary.Attributes["server_id"], rs.Primary.Attributes[names.AttrUserName], rs.Primary.Attributes["ssh_key_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSSHKeyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_ssh_key" {
				continue
			}

			_, _, err := tftransfer.FindUserSSHKeyByThreePartKey(ctx, conn, rs.Primary.Attributes["server_id"], rs.Primary.Attributes[names.AttrUserName], rs.Primary.Attributes["ssh_key_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Transfer SSH Key %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccSSHKeyConfig_basic(rName, publicKey string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "test" {
  identity_provider_type = "SERVICE_MANAGED"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoS3",
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_transfer_user" "test" {
  server_id = aws_transfer_server.test.id
  user_name = "tftestuser"
  role      = aws_iam_role.test.arn
}

resource "aws_transfer_ssh_key" "test" {
  server_id = aws_transfer_server.test.id
  user_name = aws_transfer_user.test.user_name
  body      = "%[2]s"
}
`, rName, publicKey)
}
