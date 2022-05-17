package connect_test

import (
	"fmt"
	"strconv"
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
		"basic":                     testAccUser_basic,
		"disappears":                testAccUser_disappears,
		"update_hierarchy_group_id": testAccUser_updateHierarchyGroupId,
		"update_identity_info":      testAccUser_updateIdentityInfo,
		"update_phone_config":       testAccUser_updatePhoneConfig,
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

func testAccUser_updateHierarchyGroupId(t *testing.T) {
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
				Config: testAccUserHierarchyGroupIdConfig(rName, rName2, rName3, rName4, rName5, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "hierarchy_group_id", "aws_connect_user_hierarchy_group.parent", "hierarchy_group_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: testAccUserHierarchyGroupIdConfig(rName, rName2, rName3, rName4, rName5, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "hierarchy_group_id", "aws_connect_user_hierarchy_group.child", "hierarchy_group_id"),
				),
			},
		},
	})
}

func testAccUser_updateIdentityInfo(t *testing.T) {
	var v connect.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	email_original := "exampleoriginal@example.com"
	first_name_original := "example-first-name-original"
	last_name_original := "example-last-name-original"
	email_updated := "exampleupdated@example.com"
	first_name_updated := "example-first-name-updated"
	last_name_updated := "example-last-name-updated"

	resourceName := "aws_connect_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserIdentityInfoConfig(rName, rName2, rName3, rName4, rName5, email_original, first_name_original, last_name_original),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "identity_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_info.0.email", email_original),
					resource.TestCheckResourceAttr(resourceName, "identity_info.0.first_name", first_name_original),
					resource.TestCheckResourceAttr(resourceName, "identity_info.0.last_name", last_name_original),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: testAccUserIdentityInfoConfig(rName, rName2, rName3, rName4, rName5, email_updated, first_name_updated, last_name_updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "identity_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_info.0.email", email_updated),
					resource.TestCheckResourceAttr(resourceName, "identity_info.0.first_name", first_name_updated),
					resource.TestCheckResourceAttr(resourceName, "identity_info.0.last_name", last_name_updated),
				),
			},
		},
	})
}

func testAccUser_updatePhoneConfig(t *testing.T) {
	var v connect.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	after_contact_work_time_limit_original := 0
	auto_accept_original := false
	desk_phone_number_original := "+112345678912"
	after_contact_work_time_limit_updated := 1
	auto_accept_updated := true
	desk_phone_number_updated := "+112345678913"

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
					resource.TestCheckResourceAttr(resourceName, "phone_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.after_contact_work_time_limit", "0"),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.phone_type", connect.PhoneTypeSoftPhone),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: testAccUserPhoneConfigDeskPhoneConfig(rName, rName2, rName3, rName4, rName5, after_contact_work_time_limit_original, auto_accept_original, desk_phone_number_original),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "phone_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.after_contact_work_time_limit", strconv.Itoa(after_contact_work_time_limit_original)),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.auto_accept", strconv.FormatBool(auto_accept_original)),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.desk_phone_number", desk_phone_number_original),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.phone_type", connect.PhoneTypeDeskPhone),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: testAccUserPhoneConfigDeskPhoneConfig(rName, rName2, rName3, rName4, rName5, after_contact_work_time_limit_updated, auto_accept_updated, desk_phone_number_updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "phone_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.after_contact_work_time_limit", strconv.Itoa(after_contact_work_time_limit_updated)),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.auto_accept", strconv.FormatBool(auto_accept_updated)),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.desk_phone_number", desk_phone_number_updated),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.phone_type", connect.PhoneTypeDeskPhone),
				),
			},
		},
	})
}

func testAccUser_disappears(t *testing.T) {
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
					acctest.CheckResourceDisappears(acctest.Provider, tfconnect.ResourceUser(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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

func testAccUserHierarchyGroupIdConfig(rName, rName2, rName3, rName4, rName5, selectHierarchyGroupId string) string {
	return acctest.ConfigCompose(
		testAccUserBaseConfig(rName, rName2, rName3, rName4),
		fmt.Sprintf(`
locals {
  select_hierarchy_group_id = %[2]q
}

resource "aws_connect_user" "test" {
  instance_id        = aws_connect_instance.test.id
  name               = %[1]q
  password           = "Password123"
  routing_profile_id = data.aws_connect_routing_profile.test.routing_profile_id
  hierarchy_group_id = local.select_hierarchy_group_id == "first" ? aws_connect_user_hierarchy_group.parent.hierarchy_group_id : aws_connect_user_hierarchy_group.child.hierarchy_group_id

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
}
`, rName5, selectHierarchyGroupId))
}

func testAccUserIdentityInfoConfig(rName, rName2, rName3, rName4, rName5, email, first_name, last_name string) string {
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
    email      = %[2]q
    first_name = %[3]q
    last_name  = %[4]q
  }

  phone_config {
    after_contact_work_time_limit = 0
    phone_type                    = "SOFT_PHONE"
  }
}
`, rName5, email, first_name, last_name))
}

func testAccUserPhoneConfigDeskPhoneConfig(rName, rName2, rName3, rName4, rName5 string, after_contact_work_time_limit int, auto_accept bool, desk_phone_number string) string {
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
    after_contact_work_time_limit = %[2]d
    auto_accept                   = %[3]t
    desk_phone_number             = %[4]q
    phone_type                    = "DESK_PHONE"
  }

  tags = {
    "Key1" = "Value1",
  }
}
`, rName5, after_contact_work_time_limit, auto_accept, desk_phone_number))
}
