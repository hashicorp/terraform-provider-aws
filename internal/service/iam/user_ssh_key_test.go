// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMUserSSHKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.SSHPublicKey
	resourceName := "aws_iam_user_ssh_key.user"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	publicKey, _, err := RandSSHKeyPairSize(2048, acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserSSHKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserSSHKeyConfig_encoding(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserSSHKeyExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Inactive"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccUserSSHKeyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIAMUserSSHKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.SSHPublicKey
	resourceName := "aws_iam_user_ssh_key.user"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	publicKey, _, err := RandSSHKeyPairSize(2048, acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserSSHKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserSSHKeyConfig_encoding(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserSSHKeyExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceUserSSHKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMUserSSHKey_pemEncoding(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.SSHPublicKey
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user_ssh_key.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserSSHKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSSHKeyConfig_pemEncoding(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserSSHKeyExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "Active"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccUserSSHKeyImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckUserSSHKeyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_user_ssh_key" {
				continue
			}

			_, err := tfiam.FindSSHPublicKeyByThreePartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["encoding"], rs.Primary.Attributes[names.AttrUsername])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM User SSH Key %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckUserSSHKeyExists(ctx context.Context, n string, v *awstypes.SSHPublicKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IAM User SSH Key ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		output, err := tfiam.FindSSHPublicKeyByThreePartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["encoding"], rs.Primary.Attributes[names.AttrUsername])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccUserSSHKeyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		username := rs.Primary.Attributes[names.AttrUsername]
		sshPublicKeyId := rs.Primary.Attributes["ssh_public_key_id"]
		encoding := rs.Primary.Attributes["encoding"]

		return fmt.Sprintf("%s:%s:%s", username, sshPublicKeyId, encoding), nil
	}
}

func testAccUserSSHKeyConfig_encoding(rName, publicKey string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
  name = %[1]q
  path = "/"
}

resource "aws_iam_user_ssh_key" "user" {
  username   = aws_iam_user.user.name
  encoding   = "SSH"
  public_key = %[2]q
  status     = "Inactive"
}
`, rName, publicKey)
}

func testAccSSHKeyConfig_pemEncoding(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
  name = %[1]q
  path = "/"
}

resource "aws_iam_user_ssh_key" "user" {
  username = aws_iam_user.user.name
  encoding = "PEM"

  public_key = <<EOF
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA9xercjxBRM1dC191/AbF
3TLEM9cdnBIpCgxGNGiI+NaoMTAj/4rXp3ql0iBWQaeb4sz72qCEd1JvcSuzxqFv
IIrqRp/hD7sSAOHAzOL8zqjpIjD4c+VytMIRI5Fc06OPktKbrw2bsCLHYlvZsYSX
O7YATS9HGJVkmFZM+Bv37JTX0T1uZmADOPX+H4bcT2+aJOENi4PXTylRzvwYHruc
KDHO0WNKdXo+g+AihROpcpkgyaVtGB1/8KhPfnHxGroe8WXBtKvbdrWuhen5l9Go
L6RcmaPGhW13lAa+6LEgiTYL2r1mzP9Op4lqzr2F9scFnYV5l0q21/GW2m1aIQSu
NQIDAQAB
-----END PUBLIC KEY-----
EOF
}
`, rName)
}
