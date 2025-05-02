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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfaccessanalyzer "github.com/hashicorp/terraform-provider-aws/internal/service/accessanalyzer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAnalyzer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var analyzer types.AnalyzerSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccessAnalyzerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalyzerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalyzerConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(ctx, resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, "analyzer_name", rName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "access-analyzer", fmt.Sprintf("analyzer/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.TypeAccount)),
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

func testAccAnalyzer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var analyzer types.AnalyzerSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccessAnalyzerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalyzerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalyzerConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(ctx, resourceName, &analyzer),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfaccessanalyzer.ResourceAnalyzer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAnalyzer_typeOrganization(t *testing.T) {
	ctx := acctest.Context(t)
	var analyzer types.AnalyzerSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccessAnalyzerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalyzerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalyzerConfig_typeOrganization(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(ctx, resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.TypeOrganization)),
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

func testAccAnalyzer_configuration(t *testing.T) {
	ctx := acctest.Context(t)
	var analyzer types.AnalyzerSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AccessAnalyzerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalyzerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalyzerConfig_configuration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(ctx, resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.unused_access.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.unused_access.0.unused_access_age", "180"),
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

func testAccAnalyzer_organizationUnusedAccess(t *testing.T) {
	ctx := acctest.Context(t)
	var analyzer types.AnalyzerSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccessAnalyzerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnalyzerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAnalyzerConfig_organizationUnusedAccess(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(ctx, resourceName, &analyzer),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.unused_access.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.unused_access.0.unused_access_age", "180"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.unused_access.0.analysis_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.unused_access.0.analysis_rule.0.exclusion.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.unused_access.0.analysis_rule.0.exclusion.0.account_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.unused_access.0.analysis_rule.0.exclusion.0.account_ids.0", acctest.Ct12Digit),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.unused_access.0.analysis_rule.0.exclusion.0.account_ids.1", "234567890123"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.unused_access.0.analysis_rule.0.exclusion.1.resource_tags.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.unused_access.0.analysis_rule.0.exclusion.1.resource_tags.0.key1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.unused_access.0.analysis_rule.0.exclusion.1.resource_tags.1.key2", acctest.CtValue2),
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

func testAccAnalyzer_upgradeV5_95_0(t *testing.T) {
	ctx := acctest.Context(t)
	var analyzer types.AnalyzerSummary
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_accessanalyzer_analyzer.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.AccessAnalyzerServiceID),
		CheckDestroy: testAccCheckAnalyzerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.95.0",
					},
				},
				Config: testAccAnalyzerConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(ctx, resourceName, &analyzer),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccAnalyzerConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnalyzerExists(ctx, resourceName, &analyzer),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func testAccCheckAnalyzerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AccessAnalyzerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_accessanalyzer_analyzer" {
				continue
			}

			_, err := tfaccessanalyzer.FindAnalyzerByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM Access Analyzer Analyzer %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAnalyzerExists(ctx context.Context, n string, v *types.AnalyzerSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AccessAnalyzerClient(ctx)

		output, err := tfaccessanalyzer.FindAnalyzerByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAnalyzerConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = %[1]q
}
`, rName)
}

func testAccAnalyzerConfig_typeOrganization(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["access-analyzer.${data.aws_partition.current.dns_suffix}"]
}

resource "aws_accessanalyzer_analyzer" "test" {
  depends_on = [aws_organizations_organization.test]

  analyzer_name = %[1]q
  type          = "ORGANIZATION"
}
`, rName)
}

func testAccAnalyzerConfig_configuration(rName string) string {
	return fmt.Sprintf(`
resource "aws_accessanalyzer_analyzer" "test" {
  analyzer_name = %[1]q
  type          = "ACCOUNT_UNUSED_ACCESS"

  configuration {
    unused_access {
      unused_access_age = 180
    }
  }
}
`, rName)
}

func testAccAnalyzerConfig_organizationUnusedAccess(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["access-analyzer.${data.aws_partition.current.dns_suffix}"]
}

resource "aws_accessanalyzer_analyzer" "test" {
  depends_on = [aws_organizations_organization.test]

  analyzer_name = %[1]q
  type          = "ORGANIZATION_UNUSED_ACCESS"

  configuration {
    unused_access {
      unused_access_age = 180
      analysis_rule {
        exclusion {
          account_ids = [
            "123456789012",
            "234567890123",
          ]
        }
        exclusion {
          resource_tags = [
            { key1 = "value1" },
            { key2 = "value2" },
          ]
        }
      }
    }
  }
}
`, rName)
}
