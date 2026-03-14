// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueRegistry_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var registry glue.GetRegistryOutput

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckRegistry(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName, &registry),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "glue", fmt.Sprintf("registry/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "registry_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccGlueRegistry_description(t *testing.T) {
	ctx := acctest.Context(t)
	var registry glue.GetRegistryOutput

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckRegistry(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig_description(rName, "First Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName, &registry),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "First Description"),
				),
			},
			{
				Config: testAccRegistryConfig_description(rName, "Second Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName, &registry),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Second Description"),
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

func TestAccGlueRegistry_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var registry glue.GetRegistryOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckRegistry(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName, &registry),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
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
					testAccCheckRegistryExists(ctx, t, resourceName, &registry),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRegistryConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName, &registry),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGlueRegistry_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var registry glue.GetRegistryOutput

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_registry.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckRegistry(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRegistryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRegistryExists(ctx, t, resourceName, &registry),
					acctest.CheckSDKResourceDisappears(ctx, t, tfglue.ResourceRegistry(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPreCheckRegistry(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).GlueClient(ctx)

	_, err := conn.ListRegistries(ctx, &glue.ListRegistriesInput{})

	// Some endpoints that do not support Glue Registrys return InternalFailure
	if acctest.PreCheckSkipError(err) || tfawserr.ErrCodeEquals(err, "InternalFailure") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckRegistryExists(ctx context.Context, t *testing.T, resourceName string, registry *glue.GetRegistryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Registry ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).GlueClient(ctx)
		output, err := tfglue.FindRegistryByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Glue Registry (%s) not found", rs.Primary.ID)
		}

		if aws.ToString(output.RegistryArn) == rs.Primary.ID {
			*registry = *output
			return nil
		}

		return fmt.Errorf("Glue Registry (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckRegistryDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_registry" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).GlueClient(ctx)
			output, err := tfglue.FindRegistryByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if output != nil && aws.ToString(output.RegistryArn) == rs.Primary.ID {
				return fmt.Errorf("Glue Registry %s still exists", rs.Primary.ID)
			}

			return err
		}

		return nil
	}
}

func testAccRegistryConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_glue_registry" "test" {
  registry_name = %[1]q
  description   = %[2]q
}
`, rName, description)
}

func testAccRegistryConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_registry" "test" {
  registry_name = %[1]q
}
`, rName)
}

func testAccRegistryConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_glue_registry" "test" {
  registry_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRegistryConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_glue_registry" "test" {
  registry_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
