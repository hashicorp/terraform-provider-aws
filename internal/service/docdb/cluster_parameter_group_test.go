package docdb_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdocdb "github.com/hashicorp/terraform-provider-aws/internal/service/docdb"
)

func TestAccDocDBClusterParameterGroup_basic(t *testing.T) {
	var v docdb.DBClusterParameterGroup
	resourceName := "aws_docdb_cluster_parameter_group.bar"

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_basic(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(fmt.Sprintf("cluster-pg:%s$", parameterGroupName))),
					resource.TestCheckResourceAttr(resourceName, "name", parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "family", "docdb3.6"),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccDocDBClusterParameterGroup_systemParameter(t *testing.T) {
	var v docdb.DBClusterParameterGroup
	resourceName := "aws_docdb_cluster_parameter_group.bar"

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_system(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(fmt.Sprintf("cluster-pg:%s$", parameterGroupName))),
					resource.TestCheckResourceAttr(resourceName, "name", parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "family", "docdb3.6"),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameter"},
			},
		},
	})
}

func TestAccDocDBClusterParameterGroup_namePrefix(t *testing.T) {
	var v docdb.DBClusterParameterGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists("aws_docdb_cluster_parameter_group.test", &v),
					resource.TestMatchResourceAttr("aws_docdb_cluster_parameter_group.test", "name", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:            "aws_docdb_cluster_parameter_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccDocDBClusterParameterGroup_generatedName(t *testing.T) {
	var v docdb.DBClusterParameterGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists("aws_docdb_cluster_parameter_group.test", &v),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster_parameter_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDocDBClusterParameterGroup_description(t *testing.T) {
	var v docdb.DBClusterParameterGroup
	resourceName := "aws_docdb_cluster_parameter_group.bar"

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_description(parameterGroupName, "custom description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", "custom description"),
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

func TestAccDocDBClusterParameterGroup_disappears(t *testing.T) {
	var v docdb.DBClusterParameterGroup
	resourceName := "aws_docdb_cluster_parameter_group.bar"

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_basic(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDocDBClusterParameterGroup_parameter(t *testing.T) {
	var v docdb.DBClusterParameterGroup
	resourceName := "aws_docdb_cluster_parameter_group.bar"

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-tf-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_basic2(parameterGroupName, "tls", "disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"apply_method": "pending-reboot",
						"name":         "tls",
						"value":        "disabled",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterParameterGroupConfig_basic2(parameterGroupName, "tls", "enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"apply_method": "pending-reboot",
						"name":         "tls",
						"value":        "enabled",
					}),
				),
			},
		},
	})
}

func TestAccDocDBClusterParameterGroup_tags(t *testing.T) {
	var v docdb.DBClusterParameterGroup
	resourceName := "aws_docdb_cluster_parameter_group.bar"

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-tf-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, docdb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_tags(parameterGroupName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterParameterGroupConfig_tags(parameterGroupName, "key1", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value2"),
				),
			},
			{
				Config: testAccClusterParameterGroupConfig_tags(parameterGroupName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckClusterParameterGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_docdb_cluster_parameter_group" {
			continue
		}

		resp, err := conn.DescribeDBClusterParameterGroups(
			&docdb.DescribeDBClusterParameterGroupsInput{
				DBClusterParameterGroupName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBClusterParameterGroups) != 0 &&
				aws.StringValue(resp.DBClusterParameterGroups[0].DBClusterParameterGroupName) == rs.Primary.ID {
				return errors.New("DocDB Cluster Parameter Group still exists")
			}
		}

		if err != nil {
			if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBParameterGroupNotFoundFault) {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckClusterParameterGroupDisappears(group *docdb.DBClusterParameterGroup) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn

		params := &docdb.DeleteDBClusterParameterGroupInput{
			DBClusterParameterGroupName: group.DBClusterParameterGroupName,
		}

		_, err := conn.DeleteDBClusterParameterGroup(params)
		if err != nil {
			return err
		}

		return tfdocdb.WaitForClusterParameterGroupDeletion(conn, *group.DBClusterParameterGroupName)
	}
}

func testAccCheckClusterParameterGroupAttributes(v *docdb.DBClusterParameterGroup, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *v.DBClusterParameterGroupName != name {
			return fmt.Errorf("bad name: %#v expected: %v", *v.DBClusterParameterGroupName, name)
		}

		if *v.DBParameterGroupFamily != "docdb3.6" {
			return fmt.Errorf("bad family: %#v", *v.DBParameterGroupFamily)
		}

		return nil
	}
}

func testAccCheckClusterParameterGroupExists(n string, v *docdb.DBClusterParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No DocDB Cluster Parameter Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DocDBConn

		opts := docdb.DescribeDBClusterParameterGroupsInput{
			DBClusterParameterGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDBClusterParameterGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.DBClusterParameterGroups) != 1 ||
			aws.StringValue(resp.DBClusterParameterGroups[0].DBClusterParameterGroupName) != rs.Primary.ID {
			return fmt.Errorf("DocDB Cluster Parameter Group not found: %s", rs.Primary.ID)
		}

		*v = *resp.DBClusterParameterGroups[0]

		return nil
	}
}

func testAccClusterParameterGroupConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster_parameter_group" "bar" {
  family = "docdb3.6"
  name   = "%s"
}
`, name)
}

func testAccClusterParameterGroupConfig_system(name string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster_parameter_group" "bar" {
  family = "docdb3.6"
  name   = "%s"

  parameter {
    name         = "tls"
    value        = "enabled"
    apply_method = "pending-reboot"
  }
}
`, name)
}

func testAccClusterParameterGroupConfig_description(name, description string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster_parameter_group" "bar" {
  family      = "docdb3.6"
  description = "%s"
  name        = "%s"
}
`, description, name)
}

func testAccClusterParameterGroupConfig_basic2(name, pName, pValue string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster_parameter_group" "bar" {
  name   = "%s"
  family = "docdb3.6"

  parameter {
    name  = "%s"
    value = "%s"
  }
}
`, name, pName, pValue)
}

func testAccClusterParameterGroupConfig_tags(name, tKey, tValue string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster_parameter_group" "bar" {
  name   = "%s"
  family = "docdb3.6"

  tags = {
    %s = "%s"
  }
}
`, name, tKey, tValue)
}

const testAccClusterParameterGroupConfig_namePrefix = `
resource "aws_docdb_cluster_parameter_group" "test" {
  name_prefix = "tf-test-"
  family      = "docdb3.6"
}
`
const testAccClusterParameterGroupConfig_generatedName = `
resource "aws_docdb_cluster_parameter_group" "test" {
  family = "docdb3.6"
}
`
