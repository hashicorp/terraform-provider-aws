// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interconnect_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInterconnectAttachPointsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_interconnect_attach_points.test"
	environmentsDataSourceName := "data.aws_interconnect_environments.test"
	region := testAccRegion()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckEnvironments(ctx, t, region) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InterconnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAttachPointsDataSourceConfig_basic(region),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "environment_id", environmentsDataSourceName, "environments.0.environment_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "attach_points.#"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRegion, region),
				),
			},
		},
	})
}

func testAccAttachPointsDataSourceConfig_basic(region string) string {
	return fmt.Sprintf(`
data "aws_interconnect_environments" "test" {
  region = %[1]q
}

data "aws_interconnect_attach_points" "test" {
  region         = %[1]q
  environment_id = data.aws_interconnect_environments.test.environments[0].environment_id
}
`, region)
}
