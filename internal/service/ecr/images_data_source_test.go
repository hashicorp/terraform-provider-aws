// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRImagesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_ecr_images.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccImagesDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRepositoryName, rName),
					resource.TestCheckResourceAttr(dataSourceName, "image_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccECRImagesDataSource_registryID(t *testing.T) {
	ctx := acctest.Context(t)
	registryID := "137112412989"
	repositoryName := "amazonlinux"
	dataSourceName := "data.aws_ecr_images.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccImagesDataSourceConfig_registryID(registryID, repositoryName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRepositoryName, repositoryName),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", registryID),
					resource.TestCheckResourceAttrSet(dataSourceName, "image_ids.#"),
					// Check that we have at least one image with the "latest" tag
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "image_ids.*", map[string]string{
						"image_tag": "latest",
					}),
				),
			},
		},
	})
}

func testAccImagesDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q
}

resource "aws_ecr_repository_policy" "test" {
  repository = aws_ecr_repository.test.name

  policy = jsonencode({
    Version = "2008-10-17"
    Statement = [{
      Sid       = "new policy"
      Effect    = "Allow"
      Principal = "*"
      Action = [
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage",
        "ecr:BatchCheckLayerAvailability",
        "ecr:PutImage",
        "ecr:InitiateLayerUpload",
        "ecr:UploadLayerPart",
        "ecr:CompleteLayerUpload",
        "ecr:DescribeRepositories",
        "ecr:GetRepositoryPolicy",
        "ecr:ListImages",
        "ecr:DeleteRepository",
        "ecr:BatchDeleteImage",
        "ecr:SetRepositoryPolicy",
        "ecr:DeleteRepositoryPolicy"
      ]
    }]
  })
}

data "aws_ecr_images" "test" {
  repository_name = aws_ecr_repository.test.name
}
`, rName)
}

func testAccImagesDataSourceConfig_registryID(registryID, repositoryName string) string {
	return fmt.Sprintf(`
data "aws_ecr_images" "test" {
  registry_id     = %[1]q
  repository_name = %[2]q
}
`, registryID, repositoryName)
}
