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

func TestAccECRAuthorizationTokenDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rID := "111111111111"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_ecr_authorization_token.repo"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizationTokenDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "authorization_token"),
					resource.TestCheckResourceAttrSet(dataSourceName, "proxy_endpoint"),
					resource.TestCheckResourceAttrSet(dataSourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrUserName),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrUserName, regexache.MustCompile(`AWS`)),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrPassword),
				),
			},
			{
				Config: testAccAuthorizationTokenDataSourceConfig_repository(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "registry_id", "aws_ecr_repository.repo", "registry_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "authorization_token"),
					resource.TestCheckResourceAttrSet(dataSourceName, "proxy_endpoint"),
					resource.TestCheckResourceAttrSet(dataSourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrUserName),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrUserName, regexache.MustCompile(`AWS`)),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrPassword),
				),
			},
			{
				Config: testAccAuthorizationTokenDataSourceConfig_registry(rID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "authorization_token"),
					resource.TestCheckResourceAttrSet(dataSourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrUserName),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrUserName, regexache.MustCompile(`AWS`)),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrPassword),
					resource.TestMatchResourceAttr(dataSourceName, "proxy_endpoint", regexache.MustCompile(`.*`+rID+`.*`)),
				),
			},
		},
	})
}

var testAccAuthorizationTokenDataSourceConfig_basic = `
data "aws_ecr_authorization_token" "repo" {}
`

func testAccAuthorizationTokenDataSourceConfig_repository(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "repo" {
  name = %q
}

data "aws_ecr_authorization_token" "repo" {
  registry_id = aws_ecr_repository.repo.registry_id
}
`, rName)
}

func testAccAuthorizationTokenDataSourceConfig_registry(rID string) string {
	return fmt.Sprintf(`
data "aws_ecr_authorization_token" "repo" {
  registry_id = %q
}
`, rID)
}
