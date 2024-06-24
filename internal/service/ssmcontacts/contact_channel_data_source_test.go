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

func testAccContactChannelDataSource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	contactChannelResourceName := "aws_ssmcontacts_contact_channel.test"
	dataSourceName := "data.aws_ssmcontacts_contact_channel.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccContactPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMContactsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccContactChannelDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "activation_status", contactChannelResourceName, "activation_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "delivery_address.#", contactChannelResourceName, "delivery_address.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "delivery_address.0.simple_address", contactChannelResourceName, "delivery_address.0.simple_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, contactChannelResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrType, contactChannelResourceName, names.AttrType),
					resource.TestCheckResourceAttrPair(dataSourceName, "contact_id", contactChannelResourceName, "contact_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, contactChannelResourceName, names.AttrARN),
				),
			},
		},
	})
}

func testAccContactChannelDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssmincidents_replication_set" "test" {
  region {
    name = %[1]q
  }
}

resource "aws_ssmcontacts_contact" "test" {
  alias = "test-contact-for-%[2]s"
  type  = "PERSONAL"

  depends_on = [aws_ssmincidents_replication_set.test]
}

resource "aws_ssmcontacts_contact_channel" "test" {
  contact_id = aws_ssmcontacts_contact.test.arn

  delivery_address {
    simple_address = %[3]q
  }

  name = %[2]q
  type = "EMAIL"
}

data "aws_ssmcontacts_contact_channel" "test" {
  arn = aws_ssmcontacts_contact_channel.test.arn
}
`, acctest.Region(), rName, acctest.DefaultEmailAddress)
}
