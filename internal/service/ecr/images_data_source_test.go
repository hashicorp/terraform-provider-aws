// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRImagesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ecr_images.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccImagesDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRepositoryName, rName),
					resource.TestCheckResourceAttrSet(dataSourceName, "image_ids.#"),
				),
			},
		},
	})
}

func TestAccECRImagesDataSource_publicRepo(t *testing.T) {
	ctx := acctest.Context(t)
	registry, repo := "137112412989", "amazonlinux"
	dataSourceName := "data.aws_ecr_images.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccImagesDataSourceConfig_publicRepo(registry, repo),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrID),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRepositoryName, repo),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", registry),
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

func testAccImagesDataSourceConfig_publicRepo(registry, repo string) string {
	return fmt.Sprintf(`
data "aws_ecr_images" "test" {
  registry_id     = %[1]q
  repository_name = %[2]q
}
`, registry, repo)
}
