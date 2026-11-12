// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interconnect_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInterconnectEnvironmentsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_interconnect_environments.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InterconnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentsDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "environments.#"),
				),
			},
		},
	})
}

func TestAccInterconnectEnvironmentsDataSource_providerFilter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_interconnect_environments.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InterconnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentsDataSourceConfig_providerFilter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "environments.#"),
					resource.TestCheckResourceAttr(dataSourceName, "interconnect_provider.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "interconnect_provider.0.cloud_service_provider", "gcp"),
				),
			},
		},
	})
}

func testAccEnvironmentsDataSourceConfig_basic() string {
	return `
data "aws_interconnect_environments" "test" {}
`
}

func testAccEnvironmentsDataSourceConfig_providerFilter() string {
	return `
data "aws_interconnect_environments" "test" {
  interconnect_provider {
    cloud_service_provider = "gcp"
  }
}
`
}
