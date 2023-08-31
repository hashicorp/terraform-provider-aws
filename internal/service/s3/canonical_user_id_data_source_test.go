// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccS3CanonicalUserIDDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, s3.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCanonicalUserIDDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCanonicalUserIdCheckExistsDataSource("data.aws_canonical_user_id.current"),
				),
			},
		},
	})
}

func testAccCanonicalUserIdCheckExistsDataSource(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Can't find Canonical User ID resource: %s", name)
		}

		if rs.Primary.Attributes["id"] == "" {
			return fmt.Errorf("Missing Canonical User ID")
		}
		if rs.Primary.Attributes["display_name"] == "" {
			return fmt.Errorf("Missing Display Name")
		}

		return nil
	}
}

const testAccCanonicalUserIDDataSourceConfig_basic = `
data "aws_canonical_user_id" "current" {}
`
