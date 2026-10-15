// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package savingsplans_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSavingsPlansOfferingsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_savingsplans_offerings.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.SavingsPlansServiceID),
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOfferingsDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "product_type", "EC2"),
					resource.TestCheckResourceAttrSet(dataSourceName, "offerings.#"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("offerings"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"currency":            knownvalue.StringExact("USD"),
							names.AttrDescription: knownvalue.StringExact("1 year All Upfront a1 EC2 Instance Savings Plan in us-west-2"), // lintignore:AWSAT003
							"duration_seconds":    knownvalue.Int64Exact(31536000),
							"offering_id":         knownvalue.NotNull(),
							"operation":           knownvalue.StringExact(""),
							"payment_option":      knownvalue.StringExact("All Upfront"),
							"plan_type":           knownvalue.StringExact("EC2Instance"),
							"product_types":       knownvalue.SetExact([]knownvalue.Check{knownvalue.StringExact("EC2")}),
							names.AttrProperties: knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									names.AttrName:  knownvalue.StringExact("instanceFamily"),
									names.AttrValue: knownvalue.StringExact("a1"),
								}),
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									names.AttrName:  knownvalue.StringExact(names.AttrRegion),
									names.AttrValue: knownvalue.StringExact("us-west-2"), // lintignore:AWSAT003
								}),
							}),
							"service_code": knownvalue.StringExact("ComputeSavingsPlans"),
							"usage_type":   knownvalue.StringExact("USW2-EC2SP:a1.1yrAllUpfront"),
						}),
					})),
				},
			},
		},
	})
}

func testAccOfferingsDataSourceConfig_basic() string {
	// lintignore:AWSAT003
	return `
data "aws_savingsplans_offerings" "test" {
  product_type = "EC2"
  usage_types  = ["USW2-EC2SP:a1.1yrAllUpfront"]
  durations    = [31536000]

  filter {
    name   = "region"
    values = ["us-west-2"]
  }
}
`
}
