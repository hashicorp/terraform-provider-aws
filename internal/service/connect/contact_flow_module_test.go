package connect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/connect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
)


func testAccContactFlowModuleBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccContactFlowModuleConfig_basic(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccContactFlowModuleBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_contact_flow_module" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q

  content = <<JSON
    {
		"Version": "2019-10-30",
		"StartAction": "12345678-1234-1234-1234-123456789012",
		"Actions": [
			{
				"Identifier": "12345678-1234-1234-1234-123456789012",
				"Parameters": {
					"Text": "Hello contact flow module"
				},
				"Transitions": {
					"NextAction": "abcdef-abcd-abcd-abcd-abcdefghijkl",
					"Errors": [],
					"Conditions": []
				},
				"Type": "MessageParticipant"
			},
			{
				"Identifier": "abcdef-abcd-abcd-abcd-abcdefghijkl",
				"Type": "DisconnectParticipant",
				"Parameters": {},
				"Transitions": {}
			}
		],
		"Settings": {
			"InputParameters": [],
			"OutputParameters": [],
			"Transitions": [
				{
					"DisplayName": "Success",
					"ReferenceName": "Success",
					"Description": ""
				},
				{
					"DisplayName": "Error",
					"ReferenceName": "Error",
					"Description": ""
				}
			]
		}
	}
    JSON

  tags = {
    "Name"   = "Test Contact Flow Module",
    "Method" = %[2]q
  }
}
`, rName2, label))
}

func testAccContactFlowModuleConfig_filename(rName, rName2 string, label string, filepath string) string {
	return acctest.ConfigCompose(
		testAccContactFlowModuleBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_contact_flow_module" "test" {
  instance_id  = aws_connect_instance.test.id
  name         = %[1]q
  description  = %[2]q
  filename     = %[3]q
  content_hash = filebase64sha256(%[3]q)

  tags = {
    "Name"   = "Test Contact Flow Module",
    "Method" = %[2]q
  }
}
`, rName2, label, filepath))
}
