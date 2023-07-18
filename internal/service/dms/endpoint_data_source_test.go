// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"fmt"
	"testing"

	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDMSEndpointDataSource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dms_endpoint.test"
	dataSourceName := "data.aws_dms_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, dms.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint_id", resourceName, "endpoint_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint_type", resourceName, "endpoint_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_name", resourceName, "engine_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "port", resourceName, "port"),
					resource.TestCheckResourceAttrPair(dataSourceName, "server_name", resourceName, "server_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ssl_mode", resourceName, "ssl_mode"),
				),
			},
		},
	})
}

func testAccEndpointDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_endpoint" "test" {
  database_name = "tf-test-dms-db"
  endpoint_id   = %[1]q
  endpoint_type = "source"
  engine_name   = "aurora"
  password      = "tftestpw"
  port          = 3306
  server_name   = "tftest"
  ssl_mode      = "none"

  username = "tftest"
}

data "aws_dms_endpoint" "test" {
  endpoint_id = aws_dms_endpoint.test.endpoint_id
}
`, rName)
}
