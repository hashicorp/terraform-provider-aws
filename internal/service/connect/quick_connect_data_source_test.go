package connect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccConnectQuickConnectDataSource_id(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_quick_connect.test"
	datasourceName := "data.aws_connect_quick_connect.test"
	phoneNumber := "+12345678912"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQuickConnectDataSourceConfig_id(rName, resourceName, phoneNumber),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "quick_connect_config.#", resourceName, "quick_connect_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "quick_connect_config.0.quick_connect_type", resourceName, "quick_connect_config.0.quick_connect_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "quick_connect_config.0.phone_config.#", resourceName, "quick_connect_config.0.phone_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "quick_connect_config.0.phone_config.0.phone_number", resourceName, "quick_connect_config.0.phone_config.0.phone_number"),
					resource.TestCheckResourceAttrPair(datasourceName, "quick_connect_id", resourceName, "quick_connect_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func TestAccConnectQuickConnectDataSource_name(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_quick_connect.test"
	datasourceName := "data.aws_connect_quick_connect.test"
	phoneNumber := "+12345678912"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQuickConnectDataSourceConfig_Name(rName, rName2, phoneNumber),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "quick_connect_config.#", resourceName, "quick_connect_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "quick_connect_config.0.quick_connect_type", resourceName, "quick_connect_config.0.quick_connect_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "quick_connect_config.0.phone_config.#", resourceName, "quick_connect_config.0.phone_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "quick_connect_config.0.phone_config.0.phone_number", resourceName, "quick_connect_config.0.phone_config.0.phone_number"),
					resource.TestCheckResourceAttrPair(datasourceName, "quick_connect_id", resourceName, "quick_connect_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccQuickConnectBaseDataSourceConfig(rName, rName2, phoneNumber string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}

resource "aws_connect_quick_connect" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[2]q
  description = "Test Quick Connect Description"

  quick_connect_config {
    quick_connect_type = "PHONE_NUMBER"

    phone_config {
      phone_number = %[3]q
    }
  }

  tags = {
    "Name" = "Test Quick Connect"
  }
}
	`, rName, rName2, phoneNumber)
}

func testAccQuickConnectDataSourceConfig_id(rName, rName2, phoneNumber string) string {
	return acctest.ConfigCompose(
		testAccQuickConnectBaseDataSourceConfig(rName, rName2, phoneNumber),
		`
data "aws_connect_quick_connect" "test" {
  instance_id      = aws_connect_instance.test.id
  quick_connect_id = aws_connect_quick_connect.test.quick_connect_id
}
`)
}

func testAccQuickConnectDataSourceConfig_Name(rName, rName2, phoneNumber string) string {
	return acctest.ConfigCompose(
		testAccQuickConnectBaseDataSourceConfig(rName, rName2, phoneNumber),
		`
data "aws_connect_quick_connect" "test" {
  instance_id = aws_connect_instance.test.id
  name        = aws_connect_quick_connect.test.name
}
`)
}
