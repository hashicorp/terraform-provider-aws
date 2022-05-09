package connect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccConnectPromptDataSource_name(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	datasourceName := "data.aws_connect_prompt.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPromptDataSourceConfig_Name(rName),
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

func testAccPromptDataSourceConfig_Name(rName string) string {
	return acctest.ConfigCompose(
		testAccPromptBaseDataSourceConfig(rName),
		`
data "aws_connect_prompt" "test" {
  instance_id = aws_connect_instance.test.id
  name        = "Beep.wav"
}
`)
}
