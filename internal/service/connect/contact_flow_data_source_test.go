package connect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccConnectContactFlowDataSource_contactFlowID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_contact_flow.test"
	datasourceName := "data.aws_connect_contact_flow.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccContactFlowDataSourceConfig_id(rName, resourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "contact_flow_id", resourceName, "contact_flow_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "content", resourceName, "content"),
					resource.TestCheckResourceAttrPair(datasourceName, "type", resourceName, "type"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func TestAccConnectContactFlowDataSource_name(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_contact_flow.test"
	datasourceName := "data.aws_connect_contact_flow.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccContactFlowDataSourceConfig_name(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "contact_flow_id", resourceName, "contact_flow_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "content", resourceName, "content"),
					resource.TestCheckResourceAttrPair(datasourceName, "type", resourceName, "type"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccContactFlowBaseDataSourceConfig(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}

resource "aws_connect_contact_flow" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[2]q
  description = "Test Contact Flow Description"
  type        = "CONTACT_FLOW"
  content     = file("./test-fixtures/connect_contact_flow.json")
  tags = {
    "Name"        = "Test Contact Flow",
    "Application" = "Terraform",
    "Method"      = "Create"
  }
}
	`, rName, rName2)
}

func testAccContactFlowDataSourceConfig_id(rName, rName2 string) string {
	return fmt.Sprintf(testAccContactFlowBaseDataSourceConfig(rName, rName2) + `
data "aws_connect_contact_flow" "test" {
  instance_id     = aws_connect_instance.test.id
  contact_flow_id = aws_connect_contact_flow.test.contact_flow_id
}
`)
}

func testAccContactFlowDataSourceConfig_name(rName, rName2 string) string {
	return fmt.Sprintf(testAccContactFlowBaseDataSourceConfig(rName, rName2) + `
data "aws_connect_contact_flow" "test" {
  instance_id = aws_connect_instance.test.id
  name        = aws_connect_contact_flow.test.name
}
`)
}
