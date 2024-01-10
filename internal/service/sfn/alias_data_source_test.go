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

func TestAccSFNAliasDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_sfn_alias.test"
	resourceName := "aws_sfn_alias.test"
	rString := sdkacctest.RandString(8)
	stateMachineName := fmt.Sprintf("tf_acc_state_machine_alias_basic_%s", rString)
	aliasName := fmt.Sprintf("tf_acc_state_machine_alias_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sfn.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAliasDataSourceConfig_basic(stateMachineName, aliasName, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "creation_date", dataSourceName, "creation_date"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
				),
			},
		},
	})
}

func testAccAliasDataSourceConfig_basic(statemachineName string, aliasName string, rMaxAttempts int) string {
	return acctest.ConfigCompose(testAccStateMachineAliasConfig_basic(statemachineName, aliasName, rMaxAttempts), `
data "aws_sfn_alias" "test" {
  name             = aws_sfn_alias.test.name
  statemachine_arn = aws_sfn_state_machine.test.arn
}
`)
}
