package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSNeptuneParameterGroup_basic(t *testing.T) {
	var v neptune.DBParameterGroup
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneParameterGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSNeptuneParameterGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneParameterGroupExists("aws_neptune_parameter_group.bar", &v),
					testAccCheckAWSNeptuneParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(
						"aws_neptune_parameter_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_neptune_parameter_group.bar", "family", "neptune1"),
					resource.TestCheckResourceAttr(
						"aws_neptune_parameter_group.bar", "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr("aws_neptune_parameter_group.bar", "parameter.562386247", "1"),
					resource.TestCheckResourceAttr(
						"aws_neptune_parameter_group.bar", "parameter.562386247.name", "neptune_query_timeout"),
					resource.TestCheckResourceAttr(
						"aws_neptune_parameter_group.bar", "parameter.562386247.value", "25"),
				),
			},
			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSNeptuneParameterGroup_only(t *testing.T) {
	var v neptune.DBParameterGroup
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneParameterGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSNeptuneParameterGroupOnlyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneParameterGroupExists("aws_neptune_parameter_group.bar", &v),
					testAccCheckAWSNeptuneParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(
						"aws_neptune_parameter_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_neptune_parameter_group.bar", "family", "neptune1"),
				),
			},
		},
	})
}

// Regression for https://github.com/terraform-providers/terraform-provider-aws/issues/116
func TestAccAWSNeptuneParameterGroup_removeParam(t *testing.T) {
	var v neptune.DBParameterGroup
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneParameterGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneParameterGroupExists("aws_neptune_parameter_group.bar", &v),
					testAccCheckAWSNeptuneParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(
						"aws_neptune_parameter_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_neptune_parameter_group.bar", "family", "neptune1"),
					resource.TestCheckResourceAttr(
						"aws_neptune_parameter_group.bar", "parameter.562386247.name", "neptune_query_timeout"),
					resource.TestCheckResourceAttr(
						"aws_neptune_parameter_group.bar", "parameter.562386247.value", "25"),
				),
			},
			{
				Config: testAccAWSNeptuneParameterGroupOnlyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneParameterGroupExists("aws_neptune_parameter_group.bar", &v),
					testAccCheckAWSNeptuneParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(
						"aws_neptune_parameter_group.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_neptune_parameter_group.bar", "family", "neptune1"),

					resource.TestCheckNoResourceAttr(
						"aws_neptune_parameter_group.bar", "parameter.562386247.name"),
					resource.TestCheckNoResourceAttr(
						"aws_neptune_parameter_group.bar", "parameter.562386247.value"),
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

		if err == nil {
			if len(resp.DBParameterGroups) != 0 &&
				*resp.DBParameterGroups[0].DBParameterGroupName == rs.Primary.ID {
				return fmt.Errorf("DB Parameter Group still exists")
			}
		}

		// Verify the error
		newerr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if newerr.Code() != "DBParameterGroupNotFound" {
			return err
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

func testAccAWSNeptuneParameterGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptune_parameter_group" "bar" {
	name = "%s"
	family = "neptune1"
	parameter {
	  name = "neptune_query_timeout"
	  value = "25"
          apply_method = "pending-reboot"

	}
}`, rName)
}

func testAccAWSNeptuneParameterGroupOnlyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptune_parameter_group" "bar" {
	name = "%s"
	family = "neptune1"
	description = "Test parameter group for terraform"
}`, rName)
}
