// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueDataQualityRuleset_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ruleset := "Rules = [Completeness \"colA\" between 0.4 and 0.8]"
	resourceName := "aws_glue_data_quality_ruleset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityRulesetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityRulesetConfig_basic(rName, ruleset),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataQualityRulesetExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("dataQualityRuleset/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "created_on"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_on"),
					resource.TestCheckResourceAttr(resourceName, "ruleset", ruleset),
					resource.TestCheckResourceAttr(resourceName, "target_table.#", acctest.Ct0),
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

func TestAccGlueDataQualityRuleset_updateRuleset(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	originalRuleset := "Rules = [Completeness \"colA\" between 0.4 and 0.8]"
	updatedRuleset := "Rules = [Completeness \"colA\" between 0.5 and 1.0]"
	resourceName := "aws_glue_data_quality_ruleset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityRulesetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityRulesetConfig_basic(rName, originalRuleset),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataQualityRulesetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ruleset", originalRuleset),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataQualityRulesetConfig_basic(rName, updatedRuleset),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataQualityRulesetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "ruleset", updatedRuleset),
				),
			},
		},
	})
}

func TestAccGlueDataQualityRuleset_updateDescription(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ruleset := "Rules = [Completeness \"colA\" between 0.4 and 0.8]"
	originalDescription := "original description"
	updatedDescription := "updated description"
	resourceName := "aws_glue_data_quality_ruleset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityRulesetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityRulesetConfig_description(rName, ruleset, originalDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataQualityRulesetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, originalDescription),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataQualityRulesetConfig_description(rName, ruleset, updatedDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataQualityRulesetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, updatedDescription),
				),
			},
		},
	})
}

func TestAccGlueDataQualityRuleset_targetTableRequired(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ruleset := "Rules = [Completeness \"colA\" between 0.4 and 0.8]"
	resourceName := "aws_glue_data_quality_ruleset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityRulesetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccDataQualityRulesetConfig_targetTable(rName, rName2, rName3, ruleset),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityRulesetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_table.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target_table.0.catalog_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "target_table.0.database_name", "aws_glue_catalog_database.test", names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "target_table.0.table_name", "aws_glue_catalog_table.test", names.AttrName),
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

func TestAccGlueDataQualityRuleset_targetTableFull(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ruleset := "Rules = [Completeness \"colA\" between 0.4 and 0.8]"
	resourceName := "aws_glue_data_quality_ruleset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityRulesetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccDataQualityRulesetConfig_targetTableFull(rName, rName2, rName3, ruleset),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityRulesetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_table.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "target_table.0.catalog_id", "aws_glue_catalog_table.test", names.AttrCatalogID),
					resource.TestCheckResourceAttrPair(resourceName, "target_table.0.database_name", "aws_glue_catalog_database.test", names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "target_table.0.table_name", "aws_glue_catalog_table.test", names.AttrName),
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

func TestAccGlueDataQualityRuleset_tags(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ruleset := "Rules = [Completeness \"colA\" between 0.4 and 0.8]"
	resourceName := "aws_glue_data_quality_ruleset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityRulesetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:  testAccDataQualityRulesetConfig_tags1(rName, ruleset, acctest.CtKey1, acctest.CtValue1),
				Destroy: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataQualityRulesetExists(ctx, resourceName),
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
				Config:  testAccDataQualityRulesetConfig_tags2(rName, ruleset, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Destroy: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataQualityRulesetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config:  testAccDataQualityRulesetConfig_tags1(rName, ruleset, acctest.CtKey2, acctest.CtValue2),
				Destroy: false,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataQualityRulesetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGlueDataQualityRuleset_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ruleset := "Rules = [Completeness \"colA\" between 0.4 and 0.8]"
	resourceName := "aws_glue_data_quality_ruleset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataQualityRulesetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataQualityRulesetConfig_basic(rName, ruleset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataQualityRulesetExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourceDataQualityRuleset(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDataQualityRulesetExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		resp, err := tfglue.FindDataQualityRulesetByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("No Glue Data Quality Ruleset Found")
		}

		if aws.ToString(resp.Name) != rs.Primary.ID {
			return fmt.Errorf("Glue Data Quality Ruleset Mismatch - existing: %q, state: %q",
				aws.ToString(resp.Name), rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDataQualityRulesetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_data_quality_ruleset" {
				continue
			}

			_, err := tfglue.FindDataQualityRulesetByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}
			return fmt.Errorf("Glue Data Quality Ruleset %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDataQualityRulesetConfig_basic(rName, ruleset string) string {
	return fmt.Sprintf(`
resource "aws_glue_data_quality_ruleset" "test" {
  name    = %[1]q
  ruleset = %[2]q
}
`, rName, ruleset)
}

func testAccDataQualityRulesetConfig_description(rName, ruleset, description string) string {
	return fmt.Sprintf(`
resource "aws_glue_data_quality_ruleset" "test" {
  name        = %[1]q
  ruleset     = %[2]q
  description = %[3]q
}
`, rName, ruleset, description)
}

func testAccDataQualityRulesetConfigTargetTableConfigBasic(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[2]q
  database_name = aws_glue_catalog_database.test.name
}
`, rName, rName2)
}

func testAccDataQualityRulesetConfig_targetTable(rName, rName2, rName3, ruleset string) string {
	return acctest.ConfigCompose(
		testAccDataQualityRulesetConfigTargetTableConfigBasic(rName2, rName3),
		fmt.Sprintf(`
resource "aws_glue_data_quality_ruleset" "test" {
  name    = %[1]q
  ruleset = %[2]q

  target_table {
    database_name = aws_glue_catalog_database.test.name
    table_name    = aws_glue_catalog_table.test.name
  }
}
`, rName, ruleset))
}

func testAccDataQualityRulesetConfig_targetTableFull(rName, rName2, rName3, ruleset string) string {
	return acctest.ConfigCompose(
		testAccDataQualityRulesetConfigTargetTableConfigBasic(rName2, rName3),
		fmt.Sprintf(`
resource "aws_glue_data_quality_ruleset" "test" {
  name    = %[1]q
  ruleset = %[2]q

  target_table {
    catalog_id    = aws_glue_catalog_table.test.catalog_id
    database_name = aws_glue_catalog_database.test.name
    table_name    = aws_glue_catalog_table.test.name
  }
}
`, rName, ruleset))
}

func testAccDataQualityRulesetConfig_tags1(rName, ruleset, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_glue_data_quality_ruleset" "test" {
  name    = %[1]q
  ruleset = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, ruleset, tagKey1, tagValue1)
}

func testAccDataQualityRulesetConfig_tags2(rName, ruleset, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_glue_data_quality_ruleset" "test" {
  name    = %[1]q
  ruleset = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, ruleset, tagKey1, tagValue1, tagKey2, tagValue2)
}
