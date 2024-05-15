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

func TestAccOutpostsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_outposts_outposts.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOutpostsAttributes(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckOutpostsAttributes(dataSourceName string) resource.TestCheckFunc { // nosemgrep:ci.outposts-in-func-name
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourceName)
		}

		if v := rs.Primary.Attributes["arns.#"]; v == acctest.Ct0 {
			return fmt.Errorf("expected at least one arns result, got none")
		}

		if v := rs.Primary.Attributes["ids.#"]; v == acctest.Ct0 {
			return fmt.Errorf("expected at least one ids result, got none")
		}

		return nil
	}
}

func testAccOutpostsDataSourceConfig_basic() string { // nosemgrep:ci.outposts-in-func-name
	return `
data "aws_outposts_outposts" "test" {}
`
}
