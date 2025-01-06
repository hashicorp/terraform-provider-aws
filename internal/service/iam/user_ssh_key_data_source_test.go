// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMUserSSHKeyDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_iam_user_ssh_key.test"
	dataSourceName := "data.aws_iam_user_ssh_key.test"

	username := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	publicKey, _, err := RandSSHKeyPairSize(2048, acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserSSHKeyDataSourceConfig_basic(username, publicKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "encoding", resourceName, "encoding"),
					resource.TestCheckResourceAttrPair(dataSourceName, "fingerprint", resourceName, "fingerprint"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPublicKey, resourceName, names.AttrPublicKey),
					resource.TestCheckResourceAttrPair(dataSourceName, "ssh_public_key_id", resourceName, "ssh_public_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrUsername, resourceName, names.AttrUsername),
				),
			},
		},
	})
}

func testAccUserSSHKeyDataSourceConfig_basic(username, publicKey string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
  path = "/"
}

resource "aws_iam_user_ssh_key" "test" {
  username   = aws_iam_user.test.name
  encoding   = "SSH"
  public_key = %[2]q
  status     = "Inactive"
}

data "aws_iam_user_ssh_key" "test" {
  username          = aws_iam_user.test.name
  encoding          = "SSH"
  ssh_public_key_id = aws_iam_user_ssh_key.test.ssh_public_key_id
}
`, username, publicKey)
}
