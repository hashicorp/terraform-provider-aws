// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfneptune "github.com/hashicorp/terraform-provider-aws/internal/service/neptune"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNeptuneClusterParameterGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBClusterParameterGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "rds", fmt.Sprintf("cluster-pg:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, names.AttrFamily, "neptune1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccNeptuneClusterParameterGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBClusterParameterGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfneptune.ResourceClusterParameterGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNeptuneClusterParameterGroup_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBClusterParameterGroup
	resourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, id.UniqueIdPrefix),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
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

func TestAccNeptuneClusterParameterGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBClusterParameterGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterParameterGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccClusterParameterGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccNeptuneClusterParameterGroup_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBClusterParameterGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_description(rName, "custom description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "custom description"),
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

func TestAccNeptuneClusterParameterGroup_parameter(t *testing.T) {
	ctx := acctest.Context(t)
	var v neptune.DBClusterParameterGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_one(rName, "neptune_enable_audit_log", acctest.Ct1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"apply_method":  "pending-reboot",
						names.AttrName:  "neptune_enable_audit_log",
						names.AttrValue: acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterParameterGroupConfig_one(rName, "neptune_enable_audit_log", acctest.Ct0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"apply_method":  "pending-reboot",
						names.AttrName:  "neptune_enable_audit_log",
						names.AttrValue: acctest.Ct0,
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NeptuneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterParameterGroupConfig_one(rName, "neptune_enable_audit_log", acctest.Ct0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterParameterGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"apply_method":  "pending-reboot",
						names.AttrName:  "neptune_enable_audit_log",
						names.AttrValue: acctest.Ct0,
					}),
				),
				ExpectNonEmptyPlan: false,
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

func testAccCheckClusterParameterGroupExists(ctx context.Context, n string, v *neptune.DBClusterParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
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

func testAccClusterParameterGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "test" {
  family = "neptune1"
  name   = %[1]q
}
`, rName)
}

func testAccClusterParameterGroupConfig_nameGenerated() string {
	return `
resource "aws_neptune_cluster_parameter_group" "test" {
  family = "neptune1"
}
`
}

func testAccClusterParameterGroupConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "test" {
  family      = "neptune1"
  name_prefix = %[1]q
}
`, namePrefix)
}

func testAccClusterParameterGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "test" {
  family = "neptune1"
  name   = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccClusterParameterGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "test" {
  family = "neptune1"
  name   = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccClusterParameterGroupConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "test" {
  description = %[2]q
  family      = "neptune1"
  name        = %[1]q
}
`, rName, description)
}

func testAccClusterParameterGroupConfig_one(rName, pName, pValue string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "test" {
  family = "neptune1"
  name   = %[1]q

  parameter {
    name  = %[2]q
    value = %[3]q
  }
}
`, rName, pName, pValue)
}
