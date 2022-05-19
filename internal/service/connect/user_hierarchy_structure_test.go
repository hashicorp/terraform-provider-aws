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

//Serialized acceptance tests due to Connect account limits (max 2 parallel tests)
func TestAccConnectUserHierarchyStructure_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":      testAccUserHierarchyStructure_basic,
		"disappears": testAccUserHierarchyStructure_disappears,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccUserHierarchyStructure_basic(t *testing.T) {
	var v connect.DescribeUserHierarchyStructureOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	levelOneName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	levelTwoName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	levelThreeName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	levelFourName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	levelFiveName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_user_hierarchy_structure.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserHierarchyStructureDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserHierarchyStructureBasicConfig(rName, levelOneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserHierarchyStructureExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.0.name", levelOneName),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					// resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserHierarchyStructureBasicTwoLevelsConfig(rName, levelOneName, levelTwoName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserHierarchyStructureExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.0.name", levelOneName),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_two.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_two.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_two.0.name", levelTwoName),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserHierarchyStructureBasicThreeLevelsConfig(rName, levelOneName, levelTwoName, levelThreeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserHierarchyStructureExists(resourceName, &v),
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
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserHierarchyStructureBasicFourLevelsConfig(rName, levelOneName, levelTwoName, levelThreeName, levelFourName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserHierarchyStructureExists(resourceName, &v),
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
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserHierarchyStructureBasicFiveLevelsConfig(rName, levelOneName, levelTwoName, levelThreeName, levelFourName, levelFiveName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserHierarchyStructureExists(resourceName, &v),
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
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// test removing 4 levels
				Config: testAccUserHierarchyStructureBasicConfig(rName, levelOneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserHierarchyStructureExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hierarchy_structure.0.level_one.0.id"),
					resource.TestCheckResourceAttr(resourceName, "hierarchy_structure.0.level_one.0.name", levelOneName),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_connect_instance.test", "id"),
				),
			},
		},
	})
}

func testAccUserHierarchyStructure_disappears(t *testing.T) {
	var v connect.DescribeUserHierarchyStructureOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_user_hierarchy_structure.test"
	levelOneName := sdkacctest.RandomWithPrefix("resource-test-terraform")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, connect.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserHierarchyStructureDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserHierarchyStructureBasicConfig(rName, levelOneName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserHierarchyStructureExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfconnect.ResourceUserHierarchyStructure(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUserHierarchyStructureExists(resourceName string, function *connect.DescribeUserHierarchyStructureOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect User Hierarchy Structure not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect User Hierarchy Structure ID not set")
		}
		instanceID := rs.Primary.ID

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		params := &connect.DescribeUserHierarchyStructureInput{
			InstanceId: aws.String(instanceID),
		}

		getFunction, err := conn.DescribeUserHierarchyStructure(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckUserHierarchyStructureDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_user_hierarchy_structure" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		instanceID := rs.Primary.ID

		params := &connect.DescribeUserHierarchyStructureInput{
			InstanceId: aws.String(instanceID),
		}

		resp, err := conn.DescribeUserHierarchyStructure(params)

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

func testAccUserHierarchyStructureBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccUserHierarchyStructureBasicConfig(rName, levelOneName string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyStructureBaseConfig(rName),
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

func testAccUserHierarchyStructureBasicTwoLevelsConfig(rName, levelOneName, levelTwoName string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyStructureBaseConfig(rName),
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

func testAccUserHierarchyStructureBasicThreeLevelsConfig(rName, levelOneName, levelTwoName, levelThreeName string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyStructureBaseConfig(rName),
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

func testAccUserHierarchyStructureBasicFourLevelsConfig(rName, levelOneName, levelTwoName, levelThreeName, levelFourName string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyStructureBaseConfig(rName),
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

func testAccUserHierarchyStructureBasicFiveLevelsConfig(rName, levelOneName, levelTwoName, levelThreeName, levelFourName, levelFiveName string) string {
	return acctest.ConfigCompose(
		testAccUserHierarchyStructureBaseConfig(rName),
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
