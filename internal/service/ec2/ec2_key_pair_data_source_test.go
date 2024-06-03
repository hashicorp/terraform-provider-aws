// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2KeyPairDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSource1Name := "data.aws_key_pair.by_id"
	dataSource2Name := "data.aws_key_pair.by_name"
	dataSource3Name := "data.aws_key_pair.by_filter"
	resourceName := "aws_key_pair.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairDataSourceConfig_basic(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSource1Name, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSource1Name, "fingerprint", resourceName, "fingerprint"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "key_name", resourceName, "key_name"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "key_pair_id", resourceName, "key_pair_id"),
					resource.TestCheckResourceAttrPair(dataSource1Name, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),

					resource.TestCheckResourceAttrPair(dataSource2Name, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSource2Name, "fingerprint", resourceName, "fingerprint"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "key_name", resourceName, "key_name"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "key_pair_id", resourceName, "key_pair_id"),
					resource.TestCheckResourceAttrPair(dataSource2Name, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),

					resource.TestCheckResourceAttrPair(dataSource3Name, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSource3Name, "fingerprint", resourceName, "fingerprint"),
					resource.TestCheckResourceAttrPair(dataSource3Name, "key_name", resourceName, "key_name"),
					resource.TestCheckResourceAttrPair(dataSource3Name, "key_pair_id", resourceName, "key_pair_id"),
					resource.TestCheckResourceAttrPair(dataSource3Name, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccEC2KeyPairDataSource_includePublicKey(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSource1Name := "data.aws_key_pair.by_name"
	resourceName := "aws_key_pair.test"

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPairDataSourceConfig_includePublicKey(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSource1Name, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSource1Name, "fingerprint", resourceName, "fingerprint"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "key_name", resourceName, "key_name"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "key_pair_id", resourceName, "key_pair_id"),
					resource.TestCheckResourceAttrWith(dataSource1Name, names.AttrPublicKey, func(v string) error {
						if !tfec2.OpenSSHPublicKeysEqual(v, publicKey) {
							return fmt.Errorf("Attribute 'public_key' expected %q, not equal to %q", publicKey, v)
						}

						return nil
					}),
					resource.TestCheckResourceAttrPair(dataSource1Name, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccKeyPairDataSourceConfig_basic(rName, publicKey string) string {
	return fmt.Sprintf(`
data "aws_key_pair" "by_name" {
  key_name = aws_key_pair.test.key_name
}

data "aws_key_pair" "by_id" {
  key_pair_id = aws_key_pair.test.key_pair_id
}

data "aws_key_pair" "by_filter" {
  filter {
    name   = "tag:Name"
    values = [aws_key_pair.test.tags["Name"]]
  }
}

resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, publicKey)
}

func testAccKeyPairDataSourceConfig_includePublicKey(rName, publicKey string) string {
	return fmt.Sprintf(`
data "aws_key_pair" "by_name" {
  key_name           = aws_key_pair.test.key_name
  include_public_key = true
}

resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}
`, rName, publicKey)
}
