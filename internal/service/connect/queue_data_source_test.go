package connect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccConnectQueueDataSource_queueID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_queue.test"
	datasourceName := "data.aws_connect_queue.test"
	outboundCallerConfigName := "exampleOutboundCallerConfigName"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueueDataSourceConfig_id(rName, resourceName, outboundCallerConfigName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "hours_of_operation_id", resourceName, "hours_of_operation_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "max_contacts", resourceName, "max_contacts"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "outbound_caller_config.#", resourceName, "outbound_caller_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "outbound_caller_config.0.outbound_caller_id_name", resourceName, "outbound_caller_config.0.outbound_caller_id_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "queue_id", resourceName, "queue_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func TestAccConnectQueueDataSource_name(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_queue.test"
	datasourceName := "data.aws_connect_queue.test"
	outboundCallerConfigName := "exampleOutboundCallerConfigName"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueueDataSourceConfig_name(rName, rName2, outboundCallerConfigName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "hours_of_operation_id", resourceName, "hours_of_operation_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "max_contacts", resourceName, "max_contacts"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "outbound_caller_config.#", resourceName, "outbound_caller_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "outbound_caller_config.0.outbound_caller_id_name", resourceName, "outbound_caller_config.0.outbound_caller_id_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "queue_id", resourceName, "queue_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccQueueBaseDataSourceConfig(rName, rName2, outboundCallerConfigName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}

data "aws_connect_hours_of_operation" "test" {
  instance_id = aws_connect_instance.test.id
  name        = "Basic Hours"
}

resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[2]q
  description           = "Used to test queue data source"
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id

  outbound_caller_config {
    outbound_caller_id_name = %[3]q
  }

  tags = {
    "Name" = "Test Queue",
  }
}
	`, rName, rName2, outboundCallerConfigName)
}

func testAccQueueDataSourceConfig_id(rName, rName2, outboundCallerConfigName string) string {
	return acctest.ConfigCompose(
		testAccQueueBaseDataSourceConfig(rName, rName2, outboundCallerConfigName),
		`
data "aws_connect_queue" "test" {
  instance_id = aws_connect_instance.test.id
  queue_id    = aws_connect_queue.test.queue_id
}
`)
}

func testAccQueueDataSourceConfig_name(rName, rName2, outboundCallerConfigName string) string {
	return acctest.ConfigCompose(
		testAccQueueBaseDataSourceConfig(rName, rName2, outboundCallerConfigName),
		`
data "aws_connect_queue" "test" {
  instance_id = aws_connect_instance.test.id
  name        = aws_connect_queue.test.name
}
`)
}
