// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecrpublic_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRPublicAuthorizationTokenDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecrpublic_authorization_token.repo"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizationTokenDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "authorization_token"),
					resource.TestCheckResourceAttrSet(dataSourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrUserName),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrUserName, regexache.MustCompile(`AWS`)),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrPassword),
				),
			},
		},
	})
}

func testAccAuthorizationTokenDataSourceConfig_basic() string {
	return `data "aws_ecrpublic_authorization_token" "repo" {}
`
}
