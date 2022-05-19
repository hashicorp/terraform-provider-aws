package elasticache_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
)

func TestAccElastiCacheParameterGroup_basic(t *testing.T) {
	var v elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					testAccCheckParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "family", "redis2.8"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccElastiCacheParameterGroup_addParameter(t *testing.T) {
	var v elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupParameter1Config(rName, "redis2.8", "appendonly", "yes"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "appendonly",
						"value": "yes",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupParameter2Config(rName, "redis2.8", "appendonly", "yes", "appendfsync", "always"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "appendonly",
						"value": "yes",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "appendfsync",
						"value": "always",
					}),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform-provider-aws/issues/116
func TestAccElastiCacheParameterGroup_removeAllParameters(t *testing.T) {
	var v elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupParameter2Config(rName, "redis2.8", "appendonly", "yes", "appendfsync", "always"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "appendonly",
						"value": "yes",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "appendfsync",
						"value": "always",
					}),
				),
			},
			{
				Config: testAccParameterGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "0"),
				),
			},
		},
	})
}

// The API returns errors when attempting to reset the reserved-memory parameter.
// This covers our custom logic handling for this situation.
func TestAccElastiCacheParameterGroup_RemoveReservedMemoryParameter_allParameters(t *testing.T) {
	var cacheParameterGroup1 elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupParameter1Config(rName, "redis3.2", "reserved-memory", "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "reserved-memory",
						"value": "0",
					}),
				),
			},
			{
				Config: testAccParameterGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &cacheParameterGroup1),
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

// The API returns errors when attempting to reset the reserved-memory parameter.
// This covers our custom logic handling for this situation.
func TestAccElastiCacheParameterGroup_RemoveReservedMemoryParameter_remainingParameters(t *testing.T) {
	var cacheParameterGroup1 elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupParameter2Config(rName, "redis3.2", "reserved-memory", "0", "tcp-keepalive", "360"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "reserved-memory",
						"value": "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "tcp-keepalive",
						"value": "360",
					}),
				),
			},
			{
				Config: testAccParameterGroupParameter1Config(rName, "redis3.2", "tcp-keepalive", "360"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "tcp-keepalive",
						"value": "360",
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

// The API returns errors when attempting to reset the reserved-memory parameter.
// This covers our custom logic handling for this situation.
func TestAccElastiCacheParameterGroup_switchReservedMemoryParameter(t *testing.T) {
	var cacheParameterGroup1 elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupParameter1Config(rName, "redis3.2", "reserved-memory", "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "reserved-memory",
						"value": "0",
					}),
				),
			},
			{
				Config: testAccParameterGroupParameter1Config(rName, "redis3.2", "reserved-memory-percent", "25"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "reserved-memory-percent",
						"value": "25",
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

// The API returns errors when attempting to reset the reserved-memory parameter.
// This covers our custom logic handling for this situation.
func TestAccElastiCacheParameterGroup_updateReservedMemoryParameter(t *testing.T) {
	var cacheParameterGroup1 elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupParameter1Config(rName, "redis2.8", "reserved-memory", "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "reserved-memory",
						"value": "0",
					}),
				),
			},
			{
				Config: testAccParameterGroupParameter1Config(rName, "redis2.8", "reserved-memory", "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "reserved-memory",
						"value": "1",
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

func TestAccElastiCacheParameterGroup_uppercaseName(t *testing.T) {
	var v elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rInt := sdkacctest.RandInt()
	rName := fmt.Sprintf("TF-ELASTIPG-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupParameter1Config(rName, "redis2.8", "appendonly", "yes"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-elastipg-%d", rInt)),
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

func TestAccElastiCacheParameterGroup_description(t *testing.T) {
	var v elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupDescriptionConfig(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &v),
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

func TestAccElastiCacheParameterGroup_tags(t *testing.T) {
	var cacheParameterGroup1 elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupTags1Config(rName, "redis2.8", "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccParameterGroupTags2Config(rName, "redis2.8", "key1", "updatedvalue1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "updatedvalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccParameterGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(resourceName, &cacheParameterGroup1),
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

func testAccCheckParameterGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_parameter_group" {
			continue
		}

		resp, err := conn.DescribeCacheParameterGroups(
			&elasticache.DescribeCacheParameterGroupsInput{
				CacheParameterGroupName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.CacheParameterGroups) != 0 &&
				*resp.CacheParameterGroups[0].CacheParameterGroupName == rs.Primary.ID {
				return fmt.Errorf("Cache Parameter Group still exists")
			}
		}

		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheParameterGroupNotFoundFault) {
			return nil
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckParameterGroupAttributes(v *elasticache.CacheParameterGroup, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *v.CacheParameterGroupName != rName {
			return fmt.Errorf("bad name: %#v", v.CacheParameterGroupName)
		}

		if *v.CacheParameterGroupFamily != "redis2.8" {
			return fmt.Errorf("bad family: %#v", v.CacheParameterGroupFamily)
		}

		return nil
	}
}

func testAccCheckParameterGroupExists(n string, v *elasticache.CacheParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cache Parameter Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn

		opts := elasticache.DescribeCacheParameterGroupsInput{
			CacheParameterGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeCacheParameterGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.CacheParameterGroups) != 1 ||
			*resp.CacheParameterGroups[0].CacheParameterGroupName != rs.Primary.ID {
			return fmt.Errorf("Cache Parameter Group not found")
		}

		*v = *resp.CacheParameterGroups[0]

		return nil
	}
}

func testAccParameterGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  family = "redis2.8"
  name   = %q
}
`, rName)
}

func testAccParameterGroupDescriptionConfig(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  description = %q
  family      = "redis2.8"
  name        = %q
}
`, description, rName)
}

func testAccParameterGroupParameter1Config(rName, family, parameterName1, parameterValue1 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  family = %q
  name   = %q

  parameter {
    name  = %q
    value = %q
  }
}
`, family, rName, parameterName1, parameterValue1)
}

func testAccParameterGroupParameter2Config(rName, family, parameterName1, parameterValue1, parameterName2, parameterValue2 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  family = %q
  name   = %q

  parameter {
    name  = %q
    value = %q
  }

  parameter {
    name  = %q
    value = %q
  }
}
`, family, rName, parameterName1, parameterValue1, parameterName2, parameterValue2)
}

func testAccParameterGroupTags1Config(rName, family, tagName1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  family = %[1]q
  name   = %[2]q

  tags = {
    %[3]s = %[4]q
  }
}
`, family, rName, tagName1, tagValue1)
}

func testAccParameterGroupTags2Config(rName, family, tagName1, tagValue1, tagName2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  family = %[1]q
  name   = %[2]q

  tags = {
    %[3]s = %[4]q
    %[5]s = %[6]q
  }
}
`, family, rName, tagName1, tagValue1, tagName2, tagValue2)
}

func TestFlattenParameters(t *testing.T) {
	cases := []struct {
		Input  []*elasticache.Parameter
		Output []map[string]interface{}
	}{
		{
			Input: []*elasticache.Parameter{
				{
					ParameterName:  aws.String("activerehashing"),
					ParameterValue: aws.String("yes"),
				},
			},
			Output: []map[string]interface{}{
				{
					"name":  "activerehashing",
					"value": "yes",
				},
			},
		},
	}

	for _, tc := range cases {
		output := tfelasticache.FlattenParameters(tc.Input)
		if !reflect.DeepEqual(output, tc.Output) {
			t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v", output, tc.Output)
		}
	}
}

func TestExpandParameters(t *testing.T) {
	expanded := []interface{}{
		map[string]interface{}{
			"name":         "activerehashing",
			"value":        "yes",
			"apply_method": "immediate",
		},
	}
	parameters := tfelasticache.ExpandParameters(expanded)

	expected := &elasticache.ParameterNameValue{
		ParameterName:  aws.String("activerehashing"),
		ParameterValue: aws.String("yes"),
	}

	if !reflect.DeepEqual(parameters[0], expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			parameters[0],
			expected)
	}
}

func TestParameterChanges(t *testing.T) {
	cases := []struct {
		Name                string
		Old                 *schema.Set
		New                 *schema.Set
		ExpectedRemove      []*elasticache.ParameterNameValue
		ExpectedAddOrUpdate []*elasticache.ParameterNameValue
	}{
		{
			Name:                "Empty",
			Old:                 new(schema.Set),
			New:                 new(schema.Set),
			ExpectedRemove:      []*elasticache.ParameterNameValue{},
			ExpectedAddOrUpdate: []*elasticache.ParameterNameValue{},
		},
		{
			Name: "Remove all",
			Old: schema.NewSet(tfelasticache.ParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "reserved-memory",
					"value": "0",
				},
			}),
			New: new(schema.Set),
			ExpectedRemove: []*elasticache.ParameterNameValue{
				{
					ParameterName:  aws.String("reserved-memory"),
					ParameterValue: aws.String("0"),
				},
			},
			ExpectedAddOrUpdate: []*elasticache.ParameterNameValue{},
		},
		{
			Name: "No change",
			Old: schema.NewSet(tfelasticache.ParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "reserved-memory",
					"value": "0",
				},
			}),
			New: schema.NewSet(tfelasticache.ParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "reserved-memory",
					"value": "0",
				},
			}),
			ExpectedRemove:      []*elasticache.ParameterNameValue{},
			ExpectedAddOrUpdate: []*elasticache.ParameterNameValue{},
		},
		{
			Name: "Remove partial",
			Old: schema.NewSet(tfelasticache.ParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "reserved-memory",
					"value": "0",
				},
				map[string]interface{}{
					"name":  "appendonly",
					"value": "yes",
				},
			}),
			New: schema.NewSet(tfelasticache.ParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "appendonly",
					"value": "yes",
				},
			}),
			ExpectedRemove: []*elasticache.ParameterNameValue{
				{
					ParameterName:  aws.String("reserved-memory"),
					ParameterValue: aws.String("0"),
				},
			},
			ExpectedAddOrUpdate: []*elasticache.ParameterNameValue{},
		},
		{
			Name: "Add to existing",
			Old: schema.NewSet(tfelasticache.ParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "appendonly",
					"value": "yes",
				},
			}),
			New: schema.NewSet(tfelasticache.ParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "appendonly",
					"value": "yes",
				},
				map[string]interface{}{
					"name":  "appendfsync",
					"value": "always",
				},
			}),
			ExpectedRemove: []*elasticache.ParameterNameValue{},
			ExpectedAddOrUpdate: []*elasticache.ParameterNameValue{
				{
					ParameterName:  aws.String("appendfsync"),
					ParameterValue: aws.String("always"),
				},
			},
		},
	}

	for _, tc := range cases {
		remove, addOrUpdate := tfelasticache.ParameterChanges(tc.Old, tc.New)
		if !reflect.DeepEqual(remove, tc.ExpectedRemove) {
			t.Errorf("Case %q: Remove did not match\n%#v\n\nGot:\n%#v", tc.Name, tc.ExpectedRemove, remove)
		}
		if !reflect.DeepEqual(addOrUpdate, tc.ExpectedAddOrUpdate) {
			t.Errorf("Case %q: AddOrUpdate did not match\n%#v\n\nGot:\n%#v", tc.Name, tc.ExpectedAddOrUpdate, addOrUpdate)
		}
	}
}
