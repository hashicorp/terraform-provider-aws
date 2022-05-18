package redshift_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccRedshiftParameterGroup_basic(t *testing.T) {
	var v redshift.ClusterParameterGroup
	resourceName := "aws_redshift_parameter_group.test"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
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

func TestAccRedshiftParameterGroup_withParameters(t *testing.T) {
	var v redshift.ClusterParameterGroup
	rInt := sdkacctest.RandInt()
	resourceName := "aws_redshift_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("test-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "family", "redshift-1.0"),
					resource.TestCheckResourceAttr(
						resourceName, "description", "Managed by Terraform"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "require_ssl",
						"value": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "query_group",
						"value": "example",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "enable_user_activity_logging",
						"value": "true",
					}),
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

func TestAccRedshiftParameterGroup_withoutParameters(t *testing.T) {
	var v redshift.ClusterParameterGroup
	rInt := sdkacctest.RandInt()
	resourceName := "aws_redshift_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupOnlyConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("test-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "family", "redshift-1.0"),
					resource.TestCheckResourceAttr(
						resourceName, "description", "Test parameter group for terraform"),
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

func TestAccRedshiftParameterGroup_withTags(t *testing.T) {
	var v redshift.ClusterParameterGroup
	rInt := sdkacctest.RandInt()
	resourceName := "aws_redshift_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, redshift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupWithTagsConfig(rInt, "aaa"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.name", fmt.Sprintf("test-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "tags.description", fmt.Sprintf("Test parameter group for terraform %s", "aaa")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupWithTagsConfig(rInt, "bbb"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.description", fmt.Sprintf("Test parameter group for terraform %s", "bbb")),
				),
			},
			{
				Config: testAccParameterGroupWithTagsUpdateConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.name", fmt.Sprintf("test-terraform-%d", rInt)),
				),
			},
		},
	})
}

func testAccCheckParameterGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_parameter_group" {
			continue
		}

		// Try to find the Group
		resp, err := conn.DescribeClusterParameterGroups(
			&redshift.DescribeClusterParameterGroupsInput{
				ParameterGroupName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.ParameterGroups) != 0 &&
				*resp.ParameterGroups[0].ParameterGroupName == rs.Primary.ID {
				return fmt.Errorf("Redshift Parameter Group still exists")
			}
		}

		// Verify the error
		newerr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if newerr.Code() != "ClusterParameterGroupNotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckParameterGroupExists(n string, v *redshift.ClusterParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Parameter Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn

		opts := redshift.DescribeClusterParameterGroupsInput{
			ParameterGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeClusterParameterGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.ParameterGroups) != 1 ||
			*resp.ParameterGroups[0].ParameterGroupName != rs.Primary.ID {
			return fmt.Errorf("Redshift Parameter Group not found")
		}

		*v = *resp.ParameterGroups[0]

		return nil
	}
}

func testAccParameterGroupOnlyConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_redshift_parameter_group" "test" {
  name        = "test-terraform-%d"
  family      = "redshift-1.0"
  description = "Test parameter group for terraform"
}
`, rInt)
}

func testAccParameterGroupConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_redshift_parameter_group" "test" {
  name   = "test-terraform-%d"
  family = "redshift-1.0"

  parameter {
    name  = "require_ssl"
    value = "true"
  }

  parameter {
    name  = "query_group"
    value = "example"
  }

  parameter {
    name  = "enable_user_activity_logging"
    value = "true"
  }
}
`, rInt)
}

func testAccParameterGroupWithTagsConfig(rInt int, rString string) string {
	return fmt.Sprintf(`
resource "aws_redshift_parameter_group" "test" {
  name        = "test-terraform-%[1]d"
  family      = "redshift-1.0"
  description = "Test parameter group for terraform"

  tags = {
    environment = "Production"
    name        = "test-terraform-%[1]d"
    description = "Test parameter group for terraform %[2]s"
  }
}
`, rInt, rString)
}

func testAccParameterGroupWithTagsUpdateConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_redshift_parameter_group" "test" {
  name        = "test-terraform-%[1]d"
  family      = "redshift-1.0"
  description = "Test parameter group for terraform"

  tags = {
    name = "test-terraform-%[1]d"
  }
}
`, rInt)
}
