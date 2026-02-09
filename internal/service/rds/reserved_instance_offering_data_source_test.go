// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	testInstanceClass = "db.r5.large"
)

func TestAccRDSReservedInstanceOffering_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_reserved_instance_offering.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccReservedInstanceOfferingConfig_basic(testInstanceClass, "postgresql"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "currency_code"),
					resource.TestCheckResourceAttr(dataSourceName, "db_instance_class", testInstanceClass),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDuration, "31536000"),
					resource.TestCheckResourceAttrSet(dataSourceName, "fixed_price"),
					resource.TestCheckResourceAttr(dataSourceName, "multi_az", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(dataSourceName, "offering_id"),
					resource.TestCheckResourceAttr(dataSourceName, "offering_type", "All Upfront"),
					resource.TestCheckResourceAttr(dataSourceName, "product_description", "postgresql"),
				),
			},
			{
				Config: testAccReservedInstanceOfferingConfig_basic(testInstanceClass, "aurora-postgresql"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "currency_code"),
					resource.TestCheckResourceAttr(dataSourceName, "db_instance_class", testInstanceClass),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDuration, "31536000"),
					resource.TestCheckResourceAttrSet(dataSourceName, "fixed_price"),
					resource.TestCheckResourceAttr(dataSourceName, "multi_az", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(dataSourceName, "offering_id"),
					resource.TestCheckResourceAttr(dataSourceName, "offering_type", "All Upfront"),
					resource.TestCheckResourceAttr(dataSourceName, "product_description", "aurora-postgresql"),
				),
			},
		},
	})
}

func TestAccRDSReservedInstanceOffering_mysql(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_rds_reserved_instance_offering.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccReservedInstanceOfferingConfig_basic(testInstanceClass, "mysql"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "currency_code"),
					resource.TestCheckResourceAttr(dataSourceName, "db_instance_class", testInstanceClass),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDuration, "31536000"),
					resource.TestCheckResourceAttrSet(dataSourceName, "fixed_price"),
					resource.TestCheckResourceAttr(dataSourceName, "multi_az", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(dataSourceName, "offering_id"),
					resource.TestCheckResourceAttr(dataSourceName, "offering_type", "All Upfront"),
					resource.TestCheckResourceAttr(dataSourceName, "product_description", "mysql"),
				),
			},
			{
				Config: testAccReservedInstanceOfferingConfig_basic(testInstanceClass, "aurora-mysql"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "currency_code"),
					resource.TestCheckResourceAttr(dataSourceName, "db_instance_class", testInstanceClass),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDuration, "31536000"),
					resource.TestCheckResourceAttrSet(dataSourceName, "fixed_price"),
					resource.TestCheckResourceAttr(dataSourceName, "multi_az", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(dataSourceName, "offering_id"),
					resource.TestCheckResourceAttr(dataSourceName, "offering_type", "All Upfront"),
					resource.TestCheckResourceAttr(dataSourceName, "product_description", "aurora-mysql"),
				),
			},
		},
	})
}

func testAccReservedInstanceOfferingConfig_basic(class, desc string) string {
	return fmt.Sprintf(`
data "aws_rds_reserved_instance_offering" "test" {
  db_instance_class   = %[1]q
  duration            = 31536000
  multi_az            = false
  offering_type       = "All Upfront"
  product_description = %[2]q
}
`, class, desc)
}
