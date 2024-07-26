// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSFNStateMachineVersionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_sfn_state_machine_versions.test"
	resourceName := "aws_sfn_state_machine.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SFNServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineVersionsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, "statemachine_arn"),
					resource.TestCheckResourceAttr(dataSourceName, "statemachine_versions.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccStateMachineVersionsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccStateMachineAliasConfig_base(rName, 5), `
data "aws_sfn_state_machine_versions" "test" {
  statemachine_arn = aws_sfn_state_machine.test.arn
}
`)
}
