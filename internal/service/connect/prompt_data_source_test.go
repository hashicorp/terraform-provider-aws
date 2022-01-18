package connect_test

import (
	"fmt"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

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
