// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRSigningConfiguration_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:           testAccSigningConfiguration_basic,
		acctest.CtDisappears:      testAccSigningConfiguration_disappears,
		"update":                  testAccSigningConfiguration_update,
		"noRepositoryFilter":      testAccSigningConfiguration_noRepositoryFilter,
		"Identity_basic":          testAccECRSigningConfiguration_Identity_basic,
		"Identity_regionOverride": testAccECRSigningConfiguration_Identity_regionOverride,
		"List_basic":              testAccECRSigningConfiguration_List_basic,
		"List_includeResource":    testAccECRSigningConfiguration_List_includeResource,
		"List_regionOverride":     testAccECRSigningConfiguration_List_regionOverride,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccSigningConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_signing_configuration.test"
	profileResourceName := "aws_signer_signing_profile.test"
	rName := testAccSigningProfileName(t)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSigningConfiguration(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSigningConfigurationExists(ctx, t, resourceName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.signing_profile_arn", profileResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"repository_filter.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.repository_filter.*", map[string]string{
						names.AttrFilter: "team-a/*",
						"filter_type":    "WILDCARD_MATCH",
					}),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrRegion),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "registry_id",
			},
		},
	})
}

func testAccSigningConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_signing_configuration.test"
	rName := testAccSigningProfileName(t)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSigningConfiguration(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSigningConfigurationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfecr.ResourceSigningConfiguration, resourceName),
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

func testAccSigningConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_signing_configuration.test"
	rName := testAccSigningProfileName(t)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSigningConfiguration(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningConfigurationConfig_twoProfiles(rName, testAccSigningConfigurationRules_one),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSigningConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.signing_profile_arn", "aws_signer_signing_profile.a", names.AttrARN),
				),
			},
			{
				Config: testAccSigningConfigurationConfig_twoProfiles(rName, testAccSigningConfigurationRules_two),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSigningConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.signing_profile_arn", "aws_signer_signing_profile.a", names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.signing_profile_arn", "aws_signer_signing_profile.b", names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.repository_filter.*", map[string]string{
						names.AttrFilter: "team-a/*",
						"filter_type":    "WILDCARD_MATCH",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.repository_filter.*", map[string]string{
						names.AttrFilter: "*/production",
						"filter_type":    "WILDCARD_MATCH",
					}),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrRegion),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "registry_id",
			},
			// PutSigningConfiguration replaces the whole configuration, so removing a rule must
			// take effect.
			{
				Config: testAccSigningConfigurationConfig_twoProfiles(rName, testAccSigningConfigurationRules_oneOther),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSigningConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "rule.*.signing_profile_arn", "aws_signer_signing_profile.b", names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"repository_filter.#": "1",
					}),
				),
			},
		},
	})
}

func testAccSigningConfiguration_noRepositoryFilter(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_signing_configuration.test"
	rName := testAccSigningProfileName(t)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSigningConfiguration(ctx, t, acctest.Region())
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSigningConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSigningConfigurationConfig_noRepositoryFilter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSigningConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"repository_filter.#": "0",
					}),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrRegion),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "registry_id",
			},
		},
	})
}

// testAccPreCheckSigningConfiguration skips when managed signing is unavailable, and when the
// registry already has a signing configuration: the tests replace and then delete it.
func testAccPreCheckSigningConfiguration(ctx context.Context, t *testing.T, region string) {
	ctx = acctest.NewTestResourceContext(ctx, "aws_ecr_signing_configuration", region)
	conn := acctest.ProviderMeta(ctx, t).ECRClient(ctx)

	_, err := tfecr.FindSigningConfiguration(ctx, conn)
	switch {
	case retry.NotFound(err):
		return
	case acctest.PreCheckSkipError(err):
		t.Skipf("skipping acceptance testing: %s", err)
	case err != nil:
		t.Fatalf("unexpected PreCheck error: %s", err)
	default:
		t.Skip("skipping acceptance testing: registry already has an ECR signing configuration")
	}
}

func testAccCheckSigningConfigurationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).ECRClient(ctx)

		_, err := tfecr.FindSigningConfiguration(ctx, conn)

		return err
	}
}

func testAccCheckSigningConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecr_signing_configuration" {
				continue
			}

			ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).ECRClient(ctx)

			_, err := tfecr.FindSigningConfiguration(ctx, conn)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("ECR Signing Configuration %s still exists", rs.Primary.Attributes["registry_id"])
		}

		return nil
	}
}

// testAccSigningProfileName returns a random Signer signing profile name. Signer profile names
// allow only alphanumerics and underscores, so acctest.RandomWithPrefix cannot be used.
func testAccSigningProfileName(t *testing.T) string {
	return fmt.Sprintf("tf_acc_test_%d", acctest.RandInt(t))
}

func testAccSigningConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test" {
  name        = %[1]q
  platform_id = "Notation-OCI-SHA384-ECDSA"
}

resource "aws_ecr_signing_configuration" "test" {
  rule {
    signing_profile_arn = aws_signer_signing_profile.test.arn

    repository_filter {
      filter      = "team-a/*"
      filter_type = "WILDCARD_MATCH"
    }
  }
}
`, rName)
}

func testAccSigningConfigurationConfig_noRepositoryFilter(rName string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test" {
  name        = %[1]q
  platform_id = "Notation-OCI-SHA384-ECDSA"
}

resource "aws_ecr_signing_configuration" "test" {
  rule {
    signing_profile_arn = aws_signer_signing_profile.test.arn
  }
}
`, rName)
}

const testAccSigningConfigurationRules_one = `
  rule {
    signing_profile_arn = aws_signer_signing_profile.a.arn

    repository_filter {
      filter      = "team-a/*"
      filter_type = "WILDCARD_MATCH"
    }
  }
`

const testAccSigningConfigurationRules_two = `
  rule {
    signing_profile_arn = aws_signer_signing_profile.a.arn

    repository_filter {
      filter      = "team-a/*"
      filter_type = "WILDCARD_MATCH"
    }
  }

  rule {
    signing_profile_arn = aws_signer_signing_profile.b.arn

    repository_filter {
      filter      = "*/production"
      filter_type = "WILDCARD_MATCH"
    }

    repository_filter {
      filter      = "team-b/*"
      filter_type = "WILDCARD_MATCH"
    }
  }
`

const testAccSigningConfigurationRules_oneOther = `
  rule {
    signing_profile_arn = aws_signer_signing_profile.b.arn

    repository_filter {
      filter      = "*/production"
      filter_type = "WILDCARD_MATCH"
    }
  }
`

func testAccSigningConfigurationConfig_twoProfiles(rName, rules string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "a" {
  name        = "%[1]s_a"
  platform_id = "Notation-OCI-SHA384-ECDSA"
}

resource "aws_signer_signing_profile" "b" {
  name        = "%[1]s_b"
  platform_id = "Notation-OCI-SHA384-ECDSA"
}

resource "aws_ecr_signing_configuration" "test" {
%[2]s
}
`, rName, rules)
}
