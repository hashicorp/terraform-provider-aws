package connect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccConnectRoutingProfileDataSource_routingProfileID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_routing_profile.test"
	datasourceName := "data.aws_connect_routing_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, connect.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingProfileDataSourceConfig_RoutingProfileID(rName, rName2, rName3, rName4),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "default_outbound_queue_id", resourceName, "default_outbound_queue_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "media_concurrencies.#", resourceName, "media_concurrencies.#"),
					resource.TestCheckTypeSetElemNestedAttrs(datasourceName, "media_concurrencies.*", map[string]string{
						"channel":     "VOICE",
						"concurrency": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(datasourceName, "media_concurrencies.*", map[string]string{
						"channel":     "CHAT",
						"concurrency": "2",
					}),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "queue_configs.#", resourceName, "queue_configs.#"),
					resource.TestCheckResourceAttrSet(datasourceName, "queue_configs.0.channel"),
					resource.TestCheckResourceAttrSet(datasourceName, "queue_configs.0.delay"),
					resource.TestCheckResourceAttrSet(datasourceName, "queue_configs.0.priority"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "queue_configs.*.queue_arn", "aws_connect_queue.default_outbound_queue", "arn"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "queue_configs.*.queue_id", "aws_connect_queue.default_outbound_queue", "queue_id"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "queue_configs.*.queue_name", "aws_connect_queue.default_outbound_queue", "name"),
					resource.TestCheckResourceAttrSet(datasourceName, "queue_configs.1.channel"),
					resource.TestCheckResourceAttrSet(datasourceName, "queue_configs.1.delay"),
					resource.TestCheckResourceAttrSet(datasourceName, "queue_configs.1.priority"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "queue_configs.*.queue_arn", "aws_connect_queue.test", "arn"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "queue_configs.*.queue_id", "aws_connect_queue.test", "queue_id"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "queue_configs.*.queue_name", "aws_connect_queue.test", "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "routing_profile_id", resourceName, "routing_profile_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

func TestAccConnectRoutingProfileDataSource_name(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_routing_profile.test"
	datasourceName := "data.aws_connect_routing_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, connect.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingProfileDataSourceConfig_Name(rName, rName2, rName3, rName4),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "default_outbound_queue_id", resourceName, "default_outbound_queue_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "media_concurrencies.#", resourceName, "media_concurrencies.#"),
					resource.TestCheckTypeSetElemNestedAttrs(datasourceName, "media_concurrencies.*", map[string]string{
						"channel":     "VOICE",
						"concurrency": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(datasourceName, "media_concurrencies.*", map[string]string{
						"channel":     "CHAT",
						"concurrency": "2",
					}),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "queue_configs.#", resourceName, "queue_configs.#"),
					resource.TestCheckResourceAttrSet(datasourceName, "queue_configs.0.channel"),
					resource.TestCheckResourceAttr(datasourceName, "queue_configs.0.delay", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "queue_configs.0.priority"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "queue_configs.*.queue_arn", "aws_connect_queue.default_outbound_queue", "arn"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "queue_configs.*.queue_id", "aws_connect_queue.default_outbound_queue", "queue_id"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "queue_configs.*.queue_name", "aws_connect_queue.default_outbound_queue", "name"),
					resource.TestCheckResourceAttrSet(datasourceName, "queue_configs.1.channel"),
					resource.TestCheckResourceAttr(datasourceName, "queue_configs.1.delay", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "queue_configs.1.priority"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "queue_configs.*.queue_arn", "aws_connect_queue.test", "arn"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "queue_configs.*.queue_id", "aws_connect_queue.test", "queue_id"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "queue_configs.*.queue_name", "aws_connect_queue.test", "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "routing_profile_id", resourceName, "routing_profile_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Name", resourceName, "tags.Name"),
				),
			},
		},
	})
}

func testAccRoutingProfileBaseDataSourceConfig(rName, rName2, rName3, rName4 string) string {
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

resource "aws_connect_queue" "default_outbound_queue" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[2]q
  description           = "Default Outbound Queue for Routing Profiles"
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id
}

resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[3]q
  description           = "Additional queue to routing profile queue config"
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id
}

resource "aws_connect_routing_profile" "test" {
  instance_id               = aws_connect_instance.test.id
  name                      = %[4]q
  default_outbound_queue_id = aws_connect_queue.default_outbound_queue.queue_id
  description               = "Test Routing Profile Data Source"

  media_concurrencies {
    channel     = "VOICE"
    concurrency = 1
  }

  media_concurrencies {
    channel     = "CHAT"
    concurrency = 2
  }

  queue_configs {
    channel  = "VOICE"
    delay    = 1
    priority = 2
    queue_id = aws_connect_queue.default_outbound_queue.queue_id
  }

  queue_configs {
    channel  = "CHAT"
    delay    = 1
    priority = 1
    queue_id = aws_connect_queue.test.queue_id
  }

  tags = {
    "Name" = "Test Routing Profile",
  }
}
`, rName, rName2, rName3, rName4)
}

func testAccRoutingProfileDataSourceConfig_RoutingProfileID(rName, rName2, rName3, rName4 string) string {
	return acctest.ConfigCompose(
		testAccRoutingProfileBaseDataSourceConfig(rName, rName2, rName3, rName4),
		`
data "aws_connect_routing_profile" "test" {
  instance_id        = aws_connect_instance.test.id
  routing_profile_id = aws_connect_routing_profile.test.routing_profile_id
}
`)
}

func testAccRoutingProfileDataSourceConfig_Name(rName, rName2, rName3, rName4 string) string {
	return acctest.ConfigCompose(
		testAccRoutingProfileBaseDataSourceConfig(rName, rName2, rName3, rName4),
		`
data "aws_connect_routing_profile" "test" {
  instance_id = aws_connect_instance.test.id
  name        = aws_connect_routing_profile.test.name
}
`)
}
