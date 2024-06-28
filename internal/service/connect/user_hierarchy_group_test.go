// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"context"
	"fmt"
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

func testAccUserHierarchyGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeUserHierarchyGroupOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_user_hierarchy_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserHierarchyGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserHierarchyGroupConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserHierarchyGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_group_id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_path.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "hierarchy_path.0.level_one.0.arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "hierarchy_path.0.level_one.0.id", resourceName, "hierarchy_group_id"),
					resource.TestCheckResourceAttrPair(resourceName, "hierarchy_path.0.level_one.0.name", resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "level_id", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test User Hierarchy Group"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Update name
				Config: testAccUserHierarchyGroupConfig_basic(rName, rName3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserHierarchyGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_group_id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_path.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "hierarchy_path.0.level_one.0.arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "hierarchy_path.0.level_one.0.id", resourceName, "hierarchy_group_id"),
					resource.TestCheckResourceAttrPair(resourceName, "hierarchy_path.0.level_one.0.name", resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "level_id", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test User Hierarchy Group"),
				),
			},
		},
	})
}

func testAccUserHierarchyGroup_parentGroupId(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeUserHierarchyGroupOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName3 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_user_hierarchy_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserHierarchyGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserHierarchyGroupConfig_parentID(rName, rName2, rName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserHierarchyGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_group_id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_path.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "hierarchy_path.0.level_one.0.arn", "aws_connect_user_hierarchy_group.parent", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "hierarchy_path.0.level_one.0.id", "aws_connect_user_hierarchy_group.parent", "hierarchy_group_id"),
					resource.TestCheckResourceAttrPair(resourceName, "hierarchy_path.0.level_one.0.name", "aws_connect_user_hierarchy_group.parent", names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "hierarchy_path.0.level_two.0.arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "hierarchy_path.0.level_two.0.id", resourceName, "hierarchy_group_id"),
					resource.TestCheckResourceAttrPair(resourceName, "hierarchy_path.0.level_two.0.name", resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "level_id", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName3),
					resource.TestCheckResourceAttrPair(resourceName, "parent_group_id", "aws_connect_user_hierarchy_group.parent", "hierarchy_group_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test User Hierarchy Group Child"),
				),
			},
		},
	})
}

func testAccUserHierarchyGroup_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeUserHierarchyGroupOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_user_hierarchy_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserHierarchyGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserHierarchyGroupConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserHierarchyGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test User Hierarchy Group"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserHierarchyGroupConfig_tags(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserHierarchyGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test User Hierarchy Group"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				Config: testAccUserHierarchyGroupConfig_tagsUpdated(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserHierarchyGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Test User Hierarchy Group"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
		},
	})
}

func testAccUserHierarchyGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeUserHierarchyGroupOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_user_hierarchy_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserHierarchyGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserHierarchyGroupConfig_basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserHierarchyGroupExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconnect.ResourceUserHierarchyGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUserHierarchyGroupExists(ctx context.Context, resourceName string, function *connect.DescribeUserHierarchyGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect User Hierarchy Group not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect User Hierarchy Group ID not set")
		}
		instanceID, userHierarchyGroupID, err := tfconnect.UserHierarchyGroupParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

		params := &connect.DescribeUserHierarchyGroupInput{
			HierarchyGroupId: aws.String(userHierarchyGroupID),
			InstanceId:       aws.String(instanceID),
		}

		getFunction, err := conn.DescribeUserHierarchyGroupWithContext(ctx, params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckUserHierarchyGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_user_hierarchy_group" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

			instanceID, userHierarchyGroupID, err := tfconnect.UserHierarchyGroupParseID(rs.Primary.ID)

			if err != nil {
				return err
			}

			params := &connect.DescribeUserHierarchyGroupInput{
				HierarchyGroupId: aws.String(userHierarchyGroupID),
				InstanceId:       aws.String(instanceID),
			}

			_, err = conn.DescribeUserHierarchyGroupWithContext(ctx, params)

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

func testAccUserHierarchyGroupConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
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

    level_three {
      name = "levelthree"
    }

    level_four {
      name = "levelfour"
    }

    level_five {
      name = "levelfive"
    }
  }
}
`, rName)
}

func testAccUserHierarchyGroupConfig_basic(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyGroupConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_user_hierarchy_group" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q

  tags = {
    "Name" = "Test User Hierarchy Group"
  }

  depends_on = [
    aws_connect_user_hierarchy_structure.test,
  ]
}
`, rName2))
}

func testAccUserHierarchyGroupConfig_parentID(rName, rName2, rName3 string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyGroupConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_user_hierarchy_group" "parent" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q

  tags = {
    "Name" = "Test User Hierarchy Group Parent"
  }

  depends_on = [
    aws_connect_user_hierarchy_structure.test,
  ]
}

resource "aws_connect_user_hierarchy_group" "test" {
  instance_id     = aws_connect_instance.test.id
  name            = %[2]q
  parent_group_id = aws_connect_user_hierarchy_group.parent.hierarchy_group_id

  tags = {
    "Name" = "Test User Hierarchy Group Child"
  }
}
`, rName2, rName3))
}

func testAccUserHierarchyGroupConfig_tags(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyGroupConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_user_hierarchy_group" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q

  tags = {
    "Name" = "Test User Hierarchy Group"
    "Key2" = "Value2a"
  }

  depends_on = [
    aws_connect_user_hierarchy_structure.test,
  ]
}
`, rName2))
}

func testAccUserHierarchyGroupConfig_tagsUpdated(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyGroupConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_user_hierarchy_group" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q

  tags = {
    "Name" = "Test User Hierarchy Group"
    "Key2" = "Value2b"
    "Key3" = "Value3"
  }

  depends_on = [
    aws_connect_user_hierarchy_structure.test,
  ]
}
`, rName2))
}
