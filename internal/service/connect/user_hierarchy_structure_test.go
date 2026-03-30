// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccUserHierarchyStructure_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.HierarchyStructure
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	levelOneName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	levelTwoName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	levelThreeName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	levelFourName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	levelFiveName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	resourceName := "aws_connect_user_hierarchy_structure.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserHierarchyStructureDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserHierarchyStructureConfig_basic(rName, levelOneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserHierarchyStructureExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.#", "1"),
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
					testAccCheckUserHierarchyStructureExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.0.name", levelOneName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.#", "1"),
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
					testAccCheckUserHierarchyStructureExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.0.name", levelOneName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_two.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_two.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.0.name", levelTwoName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_three.#", "1"),
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
					testAccCheckUserHierarchyStructureExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.0.name", levelOneName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_two.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_two.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.0.name", levelTwoName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_three.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_three.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_three.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_three.0.name", levelThreeName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_four.#", "1"),
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
					testAccCheckUserHierarchyStructureExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.0.name", levelOneName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_two.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_two.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.0.name", levelTwoName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_three.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_three.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_three.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_three.0.name", levelThreeName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_four.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_four.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_four.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_four.0.name", levelFourName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_five.#", "1"),
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
					testAccCheckUserHierarchyStructureExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.#", "1"),
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
	var v awstypes.HierarchyStructure
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	resourceName := "aws_connect_user_hierarchy_structure.test"
	levelOneName := acctest.RandomWithPrefix(t, "resource-test-terraform")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserHierarchyStructureDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserHierarchyStructureConfig_basic(rName, levelOneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserHierarchyStructureExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfconnect.ResourceUserHierarchyStructure(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUserHierarchyStructureExists(ctx context.Context, t *testing.T, n string, v *awstypes.HierarchyStructure) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ConnectClient(ctx)

		output, err := tfconnect.FindUserHierarchyStructureByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckUserHierarchyStructureDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_user_hierarchy_structure" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).ConnectClient(ctx)

			_, err := tfconnect.FindUserHierarchyStructureByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Connect User Hierarchy Structure %s still exists", rs.Primary.ID)
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
