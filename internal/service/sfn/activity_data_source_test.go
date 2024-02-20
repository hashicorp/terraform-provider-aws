// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/sfn"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSFNActivityDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sfn_activity.test"
	dataSource1Name := "data.aws_sfn_activity.by_name"
	dataSource2Name := "data.aws_sfn_activity.by_arn"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sfn.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActivityDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSource1Name, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "creation_date", dataSource1Name, "creation_date"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSource1Name, "name"),

					resource.TestCheckResourceAttrPair(resourceName, "id", dataSource2Name, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "creation_date", dataSource2Name, "creation_date"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSource2Name, "name"),
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
