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

func TestAccInterconnectEnvironmentDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_interconnect_environment.test"
	environmentID := acctest.SkipIfEnvVarNotSet(t, "INTERCONNECT_ENVIRONMENT_ID")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InterconnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentDataSourceConfig_basic(environmentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "environment_id", environmentID),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrLocation),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrState),
					resource.TestCheckResourceAttrSet(dataSourceName, "interconnect_provider"),
				),
			},
		},
	})
}

func testAccEnvironmentDataSourceConfig_basic(environmentID string) string {
	return fmt.Sprintf(`
data "aws_interconnect_environment" "test" {
  environment_id = %[1]q
}
`, environmentID)
}
