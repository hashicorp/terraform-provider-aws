package aws

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSNeptuneClusterParameterGroup_basic(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-terraform-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists("aws_neptune_cluster_parameter_group.bar", &v),
					testAccCheckAWSNeptuneClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster_parameter_group.bar", "name", parameterGroupName),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster_parameter_group.bar", "family", "neptune1"),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster_parameter_group.bar", "description", "Test cluster parameter group for terraform"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "parameter.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster_parameter_group.bar", "tags.%", "1"),
				),
			},
		},
	})
}

func TestAccAWSNeptuneClusterParameterGroup_namePrefix(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists("aws_neptune_cluster_parameter_group.test", &v),
					resource.TestMatchResourceAttr(
						"aws_neptune_cluster_parameter_group.test", "name", regexp.MustCompile("^tf-test-")),
				),
			},
		},
	})
}

func TestAccAWSNeptuneClusterParameterGroup_generatedName(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists("aws_neptune_cluster_parameter_group.test", &v),
				),
			},
		},
	})
}

func TestAccAWSNeptuneClusterParameterGroup_withoutParameter(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-tf-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterParameterGroupOnlyConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists("aws_neptune_cluster_parameter_group.bar", &v),
					testAccCheckAWSNeptuneClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster_parameter_group.bar", "name", parameterGroupName),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster_parameter_group.bar", "family", "neptune1"),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster_parameter_group.bar", "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "parameter.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSNeptuneClusterParameterGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).neptuneconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_neptune_cluster_parameter_group" {
			continue
		}

		resp, err := conn.DescribeDBClusterParameterGroups(
			&neptune.DescribeDBClusterParameterGroupsInput{
				DBClusterParameterGroupName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBClusterParameterGroups) != 0 &&
				aws.StringValue(resp.DBClusterParameterGroups[0].DBClusterParameterGroupName) == rs.Primary.ID {
				return errors.New("Neptune Cluster Parameter Group still exists")
			}
		}

		if err != nil {
			if isAWSErr(err, neptune.ErrCodeDBParameterGroupNotFoundFault, "") {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAWSNeptuneClusterParameterGroupAttributes(v *neptune.DBClusterParameterGroup, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *v.DBClusterParameterGroupName != name {
			return fmt.Errorf("bad name: %#v expected: %v", *v.DBClusterParameterGroupName, name)
		}

		if *v.DBParameterGroupFamily != "neptune1" {
			return fmt.Errorf("bad family: %#v", *v.DBParameterGroupFamily)
		}

		return nil
	}
}

func testAccCheckAWSNeptuneClusterParameterGroupExists(n string, v *neptune.DBClusterParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Neptune Cluster Parameter Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).neptuneconn

		opts := neptune.DescribeDBClusterParameterGroupsInput{
			DBClusterParameterGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDBClusterParameterGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.DBClusterParameterGroups) != 1 ||
			aws.StringValue(resp.DBClusterParameterGroups[0].DBClusterParameterGroupName) != rs.Primary.ID {
			return errors.New("Neptune Cluster Parameter Group not found")
		}

		*v = *resp.DBClusterParameterGroups[0]

		return nil
	}
}

func testAccAWSNeptuneClusterParameterGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "bar" {
  name        = "%s"
  family      = "neptune1"
  description = "Test cluster parameter group for terraform"

  parameter {
    name  = "neptune_enable_audit_log"
    value = 1
  }

  tags {
    foo = "bar"
  }
}
`, name)
}

func testAccAWSNeptuneClusterParameterGroupOnlyConfig(name string) string {
	return fmt.Sprintf(`resource "aws_neptune_cluster_parameter_group" "bar" {
  name        = "%s"
  family      = "neptune1"
  description = "Managed by Terraform"
}`, name)
}

const testAccAWSNeptuneClusterParameterGroupConfig_namePrefix = `
resource "aws_neptune_cluster_parameter_group" "test" {
  name_prefix = "tf-test-"
  family = "neptune1"
}
`
const testAccAWSNeptuneClusterParameterGroupConfig_generatedName = `
resource "aws_neptune_cluster_parameter_group" "test" {
  family = "neptune1"
}
`
