// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSFNActivityDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sfn_activity.test"
	dataSource1Name := "data.aws_sfn_activity.by_name"
	dataSource2Name := "data.aws_sfn_activity.by_arn"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActivityDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSource1Name, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreationDate, dataSource1Name, names.AttrCreationDate),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSource1Name, names.AttrName),

					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSource2Name, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreationDate, dataSource2Name, names.AttrCreationDate),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSource2Name, names.AttrName),
				),
			},
		},
	})
}

func testAccActivityDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource aws_sfn_activity "test" {
  name = %[1]q
}

data aws_sfn_activity "by_name" {
  name = aws_sfn_activity.test.name
}

data aws_sfn_activity "by_arn" {
  arn = aws_sfn_activity.test.id
}
`, rName)
}
