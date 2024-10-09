// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecommit_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeCommitRepositoryDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := fmt.Sprintf("tf-acctest-%d", sdkacctest.RandInt())
	resourceName := "aws_codecommit_repository.default"
	datasourceName := "data.aws_codecommit_repository.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeCommitServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "clone_url_http", resourceName, "clone_url_http"),
					resource.TestCheckResourceAttrPair(datasourceName, "clone_url_ssh", resourceName, "clone_url_ssh"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrRepositoryName, resourceName, names.AttrRepositoryName),
				),
			},
		},
	})
}

func testAccRepositoryDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_repository" "default" {
  repository_name = "%s"
}

data "aws_codecommit_repository" "default" {
  repository_name = aws_codecommit_repository.default.repository_name
}
`, rName)
}
