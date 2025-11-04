// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRImageDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	registry, repo, tag := "137112412989", "amazonlinux", "latest"
	resourceByTag := "data.aws_ecr_image.by_tag"
	resourceByDigest := "data.aws_ecr_image.by_digest"
	resourceByMostRecent := "data.aws_ecr_image.by_most_recent"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccImageDataSourceConfig_basic(registry, repo, tag),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceByTag, "image_digest"),
					resource.TestCheckResourceAttrSet(resourceByTag, "image_pushed_at"),
					resource.TestCheckResourceAttrSet(resourceByTag, "image_size_in_bytes"),
					resource.TestCheckTypeSetElemAttr(resourceByTag, "image_tags.*", tag),
					resource.TestCheckResourceAttrSet(resourceByTag, "image_uri"),
					resource.TestCheckResourceAttrSet(resourceByDigest, "image_pushed_at"),
					resource.TestCheckResourceAttrSet(resourceByDigest, "image_size_in_bytes"),
					resource.TestCheckTypeSetElemAttr(resourceByDigest, "image_tags.*", tag),
					resource.TestCheckResourceAttrSet(resourceByDigest, "image_uri"),
					resource.TestCheckResourceAttrSet(resourceByMostRecent, "image_pushed_at"),
					resource.TestCheckResourceAttrSet(resourceByMostRecent, "image_size_in_bytes"),
				),
			},
		},
	})
}

func testAccImageDataSourceConfig_basic(reg, repo, tag string) string {
	return fmt.Sprintf(`
data "aws_ecr_image" "by_tag" {
  registry_id     = %[1]q
  repository_name = %[2]q
  image_tag       = %[3]q
}

data "aws_ecr_image" "by_digest" {
  registry_id     = data.aws_ecr_image.by_tag.registry_id
  repository_name = data.aws_ecr_image.by_tag.repository_name
  image_digest    = data.aws_ecr_image.by_tag.image_digest
}

data "aws_ecr_image" "by_most_recent" {
  registry_id     = data.aws_ecr_image.by_tag.registry_id
  repository_name = data.aws_ecr_image.by_tag.repository_name
  most_recent     = true
}
`, reg, repo, tag)
}
