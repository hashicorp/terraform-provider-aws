// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPromptDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	datasourceName := "data.aws_connect_prompt.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptDataSourceConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrRegionalARNFormat(ctx, datasourceName, names.AttrARN, "connect", "instance/{instance_id}/prompt/{prompt_id}"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttr(datasourceName, names.AttrName, "Beep.wav"),
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
