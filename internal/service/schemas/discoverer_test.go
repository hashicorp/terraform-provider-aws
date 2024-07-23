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

func TestAccSchemasDiscoverer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v schemas.DescribeDiscovererOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_schemas_discoverer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.SchemasEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SchemasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDiscovererDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDiscovererConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiscovererExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "schemas", fmt.Sprintf("discoverer/events-event-bus-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
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

func TestAccSchemasDiscoverer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v schemas.DescribeDiscovererOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_schemas_discoverer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.SchemasEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SchemasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDiscovererDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDiscovererConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiscovererExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfschemas.ResourceDiscoverer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSchemasDiscoverer_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v schemas.DescribeDiscovererOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_schemas_discoverer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.SchemasEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SchemasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDiscovererDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDiscovererConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiscovererExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDiscovererConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiscovererExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
			{
				Config: testAccDiscovererConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiscovererExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
		},
	})
}

func TestAccSchemasDiscoverer_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v schemas.DescribeDiscovererOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_schemas_discoverer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.SchemasEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SchemasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDiscovererDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDiscovererConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiscovererExists(ctx, resourceName, &v),
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
				Config: testAccDiscovererConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiscovererExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDiscovererConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiscovererExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckDiscovererDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SchemasClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_schemas_discoverer" {
				continue
			}

			_, err := tfschemas.FindDiscovererByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EventBridge Schemas Discoverer %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDiscovererExists(ctx context.Context, n string, v *schemas.DescribeDiscovererOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SchemasClient(ctx)

		output, err := tfschemas.FindDiscovererByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDiscovererConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_schemas_discoverer" "test" {
  source_arn = aws_cloudwatch_event_bus.test.arn
}
`, rName)
}

func testAccDiscovererConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_schemas_discoverer" "test" {
  source_arn = aws_cloudwatch_event_bus.test.arn

  description = %[2]q
}
`, rName, description)
}

func testAccDiscovererConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_schemas_discoverer" "test" {
  source_arn = aws_cloudwatch_event_bus.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDiscovererConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_schemas_discoverer" "test" {
  source_arn = aws_cloudwatch_event_bus.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
