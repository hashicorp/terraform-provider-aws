// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package outposts_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOutpostsOutpostInstanceTypesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_outposts_outpost_instance_types.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostInstanceTypesDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOutpostInstanceTypesAttributes(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckOutpostInstanceTypesAttributes(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourceName)
		}

		if v := rs.Primary.Attributes["instance_types.#"]; v == acctest.Ct0 {
			return fmt.Errorf("expected at least one instance_types result, got none")
		}

		return nil
	}
}

func testAccOutpostInstanceTypesDataSourceConfig_basic() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost_instance_types" "test" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}
`
}
