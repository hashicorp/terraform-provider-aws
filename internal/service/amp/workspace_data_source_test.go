// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/prometheusservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAMPWorkspaceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_prometheus_workspace.test"
	dataSourceName := "data.aws_prometheus_workspace.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, prometheusservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		CheckDestroy:             nil,
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceDataSourceConfig_alias(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "alias", dataSourceName, "alias"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "created_date"),
					resource.TestCheckResourceAttrPair(resourceName, "prometheus_endpoint", dataSourceName, "prometheus_endpoint"),
					resource.TestCheckResourceAttrSet(dataSourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccWorkspaceDataSourceConfig_alias(rName string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
  alias = %[1]q
}

data "aws_prometheus_workspace" "test" {
  workspace_id = aws_prometheus_workspace.test.id
}
`, rName)
}
