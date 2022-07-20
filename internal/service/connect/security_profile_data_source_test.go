package connect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccConnectSecurityProfileDataSource_securityProfileID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_security_profile.test"
	datasourceName := "data.aws_connect_security_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityProfileDataSourceConfig_id(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "permissions.#", resourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_profile_id", resourceName, "security_profile_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

func TestAccConnectSecurityProfileDataSource_name(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_security_profile.test"
	datasourceName := "data.aws_connect_security_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityProfileDataSourceConfig_name(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "permissions.#", resourceName, "permissions.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "security_profile_id", resourceName, "security_profile_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

func testAccSecurityProfileBaseDataSourceConfig(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}

resource "aws_connect_security_profile" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[2]q
  description = "test security profile data source"

  permissions = [
    "BasicAgentAccess",
    "OutboundCallAccess",
  ]

  tags = {
    "Name" = "Test Security Profile"
  }
}
`, rName, rName2)
}

func testAccSecurityProfileDataSourceConfig_id(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccSecurityProfileBaseDataSourceConfig(rName, rName2),
		`
data "aws_connect_security_profile" "test" {
  instance_id         = aws_connect_instance.test.id
  security_profile_id = aws_connect_security_profile.test.security_profile_id
}
`)
}

func testAccSecurityProfileDataSourceConfig_name(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccSecurityProfileBaseDataSourceConfig(rName, rName2),
		`
data "aws_connect_security_profile" "test" {
  instance_id = aws_connect_instance.test.id
  name        = aws_connect_security_profile.test.name
}
`)
}
