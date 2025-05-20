// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccWorkSpacesPoolDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var pool workspaces.DescribeWorkspacesPoolsOutput

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	datasourceName := "data.aws_workspaces_pool.test"
	resourceName := "aws_workspaces_pool.test"
	resourceBundleName := "data.aws_workspaces_bundle.standard"
	resourceDirectory := "aws_workspaces_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(workspaces.ServiceID))
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(ctx, datasourceName, &pool),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "bundle_id", resourceBundleName, names.AttrID),
					resource.TestCheckResourceAttr(datasourceName, "capacity.0.desired_user_sessions", "1"),
					resource.TestCheckResourceAttr(datasourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttrPair(datasourceName, "directory_id", resourceDirectory, "directory_id"),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrID),
					resource.TestCheckResourceAttr(datasourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(datasourceName, names.AttrState),
					resource.TestCheckResourceAttr(datasourceName, "application_settings.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "application_settings.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(datasourceName, "application_settings.0.settings_group", "test"),
					resource.TestCheckResourceAttr(datasourceName, "timeout_settings.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "timeout_settings.0.disconnect_timeout_in_seconds", "2000"),
					resource.TestCheckResourceAttr(datasourceName, "timeout_settings.0.idle_disconnect_timeout_in_seconds", "2000"),
					resource.TestCheckResourceAttr(datasourceName, "timeout_settings.0.max_user_duration_in_seconds", "2000"),
				),
			},
		},
	})
}

func testAccPoolDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccPoolConfig_base(rName),
		fmt.Sprintf(`
resource "aws_workspaces_pool" "test" {
  application_settings {
    status         = "ENABLED"
    settings_group = "test"
  }
  bundle_id = data.aws_workspaces_bundle.standard.id
  capacity {
    desired_user_sessions = 1
  }
  description  = %[1]q
  directory_id = aws_workspaces_directory.test.directory_id
  name         = %[1]q
  timeout_settings {
    disconnect_timeout_in_seconds      = 2000
    idle_disconnect_timeout_in_seconds = 2000
    max_user_duration_in_seconds       = 2000
  }
}

data "aws_workspaces_pool" "test" {
  id = aws_workspaces_pool.test.id
}
`, rName))
}
