// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAMPScraperConfigurationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_prometheus_scraper_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccScraperConfigurationDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "default"),
				),
			},
		},
	})
}

func testAccScraperConfigurationDataSourceConfig_basic() string {
	return `
data "aws_prometheus_scraper_configuration" "test" {}
`
}
