// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package athena_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfathena "github.com/hashicorp/terraform-provider-aws/internal/service/athena"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAthenaDataCatalog_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_athena_data_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCatalogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "athena", fmt.Sprintf("datacatalog/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "A test data catalog"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.function", "arn:aws:lambda:us-east-1:123456789012:function:test-function"), //lintignore:AWSAT003,AWSAT005
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameters},
			},
		},
	})
}

func TestAccAthenaDataCatalog_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_athena_data_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCatalogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfathena.ResourceDataCatalog(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAthenaDataCatalog_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_athena_data_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCatalogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameters},
			},
			{
				Config: testAccDataCatalogConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDataCatalogConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccAthenaDataCatalog_type_lambda(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_athena_data_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCatalogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogConfig_typeLambda(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "A test data catalog using Lambda"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "LAMBDA"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "parameters.metadata-function", "arn:aws:lambda:us-east-1:123456789012:function:test-function"), //lintignore:AWSAT003,AWSAT005
					resource.TestCheckResourceAttr(resourceName, "parameters.record-function", "arn:aws:lambda:us-east-1:123456789012:function:test-function"),   //lintignore:AWSAT003,AWSAT005
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameters},
			},
		},
	})
}

func TestAccAthenaDataCatalog_type_hive(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_athena_data_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCatalogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogConfig_typeHive(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "A test data catalog using Hive"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "HIVE"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.metadata-function", "arn:aws:lambda:us-east-1:123456789012:function:test-function"), //lintignore:AWSAT003,AWSAT005
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameters},
			},
		},
	})
}

func TestAccAthenaDataCatalog_type_glue(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_athena_data_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCatalogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogConfig_typeGlue(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "A test data catalog using Glue"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "GLUE"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.catalog-id", "123456789012"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameters},
			},
		},
	})
}

func TestAccAthenaDataCatalog_parameters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_athena_data_catalog.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AthenaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataCatalogDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogConfig_parameters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.function", "arn:aws:lambda:us-east-1:123456789012:function:test-function-1"), //lintignore:AWSAT003,AWSAT005
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameters},
			},
			{
				Config: testAccDataCatalogConfig_parametersUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "parameters.metadata-function", "arn:aws:lambda:us-east-1:123456789012:function:test-function-2"), //lintignore:AWSAT003,AWSAT005
					resource.TestCheckResourceAttr(resourceName, "parameters.record-function", "arn:aws:lambda:us-east-1:123456789012:function:test-function-2"),   //lintignore:AWSAT003,AWSAT005
				),
			},
		},
	})
}

func testAccCheckDataCatalogExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaClient(ctx)

		_, err := tfathena.FindDataCatalogByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckDataCatalogDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AthenaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_athena_data_catalog" {
				continue
			}

			_, err := tfathena.FindDataCatalogByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Athena Data Catalog %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDataCatalogConfig_basic(rName string) string {
	//lintignore:AWSAT003,AWSAT005
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
  name        = %[1]q
  description = "A test data catalog"
  type        = "LAMBDA"

  parameters = {
    "function" = "arn:aws:lambda:us-east-1:123456789012:function:test-function"
  }
}
`, rName)
}

func testAccDataCatalogConfig_tags1(rName, tagKey1, tagValue1 string) string {
	//lintignore:AWSAT003,AWSAT005
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
  name = %[1]q
  type = "LAMBDA"

  description = "Testing tags"

  parameters = {
    "function" = "arn:aws:lambda:us-east-1:123456789012:function:test-function"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDataCatalogConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	//lintignore:AWSAT003,AWSAT005
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
  name = %[1]q
  type = "LAMBDA"

  description = "Testing tags"

  parameters = {
    "function" = "arn:aws:lambda:us-east-1:123456789012:function:test-function"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDataCatalogConfig_typeLambda(rName string) string {
	//lintignore:AWSAT003,AWSAT005
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
  name        = %[1]q
  description = "A test data catalog using Lambda"
  type        = "LAMBDA"

  parameters = {
    "metadata-function" = "arn:aws:lambda:us-east-1:123456789012:function:test-function"
    "record-function"   = "arn:aws:lambda:us-east-1:123456789012:function:test-function"
  }
}
`, rName)
}

func testAccDataCatalogConfig_typeHive(rName string) string {
	//lintignore:AWSAT003,AWSAT005
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
  name        = %[1]q
  description = "A test data catalog using Hive"
  type        = "HIVE"

  parameters = {
    "metadata-function" = "arn:aws:lambda:us-east-1:123456789012:function:test-function"
  }
}
`, rName)
}

func testAccDataCatalogConfig_typeGlue(rName string) string {
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
  name        = %[1]q
  description = "A test data catalog using Glue"
  type        = "GLUE"

  parameters = {
    "catalog-id" = "123456789012"
  }
}
`, rName)
}

func testAccDataCatalogConfig_parameters(rName string) string {
	//lintignore:AWSAT003,AWSAT005
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
  name        = %[1]q
  description = "Testing parameters attribute"
  type        = "LAMBDA"

  parameters = {
    "function" = "arn:aws:lambda:us-east-1:123456789012:function:test-function-1"
  }
}
`, rName)
}

func testAccDataCatalogConfig_parametersUpdated(rName string) string {
	//lintignore:AWSAT003,AWSAT005
	return fmt.Sprintf(`
resource "aws_athena_data_catalog" "test" {
  name        = %[1]q
  description = "Testing parameters attribute"
  type        = "LAMBDA"

  parameters = {
    "metadata-function" = "arn:aws:lambda:us-east-1:123456789012:function:test-function-2"
    "record-function"   = "arn:aws:lambda:us-east-1:123456789012:function:test-function-2"
  }
}
`, rName)
}
