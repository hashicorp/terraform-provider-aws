// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRRepository_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Repository
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ecr", fmt.Sprintf("repository/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					testAccCheckRepositoryRegistryID(ctx, resourceName),
					testAccCheckRepositoryRepositoryURL(ctx, resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", string(types.EncryptionTypeAes256)),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.kms_key", ""),
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

func TestAccECRRepository_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Repository
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfecr.ResourceRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECRRepository_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 types.Repository
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, t, resourceName, &v1),
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
				Config: testAccRepositoryConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRepositoryConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccECRRepository_immutability(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Repository
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_immutability(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability", "IMMUTABLE"),
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

func TestAccECRRepository_immutabilityWithExclusion(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Repository
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_immutabilityWithExclusion(rName, "latest*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability", string(types.ImageTagMutabilityImmutableWithExclusion)),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.0.filter", "latest*"),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.0.filter_type", string(types.ImageTagMutabilityExclusionFilterTypeWildcard)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryConfig_immutabilityWithExclusion(rName, "dev-*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.0.filter", "dev-*"),
				),
			},
		},
	})
}

func TestAccECRRepository_mutabilityWithExclusion(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.Repository
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_mutabilityWithExclusion(rName, "prod-*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability", string(types.ImageTagMutabilityMutableWithExclusion)),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.0.filter", "prod-*"),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.0.filter_type", string(types.ImageTagMutabilityExclusionFilterTypeWildcard)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryConfig_mutabilityWithExclusion(rName, "release-*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.0.filter", "release-*"),
				),
			},
		},
	})
}

func TestAccECRRepository_immutabilityWithExclusion_validation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccRepositoryConfig_immutabilityWithExclusion(rName, "invalid!@#$"),
				ExpectError: regexache.MustCompile(`must contain only letters, numbers, and special characters`),
			},
			{
				Config:      testAccRepositoryConfig_immutabilityWithExclusion(rName, "a*b*c*d"),
				ExpectError: regexache.MustCompile(`Image tag mutability exclusion filter can contain a maximum of 2 wildcards`),
			},
			{
				Config:      testAccRepositoryConfig_immutabilityWithExclusion(rName, strings.Repeat("a", 129)),
				ExpectError: regexache.MustCompile(`expected length of.*to be in the range.*128`),
			},
		},
	})
}

func TestAccECRRepository_immutabilityWithExclusion_crossValidation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccRepositoryConfig_immutabilityWithExclusionInvalid(rName),
				ExpectError: regexache.MustCompile(`image_tag_mutability_exclusion_filter can only be used when image_tag_mutability is set to IMMUTABLE_WITH_EXCLUSION`),
			},
		},
	})
}

func TestAccECRRepository_Image_scanning(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 types.Repository
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_imageScanningConfiguration(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.0.scan_on_push", acctest.CtTrue),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test that the removal of the non-default image_scanning_configuration causes plan changes
				Config: testAccRepositoryConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				// Test attribute update
				Config: testAccRepositoryConfig_imageScanningConfiguration(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.0.scan_on_push", acctest.CtFalse),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				// Test that the removal of the default image_scanning_configuration doesn't cause any plan changes
				Config: testAccRepositoryConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
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

func TestAccECRRepository_Encryption_kms(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 types.Repository
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"
	kmsKeyDataSourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_encryptionKMSDefaultkey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", string(types.EncryptionTypeKms)),
					// This will be the default ECR service KMS key. We don't currently have a way to look this up.
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "encryption_configuration.0.kms_key", "kms", regexache.MustCompile(fmt.Sprintf("key/%s$", verify.UUIDRegexPattern))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryConfig_encryptionKMSCustomkey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, t, resourceName, &v2),
					testAccCheckRepositoryRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", string(types.EncryptionTypeKms)),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.kms_key", kmsKeyDataSourceName, names.AttrARN),
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

func TestAccECRRepository_Encryption_aes256(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 types.Repository
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Test that the addition of the default encryption_configuration doesn't recreation in the next step
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, t, resourceName, &v1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccRepositoryConfig_encryptionAES256(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", string(types.EncryptionTypeAes256)),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.kms_key", ""),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test that the removal of the default encryption_configuration doesn't cause any plan changes
				Config: testAccRepositoryConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
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

func testAccCheckRepositoryDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECRClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecr_repository" {
				continue
			}

			_, err := tfecr.FindRepositoryByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ECR Repository %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRepositoryExists(ctx context.Context, t *testing.T, n string, v *types.Repository) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECR Repository ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).ECRClient(ctx)

		output, err := tfecr.FindRepositoryByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRepositoryRegistryID(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := acctest.AccountID(ctx)
		return resource.TestCheckResourceAttr(resourceName, "registry_id", attributeValue)(s)
	}
}

func testAccCheckRepositoryRepositoryURL(ctx context.Context, resourceName, repositoryName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", acctest.AccountID(ctx), acctest.Region(), repositoryName)
		return resource.TestCheckResourceAttr(resourceName, "repository_url", attributeValue)(s)
	}
}

func testAccCheckRepositoryRecreated(i, j *types.Repository) resource.TestCheckFunc { // nosemgrep:ci.ecr-in-func-name
	return func(s *terraform.State) error {
		if aws.ToTime(i.CreatedAt).Equal(aws.ToTime(j.CreatedAt)) {
			return fmt.Errorf("ECR repository was not recreated")
		}

		return nil
	}
}

func testAccRepositoryConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRepositoryConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRepositoryConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccRepositoryConfig_immutability(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name                 = %[1]q
  image_tag_mutability = "IMMUTABLE"
}
`, rName)
}

func testAccRepositoryConfig_immutabilityWithExclusion(rName, filter string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name                 = %[1]q
  image_tag_mutability = "IMMUTABLE_WITH_EXCLUSION"

  image_tag_mutability_exclusion_filter {
    filter      = %[2]q
    filter_type = "WILDCARD"
  }
}
`, rName, filter)
}

func testAccRepositoryConfig_mutabilityWithExclusion(rName, filter string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name                 = %[1]q
  image_tag_mutability = "MUTABLE_WITH_EXCLUSION"

  image_tag_mutability_exclusion_filter {
    filter      = %[2]q
    filter_type = "WILDCARD"
  }
}
`, rName, filter)
}

func testAccRepositoryConfig_immutabilityWithExclusionInvalid(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name                 = %[1]q
  image_tag_mutability = "MUTABLE"

  image_tag_mutability_exclusion_filter {
    filter      = "latest*"
    filter_type = "WILDCARD"
  }
}
`, rName)
}

func testAccRepositoryConfig_imageScanningConfiguration(rName string, scanOnPush bool) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q

  image_scanning_configuration {
    scan_on_push = %[2]t
  }
}
`, rName, scanOnPush)
}

func testAccRepositoryConfig_encryptionKMSDefaultkey(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q

  encryption_configuration {
    encryption_type = "KMS"
  }
}
`, rName)
}

func testAccRepositoryConfig_encryptionKMSCustomkey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_ecr_repository" "test" {
  name = %[1]q

  encryption_configuration {
    encryption_type = "KMS"
    kms_key         = aws_kms_key.test.arn
  }
}
`, rName)
}

func testAccRepositoryConfig_encryptionAES256(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q

  encryption_configuration {
    encryption_type = "AES256"
  }
}
`, rName)
}
