// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueUserDefinedFunction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_user_defined_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUDFDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserDefinedFunctionConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserDefinedFunctionExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("userDefinedFunction/%s/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "class_name", rName),
					resource.TestCheckResourceAttr(resourceName, "owner_name", rName),
					resource.TestCheckResourceAttr(resourceName, "owner_type", "GROUP"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserDefinedFunctionConfig_basic(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserDefinedFunctionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "class_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "owner_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "owner_type", "GROUP"),
				),
			},
		},
	})
}

func TestAccGlueUserDefinedFunction_Resource_uri(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_user_defined_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUDFDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserDefinedFunctionConfig_resourceURI1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserDefinedFunctionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_uris.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserDefinedFunctionConfig_resourceURI2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserDefinedFunctionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_uris.#", acctest.Ct2),
				),
			},
			{
				Config: testAccUserDefinedFunctionConfig_resourceURI1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserDefinedFunctionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "resource_uris.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccGlueUserDefinedFunction_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_user_defined_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUDFDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserDefinedFunctionConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserDefinedFunctionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourceUserDefinedFunction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckUDFDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_user_defined_function" {
				continue
			}

			catalogId, dbName, funcName, err := tfglue.ReadUDFID(rs.Primary.ID)
			if err != nil {
				return err
			}

			input := &glue.GetUserDefinedFunctionInput{
				CatalogId:    aws.String(catalogId),
				DatabaseName: aws.String(dbName),
				FunctionName: aws.String(funcName),
			}
			if _, err := conn.GetUserDefinedFunction(ctx, input); err != nil {
				//Verify the error is what we want
				if errs.IsA[*awstypes.EntityNotFoundException](err) {
					continue
				}

				return err
			}
			return fmt.Errorf("still exists")
		}
		return nil
	}
}

func testAccCheckUserDefinedFunctionExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		catalogId, dbName, funcName, err := tfglue.ReadUDFID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)
		out, err := conn.GetUserDefinedFunction(ctx, &glue.GetUserDefinedFunctionInput{
			CatalogId:    aws.String(catalogId),
			DatabaseName: aws.String(dbName),
			FunctionName: aws.String(funcName),
		})

		if err != nil {
			return err
		}

		if out.UserDefinedFunction == nil {
			return fmt.Errorf("No Glue User Defined Function Found")
		}

		if *out.UserDefinedFunction.FunctionName != funcName {
			return fmt.Errorf("Glue UDF Mismatch - existing: %q, state: %q",
				*out.UserDefinedFunction.FunctionName, funcName)
		}

		return nil
	}
}

func testAccUserDefinedFunctionConfig_basic(rName string, name string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_user_defined_function" "test" {
  name          = %[1]q
  catalog_id    = aws_glue_catalog_database.test.catalog_id
  database_name = aws_glue_catalog_database.test.name
  class_name    = %[2]q
  owner_name    = %[2]q
  owner_type    = "GROUP"
}
`, rName, name)
}

func testAccUserDefinedFunctionConfig_resourceURI1(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_user_defined_function" "test" {
  name          = %[1]q
  catalog_id    = aws_glue_catalog_database.test.catalog_id
  database_name = aws_glue_catalog_database.test.name
  class_name    = %[1]q
  owner_name    = %[1]q
  owner_type    = "GROUP"

  resource_uris {
    resource_type = "ARCHIVE"
    uri           = %[1]q
  }
}
`, rName)
}

func testAccUserDefinedFunctionConfig_resourceURI2(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_user_defined_function" "test" {
  name          = %[1]q
  catalog_id    = aws_glue_catalog_database.test.catalog_id
  database_name = aws_glue_catalog_database.test.name
  class_name    = %[1]q
  owner_name    = %[1]q
  owner_type    = "GROUP"

  resource_uris {
    resource_type = "ARCHIVE"
    uri           = %[1]q
  }

  resource_uris {
    resource_type = "JAR"
    uri           = %[1]q
  }
}
`, rName)
}
