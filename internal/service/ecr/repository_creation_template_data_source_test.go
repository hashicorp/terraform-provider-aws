// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRRepositoryCreationTemplateDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	dataSource := "data.aws_ecr_repository_creation_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryCreationTemplateDataSourceConfig_basic(repositoryPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrAccountID(ctx, dataSource, "registry_id"),
					resource.TestCheckResourceAttr(dataSource, "applied_for.#", "2"),
					resource.TestCheckTypeSetElemAttr(dataSource, "applied_for.*", string(types.RCTAppliedForCreateOnPush)),
					resource.TestCheckTypeSetElemAttr(dataSource, "applied_for.*", string(types.RCTAppliedForPullThroughCache)),
					resource.TestCheckResourceAttr(dataSource, "custom_role_arn", ""),
					resource.TestCheckResourceAttr(dataSource, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(dataSource, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(dataSource, "encryption_configuration.0.encryption_type", string(types.EncryptionTypeAes256)),
					resource.TestCheckResourceAttr(dataSource, "encryption_configuration.0.kms_key", ""),
					resource.TestCheckResourceAttr(dataSource, "image_tag_mutability", string(types.ImageTagMutabilityMutable)),
					resource.TestCheckResourceAttr(dataSource, "lifecycle_policy", ""),
					resource.TestCheckResourceAttr(dataSource, names.AttrPrefix, repositoryPrefix),
					resource.TestCheckResourceAttr(dataSource, "repository_policy", ""),
					resource.TestCheckResourceAttr(dataSource, "resource_tags.%", "1"),
					resource.TestCheckResourceAttr(dataSource, "resource_tags.Foo", "Bar"),
				),
			},
		},
	})
}

func TestAccECRRepositoryCreationTemplateDataSource_root(t *testing.T) {
	ctx := acctest.Context(t)
	dataSource := "data.aws_ecr_repository_creation_template.root"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryCreationTemplateDataSourceConfig_root(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource, names.AttrPrefix, "ROOT"),
				),
			},
		},
	})
}

func TestAccECRRepositoryCreationTemplateDataSource_mutabilityWithExclusion(t *testing.T) {
	ctx := acctest.Context(t)
	repositoryPrefix := "tf-test-" + sdkacctest.RandString(8)
	dataSource := "data.aws_ecr_repository_creation_template.root"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryCreationTemplateDataSourceConfig_mutabilityWithExclusion(repositoryPrefix, "latest*", "prod-*"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource, "image_tag_mutability", string(types.ImageTagMutabilityMutableWithExclusion)),
					resource.TestCheckResourceAttr(dataSource, "image_tag_mutability_exclusion_filter.#", "2"),
					resource.TestCheckResourceAttr(dataSource, "image_tag_mutability_exclusion_filter.0.filter", "latest*"),
					resource.TestCheckResourceAttr(dataSource, "image_tag_mutability_exclusion_filter.0.filter_type", string(types.ImageTagMutabilityExclusionFilterTypeWildcard)),
					resource.TestCheckResourceAttr(dataSource, "image_tag_mutability_exclusion_filter.1.filter", "prod-*"),
					resource.TestCheckResourceAttr(dataSource, "image_tag_mutability_exclusion_filter.1.filter_type", string(types.ImageTagMutabilityExclusionFilterTypeWildcard)),
				),
			},
		},
	})
}

func testAccRepositoryCreationTemplateDataSourceConfig_basic(repositoryPrefix string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository_creation_template" "test" {
  prefix = %[1]q

  applied_for = [
    "CREATE_ON_PUSH",
    "PULL_THROUGH_CACHE",
  ]

  resource_tags = {
    Foo = "Bar"
  }
}

data "aws_ecr_repository_creation_template" "test" {
  prefix = aws_ecr_repository_creation_template.test.prefix
}
`, repositoryPrefix)
}

func testAccRepositoryCreationTemplateDataSourceConfig_root() string {
	return `
resource "aws_ecr_repository_creation_template" "root" {
  prefix = "ROOT"

  applied_for = [
    "PULL_THROUGH_CACHE",
  ]
}

data "aws_ecr_repository_creation_template" "root" {
  prefix = aws_ecr_repository_creation_template.root.prefix
}
`
}

func testAccRepositoryCreationTemplateDataSourceConfig_mutabilityWithExclusion(repositoryPrefix, filter1, filter2 string) string {
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
data "aws_ecr_repository_creation_template" "root" {
  prefix = aws_ecr_repository_creation_template.test.prefix
}
`, repositoryPrefix, filter1, filter2)
}
