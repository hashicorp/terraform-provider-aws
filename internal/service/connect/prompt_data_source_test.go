// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccPromptDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	datasourceName := "data.aws_connect_prompt.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, connect.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptDataSourceConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttr(datasourceName, "name", "Beep.wav"),
					resource.TestCheckResourceAttrSet(datasourceName, "prompt_id"),
				),
			},
		},
	})
}

func testAccPromptBaseDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccPromptDataSourceConfig_name(rName string) string {
	return acctest.ConfigCompose(
		testAccPromptBaseDataSourceConfig(rName),
		`
data "aws_connect_prompt" "test" {
  instance_id = aws_connect_instance.test.id
  name        = "Beep.wav"
}
`)
}
