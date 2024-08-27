// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppConfigEnvironmentsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_appconfig_environments.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppConfigEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentsDataSourceConfig_basic(appName, rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "environment_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "environment_ids.*", "aws_appconfig_environment.test_1", "environment_id"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "environment_ids.*", "aws_appconfig_environment.test_2", "environment_id"),
				),
			},
		},
	})
}

func testAccEnvironmentsDataSourceConfig_basic(appName, rName1, rName2 string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_name(appName),
		fmt.Sprintf(`
resource "aws_appconfig_environment" "test_1" {
  application_id = aws_appconfig_application.test.id
  name           = %[1]q
}

resource "aws_appconfig_environment" "test_2" {
  application_id = aws_appconfig_application.test.id
  name           = %[2]q
}

data "aws_appconfig_environments" "test" {
  application_id = aws_appconfig_application.test.id
  depends_on     = [aws_appconfig_environment.test_1, aws_appconfig_environment.test_2]
}
`, rName1, rName2))
}
