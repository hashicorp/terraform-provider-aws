// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccContactDataSource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssmcontacts_contact.contact_one"
	dataSourceName := "data.aws_ssmcontacts_contact.contact_one"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccContactDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAlias, dataSourceName, names.AttrAlias),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrType, dataSourceName, names.AttrType),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDisplayName, dataSourceName, names.AttrDisplayName),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsKey1, dataSourceName, acctest.CtTagsKey1),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsKey2, dataSourceName, acctest.CtTagsKey2),
				),
			},
		},
	})
}

func testAccContactDataSourceConfig_base() string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[1]q
  }
}
`, acctest.Region())
}

func testAccContactDataSourceConfig_basic(alias string) string {
	return acctest.ConfigCompose(
		testAccContactDataSourceConfig_base(),
		fmt.Sprintf(`
resource "aws_ssmcontacts_contact" "contact_one" {
  alias        = %[1]q
  display_name = %[1]q
  type         = "PERSONAL"

  tags = {
    key1 = "tag1"
    key2 = "tag2"
  }

  depends_on = [aws_ssmincidents_replication_set.test]
}

data "aws_ssmcontacts_contact" "contact_one" {
  arn = aws_ssmcontacts_contact.contact_one.arn
}
`, alias))
}
