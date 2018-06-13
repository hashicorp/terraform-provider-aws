package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSNeptuneParameterGroup_basic(t *testing.T) {
	var v neptune.DBParameterGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_neptune_parameter_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneParameterGroupConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneParameterGroupExists(resourceName, &v),
					testAccCheckAWSNeptuneParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "family", "neptune1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSNeptuneParameterGroup_Description(t *testing.T) {
	var v neptune.DBParameterGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_neptune_parameter_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneParameterGroupConfig_Description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneParameterGroupExists(resourceName, &v),
					testAccCheckAWSNeptuneParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSNeptuneParameterGroup_Parameter(t *testing.T) {
	var v neptune.DBParameterGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_neptune_parameter_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneParameterGroupConfig_Parameter(rName, "neptune_query_timeout", "25", "pending-reboot"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneParameterGroupExists(resourceName, &v),
					testAccCheckAWSNeptuneParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter.2423897584.apply_method", "pending-reboot"),
					resource.TestCheckResourceAttr(resourceName, "parameter.2423897584.name", "neptune_query_timeout"),
					resource.TestCheckResourceAttr(resourceName, "parameter.2423897584.value", "25"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// This test should be updated with a dynamic parameter when available
			{
				Config:      testAccAWSNeptuneParameterGroupConfig_Parameter(rName, "neptune_query_timeout", "25", "immediate"),
				ExpectError: regexp.MustCompile(`cannot use immediate apply method for static parameter`),
			},
			// Test removing the configuration
			{
				Config: testAccAWSNeptuneParameterGroupConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneParameterGroupExists(resourceName, &v),
					testAccCheckAWSNeptuneParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSNeptuneParameterGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).neptuneconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_neptune_parameter_group" {
			continue
		}

		// Try to find the Group
		resp, err := conn.DescribeDBParameterGroups(
			&neptune.DescribeDBParameterGroupsInput{
				DBParameterGroupName: aws.String(rs.Primary.ID),
			})

		if err != nil {
			if isAWSErr(err, neptune.ErrCodeDBParameterGroupNotFoundFault, "") {
				return nil
			}
			return err
		}

		if len(resp.DBParameterGroups) != 0 && aws.StringValue(resp.DBParameterGroups[0].DBParameterGroupName) == rs.Primary.ID {
			return fmt.Errorf("DB Parameter Group still exists")
		}
	}

	return nil
}

func testAccCheckAWSNeptuneParameterGroupAttributes(v *neptune.DBParameterGroup, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *v.DBParameterGroupName != rName {
			return fmt.Errorf("bad name: %#v", v.DBParameterGroupName)
		}

		if *v.DBParameterGroupFamily != "neptune1" {
			return fmt.Errorf("bad family: %#v", v.DBParameterGroupFamily)
		}

		return nil
	}
}

func testAccCheckAWSNeptuneParameterGroupExists(n string, v *neptune.DBParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Neptune Parameter Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).neptuneconn

		opts := neptune.DescribeDBParameterGroupsInput{
			DBParameterGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDBParameterGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.DBParameterGroups) != 1 ||
			*resp.DBParameterGroups[0].DBParameterGroupName != rs.Primary.ID {
			return fmt.Errorf("Neptune Parameter Group not found")
		}

		*v = *resp.DBParameterGroups[0]

		return nil
	}
}

func testAccAWSNeptuneParameterGroupConfig_Parameter(rName, pName, pValue, pApplyMethod string) string {
	return fmt.Sprintf(`
resource "aws_neptune_parameter_group" "test" {
  family = "neptune1"
  name   = "%s"

  parameter {
    apply_method = "%s"
    name         = "%s"
    value        = "%s"
  }
}`, rName, pApplyMethod, pName, pValue)
}

func testAccAWSNeptuneParameterGroupConfig_Description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_neptune_parameter_group" "test" {
  description = "%s"
  family      = "neptune1"
  name        = "%s"
}`, description, rName)
}

func testAccAWSNeptuneParameterGroupConfig_Required(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptune_parameter_group" "test" {
  family = "neptune1"
  name   = "%s"
}`, rName)
}
