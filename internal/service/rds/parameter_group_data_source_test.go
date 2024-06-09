// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSParameterGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_db_parameter_group.test"
	resourceName := "aws_db_parameter_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccParameterGroupDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`not found`),
			},
			{
				Config: testAccParameterGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrFamily, resourceName, names.AttrFamily),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
				),
			},
		},
	})
}

const testAccParameterGroupDataSourceConfig_nonExistent = `
data "aws_db_parameter_group" "test" {
  name = "tf-acc-test-does-not-exist"
}
`

func testAccParameterGroupDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_parameter_group" "test" {
  name   = %[1]q
  family = "postgres12"

  parameter {
    name         = "client_encoding"
    value        = "UTF8"
    apply_method = "pending-reboot"
  }
}

data "aws_db_parameter_group" "test" {
  name = aws_db_parameter_group.test.name
}
`, rName)
}
