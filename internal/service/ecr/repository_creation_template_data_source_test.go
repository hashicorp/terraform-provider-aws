// Copyright (c) HashiCorp, Inc.
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryCreationTemplateDataSourceConfig_basic(repositoryPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrAccountID(dataSource, "registry_id"),
					resource.TestCheckResourceAttr(dataSource, "applied_for.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(dataSource, "applied_for.*", string(types.RCTAppliedForPullThroughCache)),
					resource.TestCheckResourceAttr(dataSource, "custom_role_arn", ""),
					resource.TestCheckResourceAttr(dataSource, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(dataSource, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSource, "encryption_configuration.0.encryption_type", string(types.EncryptionTypeAes256)),
					resource.TestCheckResourceAttr(dataSource, "encryption_configuration.0.kms_key", ""),
					resource.TestCheckResourceAttr(dataSource, "image_tag_mutability", string(types.ImageTagMutabilityMutable)),
					resource.TestCheckResourceAttr(dataSource, "lifecycle_policy", ""),
					resource.TestCheckResourceAttr(dataSource, names.AttrPrefix, repositoryPrefix),
					resource.TestCheckResourceAttr(dataSource, "repository_policy", ""),
					resource.TestCheckResourceAttr(dataSource, "resource_tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSource, "resource_tags.Foo", "Bar"),
				),
			},
		},
	})
}

func TestAccECRRepositoryCreationTemplateDataSource_root(t *testing.T) {
	ctx := acctest.Context(t)
	dataSource := "data.aws_ecr_repository_creation_template.root"

	resource.Test(t, resource.TestCase{
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

func testAccRepositoryCreationTemplateDataSourceConfig_basic(repositoryPrefix string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository_creation_template" "test" {
  prefix = %[1]q

  applied_for = [
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
