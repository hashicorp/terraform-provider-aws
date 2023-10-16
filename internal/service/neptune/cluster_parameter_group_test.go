// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/neptune"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfneptune "github.com/hashicorp/terraform-provider-aws/internal/service/neptune"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccNeptuneClusterParameterGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBClusterParameterGroup

	parameterGroupName := sdkacctest.RandomWithPrefix("cluster-parameter-group-test-terraform")

	resourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, neptune.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_basic(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
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
	ctx := acctest.Context(t)
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, neptune.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_namePrefixEmpty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "name", regexache.MustCompile("^tf-test-")),
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
	ctx := acctest.Context(t)
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, neptune.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
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
	ctx := acctest.Context(t)
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"

	parameterGroupName := sdkacctest.RandomWithPrefix("cluster-parameter-group-test-terraform")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, neptune.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_description(parameterGroupName, "custom description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
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
	ctx := acctest.Context(t)
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"
	prefix := acctest.ResourcePrefix

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, neptune.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_namePrefix(prefix, "neptune_enable_audit_log", "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
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
				Config: testAccClusterParameterGroupConfig_namePrefix(prefix, "neptune_enable_audit_log", "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
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
	ctx := acctest.Context(t)
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"

	parameterGroupName := sdkacctest.RandomWithPrefix("cluster-parameter-group-test-tf")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, neptune.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_one(parameterGroupName, "neptune_enable_audit_log", "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
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
				Config: testAccClusterParameterGroupConfig_one(parameterGroupName, "neptune_enable_audit_log", "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
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

// This test ensures that defining a parameter with a default setting is ignored
// and returns successfully as no changes being applied.
func TestAccNeptuneClusterParameterGroup_parameterDefault(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"

	parameterGroupName := sdkacctest.RandomWithPrefix("cluster-parameter-group-test-tf")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, neptune.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_one(parameterGroupName, "neptune_enable_audit_log", "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"apply_method": "pending-reboot",
						"name":         "neptune_enable_audit_log",
						"value":        "0",
					}),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccNeptuneClusterParameterGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"

	parameterGroupName := sdkacctest.RandomWithPrefix("cluster-parameter-group-test-tf")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, neptune.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_tags(parameterGroupName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
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
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value2"),
				),
			},
			{
				Config: testAccClusterParameterGroupConfig_tags(parameterGroupName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
					testAccCheckClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckClusterParameterGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_neptune_cluster_parameter_group" {
				continue
			}

			_, err := tfneptune.FindDBClusterParameterGroupByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Neptune Cluster Parameter Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
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

func testAccCheckClusterParameterGroupExists(ctx context.Context, n string, v *neptune.DBClusterParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Neptune Cluster Parameter Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn(ctx)

		output, err := tfneptune.FindDBClusterParameterGroupByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClusterParameterGroupConfig_description(name, description string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "test" {
  description = "%s"
  family      = "neptune1"
  name        = "%s"
}
`, description, name)
}

func testAccClusterParameterGroupConfig_one(name, pName, pValue string) string {
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

func testAccClusterParameterGroupConfig_namePrefix(namePrefix, pName, pValue string) string {
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

func testAccClusterParameterGroupConfig_tags(name, tKey, tValue string) string {
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

func testAccClusterParameterGroupConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "test" {
  family = "neptune1"
  name   = "%s"
}
`, name)
}

const testAccClusterParameterGroupConfig_namePrefixEmpty = `
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
