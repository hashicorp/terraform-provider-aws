// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueConnectionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_connection.test"
	datasourceName := "data.aws_glue_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	jdbcConnectionUrl := fmt.Sprintf("jdbc:mysql://%s/testdatabase", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionDataSourceConfig_basic(rName, jdbcConnectionUrl),
				Check: resource.ComposeTestCheckFunc(
					testAccConnectionCheckDataSource(datasourceName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrCatalogID, resourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_type", resourceName, "connection_type"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_properties", resourceName, "connection_properties"),
					resource.TestCheckResourceAttrPair(datasourceName, "physical_connection_requirements", resourceName, "physical_connection_requirements"),
					resource.TestCheckResourceAttrPair(datasourceName, "match_criteria", resourceName, "match_criteria"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrTags, resourceName, names.AttrTags),
				),
			},
		},
	})
}

func testAccConnectionCheckDataSource(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		return nil
	}
}

func testAccConnectionDataSourceConfig_basic(rName, jdbcConnectionUrl string) string {
	return fmt.Sprintf(`
resource "aws_glue_connection" "test" {
  name = %[1]q

  connection_properties = {
    JDBC_CONNECTION_URL = %[2]q
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }

}

data "aws_glue_connection" "test" {
  id = aws_glue_connection.test.id
}
`, rName, jdbcConnectionUrl)
}
