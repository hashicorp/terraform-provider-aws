// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFV2APIKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var apiKey awstypes.APIKeySummary
	resourceName := "aws_wafv2_api_key.test"
	domain := []string{acctest.RandomDomainName()}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_basic(domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, t, resourceName, &apiKey),
					resource.TestCheckResourceAttrSet(resourceName, "api_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, "token_domains.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccAPIKeyImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "api_key",
			},
		},
	})
}

func TestAccWAFV2APIKey_multipleTokenDomains(t *testing.T) {
	ctx := acctest.Context(t)
	var apiKey awstypes.APIKeySummary
	resourceName := "aws_wafv2_api_key.test"
	var domains []string
	for range 4 {
		domains = append(domains, acctest.RandomDomainName())
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_basic(domains),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, t, resourceName, &apiKey),
					resource.TestCheckResourceAttrSet(resourceName, "api_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, "token_domains.#", "4"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccAPIKeyImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "api_key",
			},
		},
	})
}

func TestAccWAFV2APIKey_changeTokenDomainsForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var apiKey awstypes.APIKeySummary
	resourceName := "aws_wafv2_api_key.test"
	domain := []string{acctest.RandomDomainName()}
	domainNew := []string{acctest.RandomDomainName()}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_basic(domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, t, resourceName, &apiKey),
					resource.TestCheckResourceAttrSet(resourceName, "api_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, "token_domains.#", "1"),
				),
			},
			{
				Config: testAccAPIKeyConfig_basic(domainNew),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, t, resourceName, &apiKey),
					resource.TestCheckResourceAttrSet(resourceName, "api_key"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, "token_domains.#", "1"),
				),
			},
		},
	})
}

func TestAccWAFV2APIKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var apiKey awstypes.APIKeySummary
	resourceName := "aws_wafv2_api_key.test"
	domain := []string{acctest.RandomDomainName()}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeyConfig_basic(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAPIKeyExists(ctx, t, resourceName, &apiKey),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfwafv2.ResourceAPIKey, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckAPIKeyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WAFV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafv2_api_key" {
				continue
			}

			_, err := tfwafv2.FindAPIKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_key"], awstypes.Scope(rs.Primary.Attributes[names.AttrScope]))

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAFv2 API Key %s still exists", rs.Primary.Attributes["api_key"])
		}

		return nil
	}
}

func testAccCheckAPIKeyExists(ctx context.Context, t *testing.T, n string, v *awstypes.APIKeySummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WAFV2Client(ctx)

		output, err := tfwafv2.FindAPIKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes["api_key"], awstypes.Scope(rs.Primary.Attributes[names.AttrScope]))

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAPIKeyImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["api_key"], rs.Primary.Attributes[names.AttrScope]), nil
	}
}

func testAccAPIKeyConfig_basic(domains []string) string {
	d, _ := json.Marshal(domains)
	return fmt.Sprintf(`
resource "aws_wafv2_api_key" "test" {
  scope         = "REGIONAL"
  token_domains = %[1]s
}
`, d)
}
