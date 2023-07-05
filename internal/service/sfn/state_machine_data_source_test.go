// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/sfn"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSFNStateMachineDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_sfn_state_machine.test"
	resourceName := "aws_sfn_state_machine.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, sfn.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStateMachineDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "creation_date", dataSourceName, "creation_date"),
					resource.TestCheckResourceAttrPair(resourceName, "definition", dataSourceName, "definition"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", dataSourceName, "role_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "revision_id", dataSourceName, "revision_id"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
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
