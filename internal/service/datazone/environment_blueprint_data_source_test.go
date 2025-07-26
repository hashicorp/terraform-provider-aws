// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataZoneEnvironmentBlueprintDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_datazone_environment_blueprint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(dataSourceName, "blueprint_provider"),
				),
			},
		},
	})
}

func testAccEnvironmentBlueprintDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDomainConfig_basic(rName),
		`
data "aws_datazone_environment_blueprint" "test" {
  domain_id = aws_datazone_domain.test.id
  name      = "DefaultDataLake"
  managed   = true
}
`,
	)
}
