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

func testAccUserHierarchyStructure_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeUserHierarchyStructureOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	levelOneName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	levelTwoName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	levelThreeName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	levelFourName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	levelFiveName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_user_hierarchy_structure.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserHierarchyStructureDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserHierarchyStructureConfig_basic(rName, levelOneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserHierarchyStructureExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.0.name", levelOneName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					// resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserHierarchyStructureConfig_twoLevels(rName, levelOneName, levelTwoName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserHierarchyStructureExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.0.name", levelOneName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_two.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_two.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.0.name", levelTwoName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserHierarchyStructureConfig_threeLevels(rName, levelOneName, levelTwoName, levelThreeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserHierarchyStructureExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.0.name", levelOneName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_two.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_two.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.0.name", levelTwoName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_three.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_three.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_three.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_three.0.name", levelThreeName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserHierarchyStructureConfig_fourLevels(rName, levelOneName, levelTwoName, levelThreeName, levelFourName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserHierarchyStructureExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.0.name", levelOneName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_two.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_two.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.0.name", levelTwoName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_three.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_three.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_three.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_three.0.name", levelThreeName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_four.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_four.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_four.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_four.0.name", levelFourName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserHierarchyStructureConfig_fiveLevels(rName, levelOneName, levelTwoName, levelThreeName, levelFourName, levelFiveName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserHierarchyStructureExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.0.name", levelOneName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_two.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_two.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.0.name", levelTwoName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_three.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_three.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_three.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_three.0.name", levelThreeName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_four.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_four.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_four.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_four.0.name", levelFourName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_five.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_five.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_five.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_five.0.name", levelFiveName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// test removing 4 levels
				Config: testAccUserHierarchyStructureConfig_basic(rName, levelOneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserHierarchyStructureExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.0.name", levelOneName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, "aws_connect_instance.test", names.AttrID),
				),
			},
		},
	})
}

func testAccUserHierarchyStructure_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v connect.DescribeUserHierarchyStructureOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_user_hierarchy_structure.test"
	levelOneName := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserHierarchyStructureDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserHierarchyStructureConfig_basic(rName, levelOneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserHierarchyStructureExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconnect.ResourceUserHierarchyStructure(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUserHierarchyStructureExists(ctx context.Context, resourceName string, function *connect.DescribeUserHierarchyStructureOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect User Hierarchy Structure not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect User Hierarchy Structure ID not set")
		}
		instanceID := rs.Primary.ID

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

		params := &connect.DescribeUserHierarchyStructureInput{
			InstanceId: aws.String(instanceID),
		}

		getFunction, err := conn.DescribeUserHierarchyStructureWithContext(ctx, params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckUserHierarchyStructureDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_user_hierarchy_structure" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

			instanceID := rs.Primary.ID

			params := &connect.DescribeUserHierarchyStructureInput{
				InstanceId: aws.String(instanceID),
			}

			resp, err := conn.DescribeUserHierarchyStructureWithContext(ctx, params)

			if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return err
			}

			// API returns an empty list for HierarchyStructure if there are none
			if resp.HierarchyStructure == nil {
				continue
			}
		}

		return nil
	}
}

func testAccUserHierarchyStructureConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccUserHierarchyStructureConfig_basic(rName, levelOneName string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyStructureConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_user_hierarchy_structure" "test" {
  instance_id = aws_connect_instance.test.id

  hierarchy_structure {
    level_one {
      name = %[1]q
    }
  }
}
`, levelOneName))
}

func testAccUserHierarchyStructureConfig_twoLevels(rName, levelOneName, levelTwoName string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyStructureConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_user_hierarchy_structure" "test" {
  instance_id = aws_connect_instance.test.id

  hierarchy_structure {
    level_one {
      name = %[1]q
    }

    level_two {
      name = %[2]q
    }
  }
}
`, levelOneName, levelTwoName))
}

func testAccUserHierarchyStructureConfig_threeLevels(rName, levelOneName, levelTwoName, levelThreeName string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyStructureConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_user_hierarchy_structure" "test" {
  instance_id = aws_connect_instance.test.id

  hierarchy_structure {
    level_one {
      name = %[1]q
    }

    level_two {
      name = %[2]q
    }

    level_three {
      name = %[3]q
    }
  }
}
`, levelOneName, levelTwoName, levelThreeName))
}

func testAccUserHierarchyStructureConfig_fourLevels(rName, levelOneName, levelTwoName, levelThreeName, levelFourName string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyStructureConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_user_hierarchy_structure" "test" {
  instance_id = aws_connect_instance.test.id

  hierarchy_structure {
    level_one {
      name = %[1]q
    }

    level_two {
      name = %[2]q
    }

    level_three {
      name = %[3]q
    }

    level_four {
      name = %[4]q
    }
  }
}
`, levelOneName, levelTwoName, levelThreeName, levelFourName))
}

func testAccUserHierarchyStructureConfig_fiveLevels(rName, levelOneName, levelTwoName, levelThreeName, levelFourName, levelFiveName string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyStructureConfig_base(rName),
		fmt.Sprintf(`
resource "aws_connect_user_hierarchy_structure" "test" {
  instance_id = aws_connect_instance.test.id

  hierarchy_structure {
    level_one {
      name = %[1]q
    }

    level_two {
      name = %[2]q
    }

    level_three {
      name = %[3]q
    }

    level_four {
      name = %[4]q
    }

    level_five {
      name = %[5]q
    }
  }
}
`, levelOneName, levelTwoName, levelThreeName, levelFourName, levelFiveName))
}
