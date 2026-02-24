// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_group.test"
	dataSourceName := "data.aws_quicksight_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig(rName, "text1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrGroupName, resourceName, names.AttrGroupName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDescription, "text1"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrNamespace, tfquicksight.DefaultNamespace),
					resource.TestCheckResourceAttrSet(dataSourceName, "principal_id"),
				),
			},
		},
	})
}

func testAccGroupDataSourceConfig(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_group" "test" {
  group_name  = %[1]q
  description = %[2]q
}

data "aws_quicksight_group" "test" {
  group_name = aws_quicksight_group.test.group_name
}
`, rName, description)
}
