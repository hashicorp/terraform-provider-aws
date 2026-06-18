// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectRouterConfigurationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	vifID := acctest.SkipIfEnvVarNotSet(t, "VIRTUAL_INTERFACE_ID")

	dataSourceName := "data.aws_dx_router_configuration.test"
	routerTypeIdentifier := "CiscoSystemsInc-2900SeriesRouters-IOS124"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DirectConnectEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRouterConfigurationDataSourceConfig_basic(vifID, routerTypeIdentifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "virtual_interface_id", vifID),
					resource.TestCheckResourceAttr(dataSourceName, "router_type_identifier", routerTypeIdentifier),
					resource.TestCheckResourceAttrSet(dataSourceName, "virtual_interface_name"),
					resource.TestCheckResourceAttr(dataSourceName, "router.0.platform", "2900 Series Routers"),
					resource.TestCheckResourceAttr(dataSourceName, "router.0.router_type_identifier", routerTypeIdentifier),
					resource.TestCheckResourceAttr(dataSourceName, "router.0.software", "IOS 12.4+"),
					resource.TestCheckResourceAttr(dataSourceName, "router.0.vendor", "Cisco Systems, Inc."),
					resource.TestCheckResourceAttr(dataSourceName, "router.0.xslt_template_name", "customer-router-cisco-generic.xslt"),
					resource.TestCheckResourceAttr(dataSourceName, "router.0.xslt_template_name_for_mac_sec", ""),
				),
			},
		},
	})
}

func testAccRouterConfigurationDataSourceConfig_basic(virtualInterfaceId, routerTypeIdentifier string) string {
	return fmt.Sprintf(`
data "aws_dx_router_configuration" "test" {
  virtual_interface_id   = %[1]q
  router_type_identifier = %[2]q
}
`, virtualInterfaceId, routerTypeIdentifier)
}
