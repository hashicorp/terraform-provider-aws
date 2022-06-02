package connect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
)

// Serialized acceptance tests due to Connect account limits (max 2 parallel tests)
func TestAccConnectRoutingProfile_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":                         testAccRoutingProfile_basic,
		"disappears":                    testAccRoutingProfile_disappears,
		"update_concurrency":            testAccRoutingProfile_updateConcurrency,
		"update_default_outbound_queue": testAccRoutingProfile_updateDefaultOutboundQueue,
		"update_queues":                 testAccRoutingProfile_updateQueues,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccRoutingProfile_basic(t *testing.T) {
	var v connect.DescribeRoutingProfileOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_routing_profile.test"
	originalDescription := "Created"
	updatedDescription := "Updated"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRoutingProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingProfileConfig_basic(rName, rName2, rName3, originalDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoutingProfileExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_outbound_queue_id", "aws_connect_queue.default_outbound_queue", "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "description", originalDescription),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.channel", connect.ChannelVoice),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.concurrency", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
					resource.TestCheckResourceAttrSet(resourceName, "routing_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoutingProfileConfig_basic(rName, rName2, rName3, updatedDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingProfileExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_outbound_queue_id", "aws_connect_queue.default_outbound_queue", "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "description", updatedDescription),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.channel", connect.ChannelVoice),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.concurrency", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
					resource.TestCheckResourceAttrSet(resourceName, "routing_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccRoutingProfile_disappears(t *testing.T) {
	t.Skip("Routing Profiles do not support deletion today")
	var v connect.DescribeRoutingProfileOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_routing_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRoutingProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingProfileConfig_basic(rName, rName2, rName3, "Disappear"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoutingProfileExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfconnect.ResourceRoutingProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRoutingProfile_updateConcurrency(t *testing.T) {
	var v connect.DescribeRoutingProfileOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_routing_profile.test"
	description := "testMediaConcurrencies"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRoutingProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingProfileConfig_basic(rName, rName2, rName3, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoutingProfileExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_outbound_queue_id", "aws_connect_queue.default_outbound_queue", "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.channel", connect.ChannelVoice),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.concurrency", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
					resource.TestCheckResourceAttrSet(resourceName, "routing_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoutingProfileConfig_mediaConcurrencies(rName, rName2, rName3, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingProfileExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_outbound_queue_id", "aws_connect_queue.default_outbound_queue", "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
					resource.TestCheckResourceAttrSet(resourceName, "routing_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccRoutingProfile_updateDefaultOutboundQueue(t *testing.T) {
	var v connect.DescribeRoutingProfileOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_routing_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRoutingProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingProfileConfig_defaultOutboundQueue(rName, rName2, rName3, rName4, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoutingProfileExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_outbound_queue_id", "aws_connect_queue.default_outbound_queue", "queue_id"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.channel", connect.ChannelVoice),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.concurrency", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
					resource.TestCheckResourceAttrSet(resourceName, "routing_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoutingProfileConfig_defaultOutboundQueue(rName, rName2, rName3, rName4, "second"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingProfileExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_outbound_queue_id", "aws_connect_queue.default_outbound_queue_update", "queue_id"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.channel", connect.ChannelVoice),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.concurrency", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
					resource.TestCheckResourceAttrSet(resourceName, "routing_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccRoutingProfile_updateQueues(t *testing.T) {
	var v connect.DescribeRoutingProfileOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_routing_profile.test"
	description := "testQueueConfigs"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRoutingProfileDestroy,
		Steps: []resource.TestStep{
			{
				// Routing profile without queue_configs
				Config: testAccRoutingProfileConfig_basic(rName, rName2, rName3, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoutingProfileExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_outbound_queue_id", "aws_connect_queue.default_outbound_queue", "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.channel", connect.ChannelVoice),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.concurrency", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
					resource.TestCheckResourceAttr(resourceName, "queue_configs.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "routing_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Routing profile with one queue_configs
				Config: testAccRoutingProfileConfig_queue1(rName, rName2, rName3, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingProfileExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_outbound_queue_id", "aws_connect_queue.default_outbound_queue", "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.channel", connect.ChannelVoice),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.concurrency", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
					resource.TestCheckResourceAttr(resourceName, "queue_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "queue_configs.0.channel", connect.ChannelVoice),
					resource.TestCheckResourceAttr(resourceName, "queue_configs.0.delay", "2"),
					resource.TestCheckResourceAttr(resourceName, "queue_configs.0.priority", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "queue_configs.0.queue_arn", "aws_connect_queue.default_outbound_queue", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "queue_configs.0.queue_id", "aws_connect_queue.default_outbound_queue", "queue_id"),
					resource.TestCheckResourceAttrPair(resourceName, "queue_configs.0.queue_name", "aws_connect_queue.default_outbound_queue", "name"),
					resource.TestCheckResourceAttrSet(resourceName, "routing_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Routing profile with two queue_configs (one new config and one edited config)
				Config: testAccRoutingProfileConfig_queue2(rName, rName2, rName3, rName4, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingProfileExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_outbound_queue_id", "aws_connect_queue.default_outbound_queue", "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.channel", connect.ChannelVoice),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.concurrency", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
					resource.TestCheckResourceAttr(resourceName, "queue_configs.#", "2"),
					// The delay attribute of both elements of the set are set to 1
					resource.TestCheckResourceAttr(resourceName, "queue_configs.0.delay", "1"),
					resource.TestCheckResourceAttr(resourceName, "queue_configs.1.delay", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "routing_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Routing profile with one queue_configs (remove the created queue config)
				Config: testAccRoutingProfileConfig_queue1(rName, rName2, rName3, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoutingProfileExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_outbound_queue_id", "aws_connect_queue.default_outbound_queue", "queue_id"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.channel", connect.ChannelVoice),
					resource.TestCheckResourceAttr(resourceName, "media_concurrencies.0.concurrency", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName3),
					resource.TestCheckResourceAttr(resourceName, "queue_configs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "queue_configs.0.channel", connect.ChannelVoice),
					resource.TestCheckResourceAttr(resourceName, "queue_configs.0.delay", "2"),
					resource.TestCheckResourceAttr(resourceName, "queue_configs.0.priority", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "queue_configs.0.queue_arn", "aws_connect_queue.default_outbound_queue", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "queue_configs.0.queue_id", "aws_connect_queue.default_outbound_queue", "queue_id"),
					resource.TestCheckResourceAttrPair(resourceName, "queue_configs.0.queue_name", "aws_connect_queue.default_outbound_queue", "name"),
					resource.TestCheckResourceAttrSet(resourceName, "routing_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccCheckRoutingProfileExists(resourceName string, function *connect.DescribeRoutingProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Routing Profile not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Routing Profile ID not set")
		}
		instanceID, routingProfileID, err := tfconnect.RoutingProfileParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		params := &connect.DescribeRoutingProfileInput{
			InstanceId:       aws.String(instanceID),
			RoutingProfileId: aws.String(routingProfileID),
		}

		getFunction, err := conn.DescribeRoutingProfile(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckRoutingProfileDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_routing_profile" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		instanceID, routingProfileID, err := tfconnect.RoutingProfileParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		params := &connect.DescribeRoutingProfileInput{
			InstanceId:       aws.String(instanceID),
			RoutingProfileId: aws.String(routingProfileID),
		}

		_, err = conn.DescribeRoutingProfile(params)

		if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccRoutingProfileBaseConfig(rName, rName2 string) string {
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
`, rName, rName2)
}

func testAccRoutingProfileConfig_basic(rName, rName2, rName3, label string) string {
	return acctest.ConfigCompose(
		testAccRoutingProfileBaseConfig(rName, rName2),
		fmt.Sprintf(`
resource "aws_connect_routing_profile" "test" {
  instance_id               = aws_connect_instance.test.id
  name                      = %[1]q
  default_outbound_queue_id = aws_connect_queue.default_outbound_queue.queue_id
  description               = %[2]q

  media_concurrencies {
    channel     = "VOICE"
    concurrency = 1
  }

  tags = {
    "Name" = "Test Routing Profile",
  }
}
`, rName3, label))
}

func testAccRoutingProfileConfig_mediaConcurrencies(rName, rName2, rName3, label string) string {
	return acctest.ConfigCompose(
		testAccRoutingProfileBaseConfig(rName, rName2),
		fmt.Sprintf(`
resource "aws_connect_routing_profile" "test" {
  instance_id               = aws_connect_instance.test.id
  name                      = %[1]q
  default_outbound_queue_id = aws_connect_queue.default_outbound_queue.queue_id
  description               = %[2]q

  media_concurrencies {
    channel     = "VOICE"
    concurrency = 1
  }

  media_concurrencies {
    channel     = "CHAT"
    concurrency = 2
  }

  tags = {
    "Name" = "Test Routing Profile",
  }
}
`, rName3, label))
}

func testAccRoutingProfileConfig_defaultOutboundQueue(rName, rName2, rName3, rName4, selectDefaultOutboundQueue string) string {
	return acctest.ConfigCompose(
		testAccRoutingProfileBaseConfig(rName, rName2),
		fmt.Sprintf(`
locals {
  select_default_outbound_queue_id = %[3]q
}

resource "aws_connect_queue" "default_outbound_queue_update" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[2]q
  description           = "Default Outbound Queue for Routing Profiles"
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id
}

resource "aws_connect_routing_profile" "test" {
  instance_id               = aws_connect_instance.test.id
  name                      = %[1]q
  default_outbound_queue_id = local.select_default_outbound_queue_id == "first" ? aws_connect_queue.default_outbound_queue.queue_id : aws_connect_queue.default_outbound_queue_update.queue_id
  description               = "Test updating the default outbound queue id"

  media_concurrencies {
    channel     = "VOICE"
    concurrency = 1
  }

  tags = {
    "Name" = "Test Routing Profile",
  }
}
`, rName3, rName4, selectDefaultOutboundQueue))
}

func testAccRoutingProfileConfig_queue1(rName, rName2, rName3, label string) string {
	return acctest.ConfigCompose(
		testAccRoutingProfileBaseConfig(rName, rName2),
		fmt.Sprintf(`
resource "aws_connect_routing_profile" "test" {
  instance_id               = aws_connect_instance.test.id
  name                      = %[1]q
  default_outbound_queue_id = aws_connect_queue.default_outbound_queue.queue_id
  description               = %[2]q

  media_concurrencies {
    channel     = "VOICE"
    concurrency = 1
  }

  queue_configs {
    channel  = "VOICE"
    delay    = 2
    priority = 1
    queue_id = aws_connect_queue.default_outbound_queue.queue_id
  }

  tags = {
    "Name" = "Test Routing Profile",
  }
}
`, rName3, label))
}

func testAccRoutingProfileConfig_queue2(rName, rName2, rName3, rName4, label string) string {
	return acctest.ConfigCompose(
		testAccRoutingProfileBaseConfig(rName, rName2),
		fmt.Sprintf(`
resource "aws_connect_queue" "test" {
  instance_id           = aws_connect_instance.test.id
  name                  = %[2]q
  description           = "Additional queue to routing profile queue config"
  hours_of_operation_id = data.aws_connect_hours_of_operation.test.hours_of_operation_id
}

resource "aws_connect_routing_profile" "test" {
  instance_id               = aws_connect_instance.test.id
  name                      = %[1]q
  default_outbound_queue_id = aws_connect_queue.default_outbound_queue.queue_id
  description               = %[3]q

  media_concurrencies {
    channel     = "VOICE"
    concurrency = 1
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
`, rName3, rName4, label))
}
