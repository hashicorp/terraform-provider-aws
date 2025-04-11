// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecrpublic_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"

	"github.com/hashicorp/terraform-provider-aws/names"
)

// To run this test, ensure that the Docker daemon is running on your machine.
func TestAccECRPublicImageDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	tag := "latest"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ecrpublic_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccImageDataSourceConfig_basic(rName, tag),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "image_pushed_at"),
					resource.TestCheckResourceAttrSet(dataSourceName, "image_size_in_bytes"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "image_tags.*", tag),
					resource.TestCheckResourceAttrSet(dataSourceName, "image_uri"),
				),
			},
		},
	})
}

func testAccImageDataSourceConfig_basic(repositoryName string, imageTag string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

data "aws_ecrpublic_authorization_token" "test" {
}

resource "aws_ecrpublic_repository" "test" {
  depends_on = [
    data.aws_ecrpublic_authorization_token.test
  ]
  repository_name = %[1]q
  force_destroy = true
  
  provisioner "local-exec" {
    command = <<-EOT
      docker pull nginx:latest
      docker tag nginx:latest ${self.repository_uri}:latest
      docker login --username AWS --password ${data.aws_ecrpublic_authorization_token.test.password} ${self.repository_uri}
      docker push ${self.repository_uri}:latest
    EOT
  }
}

data "aws_ecrpublic_image" "test" {
    depends_on = [ aws_ecrpublic_repository.test ]
	repository_name = %[1]q
	image_tag = %[2]q
}
`, repositoryName, imageTag)
}
