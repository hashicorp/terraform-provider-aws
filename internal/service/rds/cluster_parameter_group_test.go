package rds_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
)

func TestAccRDSClusterParameterGroup_basic(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
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
					resource.TestCheckResourceAttr(resourceName, "family", "aurora5.6"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test cluster parameter group for terraform"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterParameterGroupAddParametersConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "name", parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "family", "aurora5.6"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test cluster parameter group for terraform"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "collation_connection",
						"value": "utf8_unicode_ci",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "collation_server",
						"value": "utf8_unicode_ci",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
			{
				Config: testAccClusterParameterGroupConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					testAccCheckClusterParameterNotUserDefined(resourceName, "collation_connection"),
					testAccCheckClusterParameterNotUserDefined(resourceName, "collation_server"),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
				),
			},
		},
	})
}

func TestAccRDSClusterParameterGroup_withApplyMethod(t *testing.T) {
	var v rds.DBClusterParameterGroup
	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-terraform-%d", sdkacctest.RandInt())
	resourceName := "aws_rds_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupWithApplyMethodConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("cluster-pg:%s", parameterGroupName)),
					resource.TestCheckResourceAttr(resourceName, "name", parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "family", "aurora5.6"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test cluster parameter group for terraform"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":         "character_set_server",
						"value":        "utf8",
						"apply_method": "immediate",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":         "character_set_client",
						"value":        "utf8",
						"apply_method": "pending-reboot",
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

func TestAccRDSClusterParameterGroup_namePrefix(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupNamePrefixConfig("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
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

func TestAccRDSClusterParameterGroup_NamePrefix_parameter(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupNamePrefixParameterConfig("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
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

func TestAccRDSClusterParameterGroup_generatedName(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", resource.UniqueIdPrefix),
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

func TestAccRDSClusterParameterGroup_GeneratedName_parameter(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_generatedName_Parameter,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", resource.UniqueIdPrefix),
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

func TestAccRDSClusterParameterGroup_disappears(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccClusterParameterGroupDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRDSClusterParameterGroup_only(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-tf-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupOnlyConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(
						resourceName, "name", parameterGroupName),
					resource.TestCheckResourceAttr(
						resourceName, "family", "aurora5.6"),
					resource.TestCheckResourceAttr(
						resourceName, "description", "Managed by Terraform"),
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

func TestAccRDSClusterParameterGroup_updateParameters(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	groupName := fmt.Sprintf("cluster-parameter-group-test-tf-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupUpdateParametersInitialConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, groupName),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "family", "aurora5.6"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "utf8",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterParameterGroupUpdateParametersUpdatedConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, groupName),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "ascii",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "ascii",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
				),
			},
		},
	})
}

func TestAccRDSClusterParameterGroup_caseParameters(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, rds.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupUpperCaseConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "family", "aurora5.6"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "max_connections",
						"value": "LEAST({DBInstanceClassMemory/6000000},10)",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterParameterGroupUpperCaseConfig(rName),
			},
		},
	})
}

func testAccCheckClusterParameterGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_rds_cluster_parameter_group" {
			continue
		}

		// Try to find the Group
		resp, err := conn.DescribeDBClusterParameterGroups(
			&rds.DescribeDBClusterParameterGroupsInput{
				DBClusterParameterGroupName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBClusterParameterGroups) != 0 &&
				*resp.DBClusterParameterGroups[0].DBClusterParameterGroupName == rs.Primary.ID {
				return errors.New("DB Cluster Parameter Group still exists")
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

func testAccCheckClusterParameterNotUserDefined(n, paramName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Parameter Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		opts := rds.DescribeDBClusterParametersInput{
			DBClusterParameterGroupName: aws.String(rs.Primary.ID),
		}

		userDefined := false
		out, err := conn.DescribeDBClusterParameters(&opts)
		for _, param := range out.Parameters {
			if *param.ParameterName == paramName && aws.StringValue(param.ParameterValue) != "" {
				// Some of these resets leave the parameter name present but with a nil value
				userDefined = true
			}
		}

		if userDefined {
			return fmt.Errorf("DB Parameter %s is user defined", paramName)
		}
		return err
	}
}

func testAccCheckClusterParameterGroupAttributes(v *rds.DBClusterParameterGroup, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *v.DBClusterParameterGroupName != name {
			return fmt.Errorf("bad name: %#v expected: %v", *v.DBClusterParameterGroupName, name)
		}

		if *v.DBParameterGroupFamily != "aurora5.6" {
			return fmt.Errorf("bad family: %#v", *v.DBParameterGroupFamily)
		}

		return nil
	}
}

func testAccClusterParameterGroupDisappears(v *rds.DBClusterParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn
		opts := &rds.DeleteDBClusterParameterGroupInput{
			DBClusterParameterGroupName: v.DBClusterParameterGroupName,
		}
		if _, err := conn.DeleteDBClusterParameterGroup(opts); err != nil {
			return err
		}
		return resource.Retry(40*time.Minute, func() *resource.RetryError {
			opts := &rds.DescribeDBClusterParameterGroupsInput{
				DBClusterParameterGroupName: v.DBClusterParameterGroupName,
			}
			_, err := conn.DescribeDBClusterParameterGroups(opts)
			if err != nil {
				dbparamgrouperr, ok := err.(awserr.Error)
				if ok && dbparamgrouperr.Code() == "DBParameterGroupNotFound" {
					return nil
				}
				return resource.NonRetryableError(
					fmt.Errorf("Error retrieving DB Cluster Parameter Groups: %s", err))
			}
			return resource.RetryableError(fmt.Errorf(
				"Waiting for cluster parameter group to be deleted: %v", v.DBClusterParameterGroupName))
		})
	}
}

func testAccCheckClusterParameterGroupExists(n string, v *rds.DBClusterParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No DB Cluster Parameter Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RDSConn

		opts := rds.DescribeDBClusterParameterGroupsInput{
			DBClusterParameterGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDBClusterParameterGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.DBClusterParameterGroups) != 1 ||
			*resp.DBClusterParameterGroups[0].DBClusterParameterGroupName != rs.Primary.ID {
			return errors.New("DB Cluster Parameter Group not found")
		}

		*v = *resp.DBClusterParameterGroups[0]

		return nil
	}
}

func testAccClusterParameterGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name        = "%s"
  family      = "aurora5.6"
  description = "Test cluster parameter group for terraform"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "utf8"
  }

  tags = {
    foo = "bar"
  }
}
`, name)
}

func testAccClusterParameterGroupWithApplyMethodConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name        = "%s"
  family      = "aurora5.6"
  description = "Test cluster parameter group for terraform"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name         = "character_set_client"
    value        = "utf8"
    apply_method = "pending-reboot"
  }

  tags = {
    foo = "bar"
  }
}
`, name)
}

func testAccClusterParameterGroupAddParametersConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name        = "%s"
  family      = "aurora5.6"
  description = "Test cluster parameter group for terraform"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "utf8"
  }

  parameter {
    name  = "collation_server"
    value = "utf8_unicode_ci"
  }

  parameter {
    name  = "collation_connection"
    value = "utf8_unicode_ci"
  }

  tags = {
    foo = "bar"
    baz = "foo"
  }
}
`, name)
}

func testAccClusterParameterGroupOnlyConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name   = "%s"
  family = "aurora5.6"
}
`, name)
}

func testAccClusterParameterGroupUpdateParametersInitialConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name   = "%s"
  family = "aurora5.6"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "utf8"
  }
}
`, name)
}

func testAccClusterParameterGroupUpdateParametersUpdatedConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name   = "%s"
  family = "aurora5.6"

  parameter {
    name  = "character_set_server"
    value = "ascii"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "ascii"
  }
}
`, name)
}

func testAccClusterParameterGroupUpperCaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name   = "%s"
  family = "aurora5.6"

  parameter {
    name  = "max_connections"
    value = "LEAST({DBInstanceClassMemory/6000000},10)"
  }
}
`, rName)
}

func testAccClusterParameterGroupNamePrefixConfig(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name_prefix = %[1]q
  family      = "aurora5.6"
}
`, namePrefix)
}

func testAccClusterParameterGroupNamePrefixParameterConfig(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name_prefix = %[1]q
  family      = "aurora5.6"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

}
`, namePrefix)
}

const testAccClusterParameterGroupConfig_generatedName = `
resource "aws_rds_cluster_parameter_group" "test" {
  family = "aurora5.6"
}
`

const testAccClusterParameterGroupConfig_generatedName_Parameter = `
resource "aws_rds_cluster_parameter_group" "test" {
  family = "aurora5.6"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }
}
`
