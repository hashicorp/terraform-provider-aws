// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package accessanalyzer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfaccessanalyzer "github.com/hashicorp/terraform-provider-aws/internal/service/accessanalyzer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAnalyzerArchiveRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var archiveRule types.ArchiveRuleSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_archive_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AccessAnalyzerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccessAnalyzerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckArchiveRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveRuleExists(ctx, resourceName, &archiveRule),
					resource.TestCheckResourceAttr(resourceName, "filter.0.criteria", "isPublic"),
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

func testAccAnalyzerArchiveRule_updateFilters(t *testing.T) {
	ctx := acctest.Context(t)
	var archiveRule types.ArchiveRuleSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_archive_rule.test"

	filters := `
filter {
  criteria = "error"
  exists   = true
}
`

	filtersUpdated := `
filter {
  criteria = "error"
  exists   = true
}

filter {
  criteria = "isPublic"
  eq       = ["false"]
}
`

	filtersRemoved := `
filter {
  criteria = "isPublic"
  eq       = ["true"]
}
`
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AccessAnalyzerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccessAnalyzerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckArchiveRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveRuleConfig_updateFilters(rName, filters),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveRuleExists(ctx, resourceName, &archiveRule),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.criteria", "error"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.exists", acctest.CtTrue),
				),
			},
			{
				Config: testAccArchiveRuleConfig_updateFilters(rName, filtersUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveRuleExists(ctx, resourceName, &archiveRule),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "filter.0.criteria", "error"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.exists", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "filter.1.criteria", "isPublic"),
					resource.TestCheckResourceAttr(resourceName, "filter.1.eq.0", acctest.CtFalse),
				),
			},
			{
				Config: testAccArchiveRuleConfig_updateFilters(rName, filtersRemoved),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveRuleExists(ctx, resourceName, &archiveRule),
					resource.TestCheckResourceAttr(resourceName, "filter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "filter.0.criteria", "isPublic"),
					resource.TestCheckResourceAttr(resourceName, "filter.0.eq.0", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccAnalyzerArchiveRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var archiveRule types.ArchiveRuleSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_archive_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AccessAnalyzerEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccessAnalyzerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckArchiveRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccArchiveRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckArchiveRuleExists(ctx, resourceName, &archiveRule),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfaccessanalyzer.ResourceArchiveRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckArchiveRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AccessAnalyzerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_accessanalyzer_archive_rule" {
				continue
			}

			analyzerName, ruleName, err := tfaccessanalyzer.ArchiveRuleParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfaccessanalyzer.FindArchiveRuleByTwoPartKey(ctx, conn, analyzerName, ruleName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM Access Analyzer Archive Rule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckArchiveRuleExists(ctx context.Context, n string, v *types.ArchiveRuleSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IAM Access Analyzer Archive Rule ID is set")
		}

		analyzerName, ruleName, err := tfaccessanalyzer.ArchiveRuleParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AccessAnalyzerClient(ctx)

		output, err := tfaccessanalyzer.FindArchiveRuleByTwoPartKey(ctx, conn, analyzerName, ruleName)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccArchiveRuleBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = %[1]q
}

`, rName)
}

func testAccArchiveRuleConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccArchiveRuleBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_accessanalyzer_archive_rule" "test" {
  analyzer_name = aws_accessanalyzer_analyzer.test.analyzer_name
  rule_name     = %[1]q

  filter {
    criteria = "isPublic"
    eq       = ["false"]
  }
}
`, rName))
}

func testAccArchiveRuleConfig_updateFilters(rName, filters string) string {
	return acctest.ConfigCompose(
		testAccArchiveRuleBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_accessanalyzer_archive_rule" "test" {
  analyzer_name = aws_accessanalyzer_analyzer.test.analyzer_name
  rule_name     = %[1]q

  %[2]s
}
`, rName, filters))
}
