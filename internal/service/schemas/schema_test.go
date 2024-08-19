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

const (
	testAccSchemaContent = `
{
  "openapi": "3.0.0",
  "info": {
    "version": "1.0.0",
    "title": "Event"
  },
  "paths": {},
  "components": {
    "schemas": {
      "Event": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string"
          }
        }
      }
    }
  }
}
`

	testAccSchemaContentUpdated = `
{
  "openapi": "3.0.0",
  "info": {
    "version": "2.0.0",
    "title": "Event"
  },
  "paths": {},
  "components": {
    "schemas": {
      "Event": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string"
          },
          "created_at": {
            "type": "string",
            "format": "date-time"
          }
        }
      }
    }
  }
}
`

	testAccJSONSchemaContent = `
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.com/product.schema.json",
  "title": "Event",
  "description": "An generic example",
  "type": "object",
  "properties": {
    "name": {
      "description": "The unique identifier for a product",
      "type": "string"
    },
    "created_at": {
      "description": "Date-time format",
      "type": "string",
      "format": "date-time"
    }
  },
  "required": [ "name" ]
}
`
)

func TestAccSchemasSchema_openAPI3(t *testing.T) {
	ctx := acctest.Context(t)
	var v schemas.DescribeSchemaOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_schemas_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.SchemasEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SchemasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "schemas", fmt.Sprintf("schema/%s/%s", rName, rName)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrContent),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "registry_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "OpenApi3"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "version_created_date"),
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

func TestAccSchemasSchema_jsonSchemaDraftv4(t *testing.T) {
	ctx := acctest.Context(t)
	var v schemas.DescribeSchemaOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_schemas_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.SchemasEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SchemasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_jsonSchema(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "schemas", fmt.Sprintf("schema/%s/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrContent, testAccJSONSchemaContent),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "registry_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "JSONSchemaDraft4"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "version_created_date"),
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

func TestAccSchemasSchema_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v schemas.DescribeSchemaOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_schemas_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.SchemasEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SchemasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfschemas.ResourceSchema(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSchemasSchema_contentDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var v schemas.DescribeSchemaOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_schemas_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.SchemasEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SchemasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_contentDescription(rName, testAccSchemaContent, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrContent, testAccSchemaContent),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSchemaConfig_contentDescription(rName, testAccSchemaContentUpdated, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrContent, testAccSchemaContentUpdated),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
				),
			},
			{
				Config: testAccSchemaConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct3),
				),
			},
		},
	})
}

func TestAccSchemasSchema_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v schemas.DescribeSchemaOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_schemas_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.SchemasEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SchemasServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, resourceName, &v),
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
				Config: testAccSchemaConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccSchemaConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckSchemaDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SchemasClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_schemas_schema" {
				continue
			}

			_, err := tfschemas.FindSchemaByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes["registry_name"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EventBridge Schemas Schema %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSchemaExists(ctx context.Context, n string, v *schemas.DescribeSchemaOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SchemasClient(ctx)

		output, err := tfschemas.FindSchemaByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes["registry_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSchemaConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q
}

resource "aws_schemas_schema" "test" {
  name          = %[1]q
  registry_name = aws_schemas_registry.test.name
  type          = "OpenApi3"
  content       = %[2]q
}
`, rName, testAccSchemaContent)
}

func testAccSchemaConfig_jsonSchema(rName string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q
}

resource "aws_schemas_schema" "test" {
  name          = %[1]q
  registry_name = aws_schemas_registry.test.name
  type          = "JSONSchemaDraft4"
  content       = %[2]q
}
`, rName, testAccJSONSchemaContent)
}

func testAccSchemaConfig_contentDescription(rName, content, description string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q
}

resource "aws_schemas_schema" "test" {
  name          = %[1]q
  registry_name = aws_schemas_registry.test.name
  type          = "OpenApi3"
  content       = %[2]q
  description   = %[3]q
}
`, rName, content, description)
}

func testAccSchemaConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_schemas_registry" "test" {
  name = %[1]q
}

resource "aws_schemas_schema" "test" {
  name          = %[1]q
  registry_name = aws_schemas_registry.test.name
  type          = "OpenApi3"
  content       = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, testAccSchemaContent, tagKey1, tagValue1)
}

func testAccSchemaConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`	
resource "aws_schemas_registry" "test" {
  name = %[1]q
}

resource "aws_schemas_schema" "test" {
  name          = %[1]q
  registry_name = aws_schemas_registry.test.name
  type          = "OpenApi3"
  content       = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, testAccSchemaContent, tagKey1, tagValue1, tagKey2, tagValue2)
}
