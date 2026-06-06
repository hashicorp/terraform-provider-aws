// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bcmdashboards_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/bcmdashboards/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBCMDashboardsDashboardDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_bcmdashboards_dashboard.test"
	resourceName := "aws_bcmdashboards_dashboard.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BCMDashboardsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDashboardDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "dashboard_type", string(awstypes.DashboardTypeCustom)),
					resource.TestCheckResourceAttr(dataSourceName, "widget.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "widget.0.title", "example"),
				),
			},
		},
	})
}

func testAccDashboardDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDashboardConfig_basic(rName), `
data "aws_bcmdashboards_dashboard" "test" {
  arn = aws_bcmdashboards_dashboard.test.arn
}
`)
}
