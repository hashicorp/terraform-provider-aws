// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sfn_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSFNStateMachineDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_sfn_state_machine.test"
	resourceName := "aws_sfn_state_machine.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreationDate, dataSourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttrPair(resourceName, "definition", dataSourceName, "definition"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, dataSourceName, names.AttrRoleARN),
					resource.TestCheckResourceAttrPair(resourceName, "revision_id", dataSourceName, "revision_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrStatus, dataSourceName, names.AttrStatus),
				),
			},
		},
	})
}

func testAccStateMachineDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccStateMachineConfig_basic(rName, 5), `
data "aws_sfn_state_machine" "test" {
  name = aws_sfn_state_machine.test.name
}
`)
}
