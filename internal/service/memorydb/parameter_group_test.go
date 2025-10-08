// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/memorydb/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfmemorydb "github.com/hashicorp/terraform-provider-aws/internal/service/memorydb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestParameterChanges(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name                string
		Old                 *schema.Set
		New                 *schema.Set
		ExpectedRemove      []awstypes.ParameterNameValue
		ExpectedAddOrUpdate []awstypes.ParameterNameValue
	}{
		{
			Name:                "Empty",
			Old:                 new(schema.Set),
			New:                 new(schema.Set),
			ExpectedRemove:      []awstypes.ParameterNameValue{},
			ExpectedAddOrUpdate: []awstypes.ParameterNameValue{},
		},
		{
			Name: "Remove all",
			Old: schema.NewSet(tfmemorydb.ParameterHash, []any{
				map[string]any{
					names.AttrName:  "reserved-memory",
					names.AttrValue: "0",
				},
			}),
			New: new(schema.Set),
			ExpectedRemove: []awstypes.ParameterNameValue{
				{
					ParameterName:  aws.String("reserved-memory"),
					ParameterValue: aws.String("0"),
				},
			},
			ExpectedAddOrUpdate: []awstypes.ParameterNameValue{},
		},
		{
			Name: "No change",
			Old: schema.NewSet(tfmemorydb.ParameterHash, []any{
				map[string]any{
					names.AttrName:  "reserved-memory",
					names.AttrValue: "0",
				},
			}),
			New: schema.NewSet(tfmemorydb.ParameterHash, []any{
				map[string]any{
					names.AttrName:  "reserved-memory",
					names.AttrValue: "0",
				},
			}),
			ExpectedRemove:      []awstypes.ParameterNameValue{},
			ExpectedAddOrUpdate: []awstypes.ParameterNameValue{},
		},
		{
			Name: "Remove partial",
			Old: schema.NewSet(tfmemorydb.ParameterHash, []any{
				map[string]any{
					names.AttrName:  "reserved-memory",
					names.AttrValue: "0",
				},
				map[string]any{
					names.AttrName:  "appendonly",
					names.AttrValue: "yes",
				},
			}),
			New: schema.NewSet(tfmemorydb.ParameterHash, []any{
				map[string]any{
					names.AttrName:  "appendonly",
					names.AttrValue: "yes",
				},
			}),
			ExpectedRemove: []awstypes.ParameterNameValue{
				{
					ParameterName:  aws.String("reserved-memory"),
					ParameterValue: aws.String("0"),
				},
			},
			ExpectedAddOrUpdate: []awstypes.ParameterNameValue{},
		},
		{
			Name: "Add to existing",
			Old: schema.NewSet(tfmemorydb.ParameterHash, []any{
				map[string]any{
					names.AttrName:  "appendonly",
					names.AttrValue: "yes",
				},
			}),
			New: schema.NewSet(tfmemorydb.ParameterHash, []any{
				map[string]any{
					names.AttrName:  "appendonly",
					names.AttrValue: "yes",
				},
				map[string]any{
					names.AttrName:  "appendfsync",
					names.AttrValue: "always",
				},
			}),
			ExpectedRemove: []awstypes.ParameterNameValue{},
			ExpectedAddOrUpdate: []awstypes.ParameterNameValue{
				{
					ParameterName:  aws.String("appendfsync"),
					ParameterValue: aws.String("always"),
				},
			},
		},
	}

	ignoreExportedOpts := cmpopts.IgnoreUnexported(
		awstypes.ParameterNameValue{},
	)

	for _, tc := range cases {
		remove, addOrUpdate := tfmemorydb.ParameterChanges(tc.Old, tc.New)
		if diff := cmp.Diff(remove, tc.ExpectedRemove, ignoreExportedOpts); diff != "" {
			t.Errorf("unexpected Remove diff (+wanted, -got): %s", diff)
		}
		if diff := cmp.Diff(addOrUpdate, tc.ExpectedAddOrUpdate, ignoreExportedOpts); diff != "" {
			t.Errorf("unexpected AddOrUpdate diff (+wanted, -got): %s", diff)
		}
	}
}

func TestAccMemoryDBParameterGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "memorydb", "parametergroup/"+rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "memorydb_redis6"),
					resource.TestCheckResourceAttr(resourceName, names.AttrID, rName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "active-defrag-cycle-max",
						names.AttrValue: "70",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "active-defrag-cycle-min",
						names.AttrValue: "10",
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
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
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfmemorydb.ResourceParameterGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMemoryDBParameterGroup_update_parameters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_none(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_one(rName, names.AttrTimeout, "0"), // 0 is the default value
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  names.AttrTimeout,
						names.AttrValue: "0",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Setting timeout to its default value will cause
				// the import to diverge on the initial read.
				ImportStateVerifyIgnore: []string{names.AttrParameter},
			},
			{
				Config: testAccParameterGroupConfig_one(rName, names.AttrTimeout, "20"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  names.AttrTimeout,
						names.AttrValue: "20",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_multiple(rName, names.AttrTimeout, "20", "activerehashing", "no"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  names.AttrTimeout,
						names.AttrValue: "20",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  "activerehashing",
						names.AttrValue: "no",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_one(rName, names.AttrTimeout, "20"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						names.AttrName:  names.AttrTimeout,
						names.AttrValue: "20",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_none(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName),
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
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_memorydb_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MemoryDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubnetGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_tags0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_tags2(rName, "Key1", acctest.CtValue1, "Key2", acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key2", acctest.CtValue2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_tags1(rName, "Key1", acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key1", acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_tags0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
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

func testAccCheckParameterGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_memorydb_parameter_group" {
				continue
			}

			_, err := tfmemorydb.FindParameterGroupByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

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
}

func testAccCheckParameterGroupExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No MemoryDB Parameter Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MemoryDBClient(ctx)

		_, err := tfmemorydb.FindParameterGroupByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

		return err
	}
}

func testAccParameterGroupConfig_basic(rName string) string {
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

func testAccParameterGroupConfig_none(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_parameter_group" "test" {
  name   = %[1]q
  family = "memorydb_redis6"
}
`, rName)
}

func testAccParameterGroupConfig_one(rName, paramName1, paramValue1 string) string {
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

func testAccParameterGroupConfig_multiple(rName, paramName1, paramValue1, paramName2, paramValue2 string) string {
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

func testAccParameterGroupConfig_tags0(rName string) string {
	return fmt.Sprintf(`
resource "aws_memorydb_parameter_group" "test" {
  name   = %[1]q
  family = "memorydb_redis6"
}
`, rName)
}

func testAccParameterGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
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

func testAccParameterGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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
