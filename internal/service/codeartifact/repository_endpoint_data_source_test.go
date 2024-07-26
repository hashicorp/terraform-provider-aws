// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeartifact_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRepositoryEndpointDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_codeartifact_repository_endpoint.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryEndpointDataSourceConfig_basic(rName, "npm"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
			{
				Config: testAccRepositoryEndpointDataSourceConfig_basic(rName, "pypi"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
			{
				Config: testAccRepositoryEndpointDataSourceConfig_basic(rName, "maven"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
			{
				Config: testAccRepositoryEndpointDataSourceConfig_basic(rName, "nuget"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
		},
	})
}

func testAccRepositoryEndpointDataSource_owner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_codeartifact_repository_endpoint.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryEndpointDataSourceConfig_owner(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
		},
	})
}

func testAccCheckRepositoryEndpointBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain
}
`, rName)
}

func testAccRepositoryEndpointDataSourceConfig_basic(rName, format string) string {
	return acctest.ConfigCompose(
		testAccCheckRepositoryEndpointBaseConfig(rName),
		fmt.Sprintf(`
data "aws_codeartifact_repository_endpoint" "test" {
  domain     = aws_codeartifact_domain.test.domain
  repository = aws_codeartifact_repository.test.repository
  format     = %[1]q
}
`, format))
}

func testAccRepositoryEndpointDataSourceConfig_owner(rName string) string {
	return acctest.ConfigCompose(
		testAccCheckRepositoryEndpointBaseConfig(rName),
		`
data "aws_codeartifact_repository_endpoint" "test" {
  domain       = aws_codeartifact_domain.test.domain
  repository   = aws_codeartifact_repository.test.repository
  domain_owner = aws_codeartifact_domain.test.owner
  format       = "npm"
}
`)
}
