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

// Serialized acceptance tests due to Connect account limits (max 2 parallel tests)
func TestAccConnectRoutingProfile_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":      testAccRoutingProfile_basic,
		"disappears": testAccRoutingProfile_disappears,
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, connect.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoutingProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingProfileBasicConfig(rName, rName2, rName3, originalDescription),
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
				Config: testAccRoutingProfileBasicConfig(rName, rName2, rName3, updatedDescription),
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, connect.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoutingProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutingProfileBasicConfig(rName, rName2, rName3, "Disappear"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoutingProfileExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfconnect.ResourceRoutingProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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

		_, experr := conn.DescribeRoutingProfile(params)
		// Verify the error is what we want
		if experr != nil {
			if awsErr, ok := experr.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				continue
			}
			return experr
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

func testAccRoutingProfileBasicConfig(rName, rName2, rName3, label string) string {
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
