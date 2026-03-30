// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRRepositoryDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"
	dataSourceName := "data.aws_ecr_repository.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "registry_id", dataSourceName, "registry_id"),
					resource.TestCheckResourceAttrPair(resourceName, "repository_url", dataSourceName, "repository_url"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, "image_scanning_configuration.#", dataSourceName, "image_scanning_configuration.#"),
					resource.TestCheckResourceAttrPair(resourceName, "image_tag_mutability", dataSourceName, "image_tag_mutability"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.#", dataSourceName, "encryption_configuration.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "most_recent_image_tags.#"),
				),
			},
		},
	})
}

func TestAccECRRepositoryDataSource_encryption(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"
	dataSourceName := "data.aws_ecr_repository.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryDataSourceConfig_encryption(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "registry_id", dataSourceName, "registry_id"),
					resource.TestCheckResourceAttrPair(resourceName, "repository_url", dataSourceName, "repository_url"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, "image_scanning_configuration.#", dataSourceName, "image_scanning_configuration.#"),
					resource.TestCheckResourceAttrPair(resourceName, "image_tag_mutability", dataSourceName, "image_tag_mutability"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.#", dataSourceName, "encryption_configuration.#"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.encryption_type", dataSourceName, "encryption_configuration.0.encryption_type"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.kms_key", dataSourceName, "encryption_configuration.0.kms_key"),
				),
			},
		},
	})
}

func TestAccECRRepositoryDataSource_mutabilityWithExclusion(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecr_repository.test"
	dataSourceName := "data.aws_ecr_repository.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryDataSourceConfig_mutabilityWithExclusion(rName, "test*"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "image_tag_mutability_exclusion_filter.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "image_tag_mutability_exclusion_filter.0.filter", dataSourceName, "image_tag_mutability_exclusion_filter.0.filter"),
					resource.TestCheckResourceAttrPair(resourceName, "image_tag_mutability_exclusion_filter.0.filter_type", dataSourceName, "image_tag_mutability_exclusion_filter.0.filter_type"),
				),
			},
		},
	})
}

func TestAccECRRepositoryDataSource_nonExistent(t *testing.T) {
	ctx := acctest.Context(t)
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccRepositoryDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`couldn't find resource`),
			},
		},
	})
}

const testAccRepositoryDataSourceConfig_nonExistent = `
data "aws_ecr_repository" "test" {
  name = "tf-acc-test-non-existent"
}
`

func testAccRepositoryDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %q

  tags = {
    Environment = "production"
    Usage       = "original"
  }
}

data "aws_ecr_repository" "test" {
  name = aws_ecr_repository.test.name
}
`, rName)
}

func testAccRepositoryDataSourceConfig_encryption(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_ecr_repository" "test" {
  name = %q

  encryption_configuration {
    encryption_type = "KMS"
    kms_key         = aws_kms_key.test.arn
  }
}

data "aws_ecr_repository" "test" {
  name = aws_ecr_repository.test.name
}
`, rName)
}

func testAccRepositoryDataSourceConfig_mutabilityWithExclusion(rName, filter string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name                 = %[1]q
  image_tag_mutability = "MUTABLE_WITH_EXCLUSION"

  image_tag_mutability_exclusion_filter {
    filter      = %[2]q
    filter_type = "WILDCARD"
  }
}
data "aws_ecr_repository" "test" {
  name = aws_ecr_repository.test.name
}
`, rName, filter)
}
