// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemas_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/schemas"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfschemas "github.com/hashicorp/terraform-provider-aws/internal/service/schemas"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSchemasRegistry_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v schemas.DescribeRegistryOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_schemas_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.SchemasEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SchemasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "schemas", fmt.Sprintf("registry/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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

func TestAccSchemasRegistry_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v schemas.DescribeRegistryOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_schemas_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.SchemasEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SchemasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfschemas.ResourceRegistry(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSchemasRegistry_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v schemas.DescribeRegistryOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_schemas_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.SchemasEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SchemasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRegistryConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
			{
				Config: testAccRegistryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
		},
	})
}

func TestAccSchemasRegistry_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v schemas.DescribeRegistryOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_schemas_registry.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.SchemasEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SchemasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryExists(ctx, resourceName, &v),
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
				Config: testAccRegistryConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRegistryConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckRegistryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SchemasClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_schemas_registry" {
				continue
			}

			_, err := tfschemas.FindRegistryByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EventBridge Schemas Registry %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRegistryExists(ctx context.Context, n string, v *schemas.DescribeRegistryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SchemasClient(ctx)

		output, err := tfschemas.FindRegistryByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRegistryConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRegistryConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}

func testAccRegistryConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRegistryConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
