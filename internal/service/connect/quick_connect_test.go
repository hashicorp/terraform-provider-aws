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

//Serialized acceptance tests due to Connect account limits (max 2 parallel tests)
func TestAccConnectQuickConnect_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":      testAccQuickConnect_phoneNumber,
		"disappears": testAccQuickConnect_disappears,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}


func testAccQuickConnectBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccQuickConnectPhoneNumberConfig(rName, rName2, label string) string {
	return acctest.ConfigCompose(
		testAccQuickConnectBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_connect_quick_connect" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  description = %[2]q

  quick_connect_config {
    quick_connect_type = "PHONE_NUMBER"

    phone_config {
	  phone_number   = "+12345678912"
    }
  }

  tags = {
    "Name" = "Test Quick Connect"
  }
}
`, rName2, label))
}
