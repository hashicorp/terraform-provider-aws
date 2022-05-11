package memorydb_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/memorydb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmemorydb "github.com/hashicorp/terraform-provider-aws/internal/service/memorydb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccMemoryDBParameterGroup_basic(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "memorydb", "parametergroup/"+rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "family", "memorydb_redis6"),
					resource.TestCheckResourceAttr(resourceName, "id", rName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "active-defrag-cycle-max",
						"value": "70",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "active-defrag-cycle-min",
						"value": "10",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Test", "test"),
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

func TestAccMemoryDBParameterGroup_disappears(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfmemorydb.ResourceParameterGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMemoryDBParameterGroup_update_parameters(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_withParameter0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_withParameter1(rName, "timeout", "0"), // 0 is the default value
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "timeout",
						"value": "0",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Setting timeout to its default value will cause
				// the import to diverge on the initial read.
				ImportStateVerifyIgnore: []string{"parameter"},
			},
			{
				Config: testAccParameterGroupConfig_withParameter1(rName, "timeout", "20"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "timeout",
						"value": "20",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_withParameter2(rName, "timeout", "20", "activerehashing", "no"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "timeout",
						"value": "20",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "activerehashing",
						"value": "no",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_withParameter1(rName, "timeout", "20"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "timeout",
						"value": "20",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_withParameter0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName),
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

func TestAccMemoryDBParameterGroup_update_tags(t *testing.T) {
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, memorydb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_withTags0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_withTags2(rName, "Key1", "value1", "Key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key2", "value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_withTags1(rName, "Key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_withTags0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
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

func testAccCheckParameterGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_memorydb_parameter_group" {
			continue
		}

		_, err := tfmemorydb.FindParameterGroupByName(context.Background(), conn, rs.Primary.Attributes["name"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("MemoryDB Parameter Group %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckParameterGroupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MemoryDB Parameter Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBConn

		_, err := tfmemorydb.FindParameterGroupByName(context.Background(), conn, rs.Primary.Attributes["name"])

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccParameterGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_parameter_group" "test" {
  name   = %[1]q
  family = "memorydb_redis6"

  parameter {
    name  = "active-defrag-cycle-max"
    value = "70"
  }

  parameter {
    name  = "active-defrag-cycle-min"
    value = "10"
  }

  tags = {
    Test = "test"
  }
}
`, rName)
}

func testAccParameterGroupConfig_withParameter0(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_parameter_group" "test" {
  name   = %[1]q
  family = "memorydb_redis6"
}
`, rName)
}

func testAccParameterGroupConfig_withParameter1(rName, paramName1, paramValue1 string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_parameter_group" "test" {
  name   = %[1]q
  family = "memorydb_redis6"

  parameter {
    name  = %[2]q
    value = %[3]q
  }
}
`, rName, paramName1, paramValue1)
}

func testAccParameterGroupConfig_withParameter2(rName, paramName1, paramValue1, paramName2, paramValue2 string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_parameter_group" "test" {
  name   = %[1]q
  family = "memorydb_redis6"

  parameter {
    name  = %[2]q
    value = %[3]q
  }

  parameter {
    name  = %[4]q
    value = %[5]q
  }
}
`, rName, paramName1, paramValue1, paramName2, paramValue2)
}

func testAccParameterGroupConfig_withTags0(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_parameter_group" "test" {
  name   = %[1]q
  family = "memorydb_redis6"
}
`, rName)
}

func testAccParameterGroupConfig_withTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_parameter_group" "test" {
  name   = %[1]q
  family = "memorydb_redis6"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccParameterGroupConfig_withTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_parameter_group" "test" {
  name   = %[1]q
  family = "memorydb_redis6"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

// TestParameterChanges was copy-pasted from the ElastiCache implementation.
func TestParameterChanges(t *testing.T) {
	cases := []struct {
		Name                string
		Old                 *schema.Set
		New                 *schema.Set
		ExpectedRemove      []*memorydb.ParameterNameValue
		ExpectedAddOrUpdate []*memorydb.ParameterNameValue
	}{
		{
			Name:                "Empty",
			Old:                 new(schema.Set),
			New:                 new(schema.Set),
			ExpectedRemove:      []*memorydb.ParameterNameValue{},
			ExpectedAddOrUpdate: []*memorydb.ParameterNameValue{},
		},
		{
			Name: "Remove all",
			Old: schema.NewSet(tfmemorydb.ParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "reserved-memory",
					"value": "0",
				},
			}),
			New: new(schema.Set),
			ExpectedRemove: []*memorydb.ParameterNameValue{
				{
					ParameterName:  aws.String("reserved-memory"),
					ParameterValue: aws.String("0"),
				},
			},
			ExpectedAddOrUpdate: []*memorydb.ParameterNameValue{},
		},
		{
			Name: "No change",
			Old: schema.NewSet(tfmemorydb.ParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "reserved-memory",
					"value": "0",
				},
			}),
			New: schema.NewSet(tfmemorydb.ParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "reserved-memory",
					"value": "0",
				},
			}),
			ExpectedRemove:      []*memorydb.ParameterNameValue{},
			ExpectedAddOrUpdate: []*memorydb.ParameterNameValue{},
		},
		{
			Name: "Remove partial",
			Old: schema.NewSet(tfmemorydb.ParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "reserved-memory",
					"value": "0",
				},
				map[string]interface{}{
					"name":  "appendonly",
					"value": "yes",
				},
			}),
			New: schema.NewSet(tfmemorydb.ParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "appendonly",
					"value": "yes",
				},
			}),
			ExpectedRemove: []*memorydb.ParameterNameValue{
				{
					ParameterName:  aws.String("reserved-memory"),
					ParameterValue: aws.String("0"),
				},
			},
			ExpectedAddOrUpdate: []*memorydb.ParameterNameValue{},
		},
		{
			Name: "Add to existing",
			Old: schema.NewSet(tfmemorydb.ParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "appendonly",
					"value": "yes",
				},
			}),
			New: schema.NewSet(tfmemorydb.ParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "appendonly",
					"value": "yes",
				},
				map[string]interface{}{
					"name":  "appendfsync",
					"value": "always",
				},
			}),
			ExpectedRemove: []*memorydb.ParameterNameValue{},
			ExpectedAddOrUpdate: []*memorydb.ParameterNameValue{
				{
					ParameterName:  aws.String("appendfsync"),
					ParameterValue: aws.String("always"),
				},
			},
		},
	}

	for _, tc := range cases {
		remove, addOrUpdate := tfmemorydb.ParameterChanges(tc.Old, tc.New)
		if !reflect.DeepEqual(remove, tc.ExpectedRemove) {
			t.Errorf("Case %q: Remove did not match\n%#v\n\nGot:\n%#v", tc.Name, tc.ExpectedRemove, remove)
		}
		if !reflect.DeepEqual(addOrUpdate, tc.ExpectedAddOrUpdate) {
			t.Errorf("Case %q: AddOrUpdate did not match\n%#v\n\nGot:\n%#v", tc.Name, tc.ExpectedAddOrUpdate, addOrUpdate)
		}
	}
}
