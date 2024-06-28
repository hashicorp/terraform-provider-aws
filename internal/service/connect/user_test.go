// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccUser_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resourceName := "aws_connect_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "directory_user_id"),
					resource.TestCheckResourceAttr(resourceName, "identity_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "identity_info.0.first_name", "example"),
					resource.TestCheckResourceAttr(resourceName, "identity_info.0.last_name", "example2"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName5),
					resource.TestCheckResourceAttr(resourceName, names.AttrPassword, "Password123"),
					resource.TestCheckResourceAttr(resourceName, "phone_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.after_contact_work_time_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.phone_type", connect.PhoneTypeSoftPhone),
					resource.TestCheckResourceAttrPair(resourceName, "routing_profile_id", "data.aws_connect_routing_profile.test", "routing_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "security_profile_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_profile_ids.*", "data.aws_connect_security_profile.agent", "security_profile_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttrSet(resourceName, "user_id"),
				),
			},
		},
	})
}

func testAccUser_updateHierarchyGroupId(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resourceName := "aws_connect_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_hierarchyGroupID(rName, rName2, rName3, rName4, rName5, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "hierarchy_group_id", "aws_connect_user_hierarchy_group.parent", "hierarchy_group_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccUserConfig_hierarchyGroupID(rName, rName2, rName3, rName4, rName5, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "hierarchy_group_id", "aws_connect_user_hierarchy_group.child", "hierarchy_group_id"),
				),
			},
		},
	})
}

func testAccUser_updateIdentityInfo(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	domain := acctest.RandomDomainName()
	emailOriginal := acctest.RandomEmailAddress(domain)
	firstNameOriginal := "example-first-name-original"
	lastNameOriginal := "example-last-name-original"
	emailUpdated := acctest.RandomEmailAddress(domain)
	firstNameUpdated := "example-first-name-updated"
	lastNameUpdated := "example-last-name-updated"

	resourceName := "aws_connect_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_identityInfo(rName, rName2, rName3, rName4, rName5, emailOriginal, firstNameOriginal, lastNameOriginal),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "identity_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "identity_info.0.email", emailOriginal),
					resource.TestCheckResourceAttr(resourceName, "identity_info.0.first_name", firstNameOriginal),
					resource.TestCheckResourceAttr(resourceName, "identity_info.0.last_name", lastNameOriginal),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccUserConfig_identityInfo(rName, rName2, rName3, rName4, rName5, emailUpdated, firstNameUpdated, lastNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "identity_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "identity_info.0.email", emailUpdated),
					resource.TestCheckResourceAttr(resourceName, "identity_info.0.first_name", firstNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "identity_info.0.last_name", lastNameUpdated),
				),
			},
		},
	})
}

func testAccUser_updatePhoneConfig(t *testing.T) {
	ctx := acctest.Context(t)
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "phone_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.after_contact_work_time_limit", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.phone_type", connect.PhoneTypeSoftPhone),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccUserConfig_phoneDeskPhone(rName, rName2, rName3, rName4, rName5, after_contact_work_time_limit_original, auto_accept_original, desk_phone_number_original),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "phone_config.#", acctest.Ct1),
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
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccUserConfig_phoneDeskPhone(rName, rName2, rName3, rName4, rName5, after_contact_work_time_limit_updated, auto_accept_updated, desk_phone_number_updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "phone_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.after_contact_work_time_limit", strconv.Itoa(after_contact_work_time_limit_updated)),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.auto_accept", strconv.FormatBool(auto_accept_updated)),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.desk_phone_number", desk_phone_number_updated),
					resource.TestCheckResourceAttr(resourceName, "phone_config.0.phone_type", connect.PhoneTypeDeskPhone),
				),
			},
		},
	})
}

func testAccUser_updateSecurityProfileIds(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resourceName := "aws_connect_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_securityProfileIDs(rName, rName2, rName3, rName4, rName5, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_profile_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_profile_ids.*", "data.aws_connect_security_profile.agent", "security_profile_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccUserConfig_securityProfileIDs(rName, rName2, rName3, rName4, rName5, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_profile_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_profile_ids.*", "data.aws_connect_security_profile.agent", "security_profile_id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_profile_ids.*", "data.aws_connect_security_profile.call_center_manager", "security_profile_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccUserConfig_securityProfileIDs(rName, rName2, rName3, rName4, rName5, "third"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "security_profile_ids.#", acctest.Ct3),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_profile_ids.*", "data.aws_connect_security_profile.agent", "security_profile_id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_profile_ids.*", "data.aws_connect_security_profile.call_center_manager", "security_profile_id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_profile_ids.*", "data.aws_connect_security_profile.admin", "security_profile_id"),
				),
			},
		},
	})
}

func testAccUser_updateRoutingProfileId(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resourceName := "aws_connect_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_routingProfileID(rName, rName2, rName3, rName4, rName5, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "routing_profile_id", "aws_connect_routing_profile.test", "routing_profile_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccUserConfig_routingProfileID(rName, rName2, rName3, rName4, rName5, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "routing_profile_id", "data.aws_connect_routing_profile.test", "routing_profile_id"),
				),
			},
		},
	})
}

func testAccUser_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resourceName := "aws_connect_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPassword},
			},
			{
				Config: testAccUserConfig_tags(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				Config: testAccUserConfig_tagsUpdated(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
		},
	})
}

func testAccUser_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeUserOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName4 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName5 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_user.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_basic(rName, rName2, rName3, rName4, rName5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconnect.ResourceUser(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUserExists(ctx context.Context, resourceName string, function *connect.DescribeUserOutput) resource.TestCheckFunc {
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

		params := &connect.DescribeUserInput{
			UserId:     aws.String(userID),
			InstanceId: aws.String(instanceID),
		}

		getFunction, err := conn.DescribeUserWithContext(ctx, params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckUserDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_user" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

			instanceID, userID, err := tfconnect.UserParseID(rs.Primary.ID)

			if err != nil {
				return err
			}

			params := &connect.DescribeUserInput{
				UserId:     aws.String(userID),
				InstanceId: aws.String(instanceID),
			}

			_, err = conn.DescribeUserWithContext(ctx, params)

			if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccUserConfig_base(rName, rName2, rName3, rName4 string) string {
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

func testAccUserConfig_basic(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccUserConfig_base(rName, rName2, rName3, rName4),
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

func testAccUserConfig_hierarchyGroupID(rName, rName2, rName3, rName4, rName5, selectHierarchyGroupId string) string {
	return acctest.ConfigCompose(
		testAccUserConfig_base(rName, rName2, rName3, rName4),
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

func testAccUserConfig_identityInfo(rName, rName2, rName3, rName4, rName5, email, first_name, last_name string) string {
	return acctest.ConfigCompose(
		testAccUserConfig_base(rName, rName2, rName3, rName4),
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

func testAccUserConfig_phoneDeskPhone(rName, rName2, rName3, rName4, rName5 string, after_contact_work_time_limit int, auto_accept bool, desk_phone_number string) string {
	return acctest.ConfigCompose(
		testAccUserConfig_base(rName, rName2, rName3, rName4),
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

func testAccUserConfig_routingProfileID(rName, rName2, rName3, rName4, rName5, selectRoutingProfileId string) string {
	return acctest.ConfigCompose(
		testAccUserConfig_base(rName, rName2, rName3, rName4),
		fmt.Sprintf(`
locals {
  selectRoutingProfileId = %[2]q
}

resource "aws_connect_user" "test" {
  instance_id        = aws_connect_instance.test.id
  name               = %[1]q
  password           = "Password123"
  routing_profile_id = local.selectRoutingProfileId == "first" ? aws_connect_routing_profile.test.routing_profile_id : data.aws_connect_routing_profile.test.routing_profile_id

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
`, rName5, selectRoutingProfileId))
}

func testAccUserConfig_securityProfileIDs(rName, rName2, rName3, rName4, rName5, selectSecurityProfileIds string) string {
	return acctest.ConfigCompose(
		testAccUserConfig_base(rName, rName2, rName3, rName4),
		fmt.Sprintf(`
locals {
  security_profile_ids_map = {
    "first" = [data.aws_connect_security_profile.agent.security_profile_id],
    "second" = [
      data.aws_connect_security_profile.agent.security_profile_id,
      data.aws_connect_security_profile.call_center_manager.security_profile_id,
    ]
    "third" = [
      data.aws_connect_security_profile.agent.security_profile_id,
      data.aws_connect_security_profile.call_center_manager.security_profile_id,
      data.aws_connect_security_profile.admin.security_profile_id,
    ]
  }
}

resource "aws_connect_user" "test" {
  instance_id        = aws_connect_instance.test.id
  name               = %[1]q
  password           = "Password123"
  routing_profile_id = data.aws_connect_routing_profile.test.routing_profile_id

  security_profile_ids = local.security_profile_ids_map[%[2]q]

  identity_info {
    first_name = "example"
    last_name  = "example2"
  }

  phone_config {
    after_contact_work_time_limit = 0
    phone_type                    = "SOFT_PHONE"
  }
}
`, rName5, selectSecurityProfileIds))
}

func testAccUserConfig_tags(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccUserConfig_base(rName, rName2, rName3, rName4),
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
    "Key2" = "Value2a",
  }
}
`, rName5))
}

func testAccUserConfig_tagsUpdated(rName, rName2, rName3, rName4, rName5 string) string {
	return acctest.ConfigCompose(
		testAccUserConfig_base(rName, rName2, rName3, rName4),
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
    "Key2" = "Value2b"
    "Key3" = "Value3"
  }
}
`, rName5))
}
