package neptune_test

import (
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

func TestAccNeptuneParameterGroup_basic(t *testing.T) {
	var v neptune.DBParameterGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("pg:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "family", "neptune1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccNeptuneParameterGroup_description(t *testing.T) {
	var v neptune.DBParameterGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_Description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, rName),
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

func TestAccNeptuneParameterGroup_parameter(t *testing.T) {
	var v neptune.DBParameterGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_Parameter(rName, "neptune_query_timeout", "25", "pending-reboot"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"apply_method": "pending-reboot",
						"name":         "neptune_query_timeout",
						"value":        "25",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// This test should be updated with a dynamic parameter when available
			{
				Config:      testAccParameterGroupConfig_Parameter(rName, "neptune_query_timeout", "25", "immediate"),
				ExpectError: regexp.MustCompile(`cannot use immediate apply method for static parameter`),
			},
			// Test removing the configuration
			{
				Config: testAccParameterGroupConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "0"),
				),
			},
		},
	})
}

func TestAccNeptuneParameterGroup_tags(t *testing.T) {
	var v neptune.DBParameterGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, neptune.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_Tags_SingleTag(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
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
				Config: testAccParameterGroupConfig_Tags_SingleTag(rName, "key1", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value2"),
				),
			},
			{
				Config: testAccParameterGroupConfig_Tags_MultipleTags(rName, "key2", "value2", "key3", "value3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key3", "value3"),
				),
			},
		},
	})
}

func testAccCheckParameterGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn

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
			if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBParameterGroupNotFoundFault) {
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

func testAccCheckParameterGroupAttributes(v *neptune.DBParameterGroup, rName string) resource.TestCheckFunc {
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

func testAccCheckParameterGroupExists(n string, v *neptune.DBParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Neptune Parameter Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn

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

func testAccParameterGroupConfig_Parameter(rName, pName, pValue, pApplyMethod string) string {
	return fmt.Sprintf(`
resource "aws_neptune_parameter_group" "test" {
  family = "neptune1"
  name   = %q

  parameter {
    apply_method = %q
    name         = %q
    value        = %q
  }
}
`, rName, pApplyMethod, pName, pValue)
}

func testAccParameterGroupConfig_Description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_neptune_parameter_group" "test" {
  description = %q
  family      = "neptune1"
  name        = %q
}
`, description, rName)
}

func testAccParameterGroupConfig_Required(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptune_parameter_group" "test" {
  family = "neptune1"
  name   = %q
}
`, rName)
}

func testAccParameterGroupConfig_Tags_SingleTag(name, tKey, tValue string) string {
	return fmt.Sprintf(`
resource "aws_neptune_parameter_group" "test" {
  family = "neptune1"
  name   = %q

  tags = {
    %s = %q
  }
}
`, name, tKey, tValue)
}

func testAccParameterGroupConfig_Tags_MultipleTags(name, tKey1, tValue1, tKey2, tValue2 string) string {
	return fmt.Sprintf(`
resource "aws_neptune_parameter_group" "test" {
  family = "neptune1"
  name   = %q

  tags = {
    %s = %q
    %s = %q
  }
}
`, name, tKey1, tValue1, tKey2, tValue2)
}
