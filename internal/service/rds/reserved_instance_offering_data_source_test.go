// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSInstanceOffering_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_reserved_instance_offering.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceOfferingConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "currency_code"),
					resource.TestCheckResourceAttr(dataSourceName, "db_instance_class", "db.t2.micro"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDuration, "31536000"),
					resource.TestCheckResourceAttrSet(dataSourceName, "fixed_price"),
					resource.TestCheckResourceAttr(dataSourceName, "multi_az", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(dataSourceName, "offering_id"),
					resource.TestCheckResourceAttr(dataSourceName, "offering_type", "All Upfront"),
					resource.TestCheckResourceAttr(dataSourceName, "product_description", "mysql"),
				),
			},
		},
	})
}

func testAccInstanceOfferingConfig_basic() string {
	return `
data "aws_rds_reserved_instance_offering" "test" {
  db_instance_class   = "db.t2.micro"
  duration            = 31536000
  multi_az            = false
  offering_type       = "All Upfront"
  product_description = "mysql"
}
`
}
