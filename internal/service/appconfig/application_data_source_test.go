// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appconfig_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppConfigApplicationDataSource_basic_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_appconfig_application.test"
	resourceName := "aws_appconfig_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppConfigEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDataSourceConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrID, regexache.MustCompile(`[a-z\d]{4,7}`)),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccAppConfigApplicationDataSource_basic_id(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_appconfig_application.test"
	resourceName := "aws_appconfig_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AppConfigEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationDataSourceConfig_id(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrID, regexache.MustCompile(`[a-z\d]{4,7}`)),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccApplicationDataSourceConfig_baseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_application" "test" {
  name        = %[1]q
  description = "Example AppConfig Application"
}
`, rName)
}

func testAccApplicationDataSourceConfig_name(rName string) string {
	return acctest.ConfigCompose(testAccApplicationDataSourceConfig_baseConfig(rName), `
data "aws_appconfig_application" "test" {
  name = aws_appconfig_application.test.name
}
`)
}

func testAccApplicationDataSourceConfig_id(rName string) string {
	return acctest.ConfigCompose(testAccApplicationDataSourceConfig_baseConfig(rName), `
data "aws_appconfig_application" "test" {
  id = aws_appconfig_application.test.id
}
`)
}
