// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectRouterConfigurationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	key := "VIRTUAL_INTERFACE_ID"
	virtualInterfaceId := os.Getenv(key)
	if virtualInterfaceId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	dataSourceName := "data.aws_dx_router_configuration.test"
	routerTypeIdentifier := "CiscoSystemsInc-2900SeriesRouters-IOS124"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, directconnect.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRouterConfigurationDataSourceConfig_basic(virtualInterfaceId, routerTypeIdentifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "virtual_interface_id", virtualInterfaceId),
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
