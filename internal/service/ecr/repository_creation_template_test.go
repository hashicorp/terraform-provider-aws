// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfecr "github.com/hashicorp/terraform-provider-aws/internal/service/ecr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRRepositoryCreationTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_ecr_repository_creation_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryCreationTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryCreationTemplateConfig_basic(repositoryPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryCreationTemplateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "applied_for.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "applied_for.*", string(types.RCTAppliedForCreateOnPush)),
					resource.TestCheckTypeSetElemAttr(resourceName, "applied_for.*", string(types.RCTAppliedForPullThroughCache)),
					resource.TestCheckTypeSetElemAttr(resourceName, "applied_for.*", string(types.RCTAppliedForReplication)),
					resource.TestCheckResourceAttr(resourceName, "custom_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", string(types.EncryptionTypeAes256)),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.kms_key", ""),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability", string(types.ImageTagMutabilityMutable)),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrefix, repositoryPrefix),
					resource.TestCheckResourceAttr(resourceName, "repository_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_tags.Foo", "Bar"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
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

func TestAccECRRepositoryCreationTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_ecr_repository_creation_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryCreationTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryCreationTemplateConfig_basic(repositoryPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryCreationTemplateExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfecr.ResourceRepositoryCreationTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECRRepositoryCreationTemplate_failWhenAlreadyExists(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryCreationTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccRepositoryCreationTemplateConfig_failWhenAlreadyExists(repositoryPrefix),
				ExpectError: regexache.MustCompile(`TemplateAlreadyExistsException`),
			},
		},
	})
}

func TestAccECRRepositoryCreationTemplate_ignoreEquivalentLifecycle(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_ecr_repository_creation_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryCreationTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryCreationTemplateConfig_lifecycleOrder(repositoryPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryCreationTemplateExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccRepositoryCreationTemplateConfig_lifecycleNewOrder(repositoryPrefix),
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

func TestAccECRRepositoryCreationTemplate_repository(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_ecr_repository_creation_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryCreationTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryCreationTemplateConfig_repositoryInitial(repositoryPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryCreationTemplateExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, "repository_policy", regexache.MustCompile(repositoryPrefix)),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryCreationTemplateConfig_repositoryUpdated(repositoryPrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryCreationTemplateExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, "repository_policy", regexache.MustCompile(repositoryPrefix)),
					resource.TestMatchResourceAttr(resourceName, "repository_policy", regexache.MustCompile("ecr:DescribeImages")),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "registry_id"),
				),
			},
		},
	})
}

func TestAccECRRepositoryCreationTemplate_root(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecr_repository_creation_template.root"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryCreationTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryCreationTemplateConfig_root(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryCreationTemplateExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccECRRepositoryCreationTemplate_mutabilityWithExclusion(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	resourceName := "aws_ecr_repository_creation_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryCreationTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryCreationTemplateConfig_mutabilityWithExclusion(repositoryPrefix, "latest*", "prod-*"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryCreationTemplateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability", string(types.ImageTagMutabilityMutableWithExclusion)),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.0.filter", "latest*"),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.0.filter_type", string(types.ImageTagMutabilityExclusionFilterTypeWildcard)),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.1.filter", "prod-*"),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.1.filter_type", string(types.ImageTagMutabilityExclusionFilterTypeWildcard)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryCreationTemplateConfig_mutabilityWithExclusion(repositoryPrefix, "prod-*", "latest*"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryCreationTemplateExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability", string(types.ImageTagMutabilityMutableWithExclusion)),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.0.filter", "prod-*"),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.0.filter_type", string(types.ImageTagMutabilityExclusionFilterTypeWildcard)),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.1.filter", "latest*"),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability_exclusion_filter.1.filter_type", string(types.ImageTagMutabilityExclusionFilterTypeWildcard)),
				),
			},
		},
	})
}

func testAccCheckRepositoryCreationTemplateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECRClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecr_repository_creation_template" {
				continue
			}

			_, _, err := tfecr.FindRepositoryCreationTemplateByRepositoryPrefix(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ECR Repository Creation Template %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRepositoryCreationTemplateExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ECRClient(ctx)

		_, _, err := tfecr.FindRepositoryCreationTemplateByRepositoryPrefix(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccRepositoryCreationTemplateConfig_basic(repositoryPrefix string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository_creation_template" "test" {
  prefix = %[1]q

  applied_for = [
    "CREATE_ON_PUSH",
    "PULL_THROUGH_CACHE",
    "REPLICATION",
  ]

  resource_tags = {
    Foo = "Bar"
  }
}
`, repositoryPrefix)
}

func testAccRepositoryCreationTemplateConfig_failWhenAlreadyExists(repositoryPrefix string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository_creation_template" "test" {
  prefix = %[1]q

  applied_for = [
    "PULL_THROUGH_CACHE",
    "REPLICATION",
  ]
}

resource "aws_ecr_repository_creation_template" "duplicate" {
  prefix = %[1]q

  applied_for = [
    "PULL_THROUGH_CACHE",
    "REPLICATION",
  ]

  depends_on = [
    aws_ecr_repository_creation_template.test,
  ]
}
`, repositoryPrefix)
}

func testAccRepositoryCreationTemplateConfig_lifecycleOrder(repositoryPrefix string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository_creation_template" "test" {
  prefix = %[1]q

  applied_for = [
    "PULL_THROUGH_CACHE",
  ]

  lifecycle_policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Expire images older than 14 days"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 14
        }
        action = {
          type = "expire"
        }
      },
      {
        rulePriority = 2
        description  = "Expire tagged images older than 14 days"
        selection = {
          tagStatus = "tagged"
          tagPrefixList = [
            "first",
            "second",
            "third",
          ]
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 14
        }
        action = {
          type = "expire"
        }
      },
    ]
  })
}
`, repositoryPrefix)
}

func testAccRepositoryCreationTemplateConfig_lifecycleNewOrder(repositoryPrefix string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository_creation_template" "test" {
  prefix = %[1]q

  applied_for = [
    "PULL_THROUGH_CACHE",
  ]

  lifecycle_policy = jsonencode({
    rules = [
      {
        rulePriority = 2
        description  = "Expire tagged images older than 14 days"
        selection = {
          tagStatus = "tagged"
          tagPrefixList = [
            "third",
            "second",
            "first",
          ]
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 14
        }
        action = {
          type = "expire"
        }
      },
      {
        rulePriority = 1
        description  = "Expire images older than 14 days"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 14
        }
        action = {
          type = "expire"
        }
      },
    ]
  })
}
`, repositoryPrefix)
}

func testAccRepositoryCreationTemplateConfig_repositoryInitial(repositoryPrefix string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository_creation_template" "test" {
  prefix = %[1]q

  applied_for = [
    "REPLICATION",
  ]

  repository_policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Sid       = %[1]q
      Effect    = "Allow"
      Principal = "*"
      Action    = "ecr:ListImages"
    }]
  })
}
`, repositoryPrefix)
}

func testAccRepositoryCreationTemplateConfig_repositoryUpdated(repositoryPrefix string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository_creation_template" "test" {
  prefix = %[1]q

  applied_for = [
    "REPLICATION",
  ]

  repository_policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Sid       = %[1]q
      Effect    = "Allow"
      Principal = "*"
      Action = [
        "ecr:ListImages",
        "ecr:DescribeImages",
      ]
    }]
  })
}
`, repositoryPrefix)
}

func testAccRepositoryCreationTemplateConfig_root() string {
	return `
resource "aws_ecr_repository_creation_template" "root" {
  prefix = "ROOT"

  applied_for = [
    "PULL_THROUGH_CACHE",
  ]
}
`
}

func testAccRepositoryCreationTemplateConfig_mutabilityWithExclusion(repositoryPrefix, filter1, filter2 string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository_creation_template" "test" {
  prefix = %[1]q

  applied_for = [
    "PULL_THROUGH_CACHE",
    "REPLICATION",
  ]

  resource_tags = {
    Foo = "Bar"
  }

  image_tag_mutability = "MUTABLE_WITH_EXCLUSION"

  image_tag_mutability_exclusion_filter {
    filter      = %[2]q
    filter_type = "WILDCARD"
  }

  image_tag_mutability_exclusion_filter {
    filter      = %[3]q
    filter_type = "WILDCARD"
  }
}
`, repositoryPrefix, filter1, filter2)
}
