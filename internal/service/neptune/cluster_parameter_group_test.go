package neptune_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccNeptuneClusterParameterGroup_basic(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	parameterGroupName := sdkacctest.RandomWithPrefix("cluster-parameter-group-test-terraform")

	resourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("cluster-pg:%s", parameterGroupName)),
					resource.TestCheckResourceAttr(resourceName, "name", parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "family", "neptune1"),
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

func TestAccNeptuneClusterParameterGroup_namePrefix(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccNeptuneClusterParameterGroup_generatedName(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
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

func TestAccNeptuneClusterParameterGroup_description(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"

	parameterGroupName := sdkacctest.RandomWithPrefix("cluster-parameter-group-test-terraform")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_Description(parameterGroupName, "custom description"),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17870
func TestAccNeptuneClusterParameterGroup_NamePrefix_parameter(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"
	prefix := acctest.ResourcePrefix

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_NamePrefix_Parameter(prefix, "neptune_enable_audit_log", "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"apply_method": "pending-reboot",
						"name":         "neptune_enable_audit_log",
						"value":        "1",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
			{
				Config: testAccClusterParameterGroupConfig_NamePrefix_Parameter(prefix, "neptune_enable_audit_log", "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"apply_method": "pending-reboot",
						"name":         "neptune_enable_audit_log",
						"value":        "0",
					}),
				),
			},
		},
	})
}

func TestAccNeptuneClusterParameterGroup_parameter(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"

	parameterGroupName := sdkacctest.RandomWithPrefix("cluster-parameter-group-test-tf")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_Parameter(parameterGroupName, "neptune_enable_audit_log", "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"apply_method": "pending-reboot",
						"name":         "neptune_enable_audit_log",
						"value":        "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterParameterGroupConfig_Parameter(parameterGroupName, "neptune_enable_audit_log", "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"apply_method": "pending-reboot",
						"name":         "neptune_enable_audit_log",
						"value":        "0",
					}),
				),
			},
		},
	})
}

func TestAccNeptuneClusterParameterGroup_tags(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"

	parameterGroupName := sdkacctest.RandomWithPrefix("cluster-parameter-group-test-tf")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_Tags(parameterGroupName, "key1", "value1"),
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
				Config: testAccClusterParameterGroupConfig_Tags(parameterGroupName, "key1", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value2"),
				),
			},
			{
				Config: testAccClusterParameterGroupConfig_Tags(parameterGroupName, "key2", "value2"),
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
	conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn

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
			if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBParameterGroupNotFoundFault) {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckClusterParameterGroupAttributes(v *neptune.DBClusterParameterGroup, name string) resource.TestCheckFunc {
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

func testAccCheckClusterParameterGroupExists(n string, v *neptune.DBClusterParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Neptune Cluster Parameter Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn

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

func testAccClusterParameterGroupConfig_Description(name, description string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "test" {
  description = "%s"
  family      = "neptune1"
  name        = "%s"
}
`, description, name)
}

func testAccClusterParameterGroupConfig_Parameter(name, pName, pValue string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "test" {
  family = "neptune1"
  name   = "%s"

  parameter {
    name  = "%s"
    value = "%s"
  }
}
`, name, pName, pValue)
}

func testAccClusterParameterGroupConfig_NamePrefix_Parameter(namePrefix, pName, pValue string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "test" {
  family      = "neptune1"
  name_prefix = %q

  parameter {
    name  = %q
    value = %q
  }
}
`, namePrefix, pName, pValue)
}

func testAccClusterParameterGroupConfig_Tags(name, tKey, tValue string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "test" {
  family = "neptune1"
  name   = "%s"

  tags = {
    %s = "%s"
  }
}
`, name, tKey, tValue)
}

func testAccClusterParameterGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "test" {
  family = "neptune1"
  name   = "%s"
}
`, name)
}

const testAccClusterParameterGroupConfig_namePrefix = `
resource "aws_neptune_cluster_parameter_group" "test" {
  family      = "neptune1"
  name_prefix = "tf-test-"
}
`

const testAccClusterParameterGroupConfig_generatedName = `
resource "aws_neptune_cluster_parameter_group" "test" {
  family = "neptune1"
}
`
