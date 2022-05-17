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
func TestAccConnectUser_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic": testAccUser_basic,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccUser_basic(t *testing.T) {
	var v connect.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resourceName := "aws_connect_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserBasicConfig(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "directory_user_id"),
					resource.TestCheckResourceAttr(resourceName, "identity_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_info.0.first_name", "example"),
					resource.TestCheckResourceAttr(resourceName, "identity_info.0.last_name", "example2"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName5),
					resource.TestCheckResourceAttr(resourceName, "password", "Password123"),
					resource.TestCheckResourceAttr(resourceName, "phone_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.after_contact_work_time_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.phone_type", connect.PhoneTypeSoftPhone),
					resource.TestCheckResourceAttrPair(resourceName, "routing_profile_id", "data.aws_connect_routing_profile.test", "routing_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "security_profile_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_profile_ids.*", "data.aws_connect_security_profile.agent", "security_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttrSet(resourceName, "user_id"),
				),
			},
		},
	})
}

func testAccCheckUserExists(resourceName string, function *connect.DescribeUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect User not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect User ID not set")
		}
		instanceID, userID, err := tfconnect.UserParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		params := &connect.DescribeUserInput{
			UserId:     aws.String(userID),
			InstanceId: aws.String(instanceID),
		}

		getFunction, err := conn.DescribeUser(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckUserDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_user" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		instanceID, userID, err := tfconnect.UserParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		params := &connect.DescribeUserInput{
			UserId:     aws.String(userID),
			InstanceId: aws.String(instanceID),
		}

		_, err = conn.DescribeUser(params)

		if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccUserBaseConfig(rName, rName2, rName3, rName4 string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}

data "aws_connect_queue" "test" {
  instance_id = aws_connect_instance.test.id
  name        = "BasicQueue"
}

resource "aws_connect_routing_profile" "test" {
  instance_id               = aws_connect_instance.test.id
  name                      = %[2]q
  default_outbound_queue_id = data.aws_connect_queue.test.queue_id
  description               = "test routing profile"

  media_concurrencies {
    channel     = "VOICE"
    concurrency = 1
  }
}

data "aws_connect_routing_profile" "test" {
  instance_id = aws_connect_instance.test.id
  name        = "Basic Routing Profile"
}

data "aws_connect_security_profile" "admin" {
  instance_id = aws_connect_instance.test.id
  name        = "Admin"
}

data "aws_connect_security_profile" "agent" {
  instance_id = aws_connect_instance.test.id
  name        = "Agent"
}

data "aws_connect_security_profile" "call_center_manager" {
  instance_id = aws_connect_instance.test.id
  name        = "CallCenterManager"
}

resource "aws_connect_user_hierarchy_structure" "test" {
  instance_id = aws_connect_instance.test.id

  hierarchy_structure {
    level_one {
      name = "levelone"
    }

    level_two {
      name = "leveltwo"
    }
  }
}

resource "aws_connect_user_hierarchy_group" "parent" {
  instance_id = aws_connect_instance.test.id
  name        = %[3]q

  tags = {
    "Name" = "Test User Hierarchy Group Parent"
  }

  depends_on = [
    aws_connect_user_hierarchy_structure.test,
  ]
}

resource "aws_connect_user_hierarchy_group" "child" {
  instance_id     = aws_connect_instance.test.id
  name            = %[4]q
  parent_group_id = aws_connect_user_hierarchy_group.parent.hierarchy_group_id

  tags = {
    "Name" = "Test User Hierarchy Group Child"
  }
}
`, rName, rName2, rName3, rName4)
}

func testAccUserBasicConfig(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccUserBaseConfig(rName, rName2, rName3, rName4),
		fmt.Sprintf(`
resource "aws_connect_user" "test" {
  instance_id        = aws_connect_instance.test.id
  name               = %[1]q
  password           = "Password123"
  routing_profile_id = data.aws_connect_routing_profile.test.routing_profile_id

  security_profile_ids = [
    data.aws_connect_security_profile.agent.security_profile_id
  ]

  identity_info {
    first_name = "example"
    last_name  = "example2"
  }

  phone_config {
    after_contact_work_time_limit = 0
    phone_type                    = "SOFT_PHONE"
  }

  tags = {
    "Key1" = "Value1",
  }
}
`, rName5))
}
