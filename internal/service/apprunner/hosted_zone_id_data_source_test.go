// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppRunnerHostedZoneIDDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_apprunner_hosted_zone_id.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppRunner)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunner),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccHostedZoneIDDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrID),
				),
			},
			{
				Config: testAccHostedZoneIDDataSourceConfig_explicitRegion(names.APNortheast1RegionID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, names.AttrID, "Z08491812XW6IPYLR6CCA"),
				),
			},
		},
	})
}

const testAccHostedZoneIDDataSourceConfig_basic = `
data "aws_apprunner_hosted_zone_id" "test" {}
`

func testAccHostedZoneIDDataSourceConfig_explicitRegion(region string) string {
	return fmt.Sprintf(`
data "aws_apprunner_hosted_zone_id" "test" {
  region = %[1]q
}
`, region)
}
