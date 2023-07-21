// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecrpublic_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccECRPublicAuthorizationTokenDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecrpublic_authorization_token.repo"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, endpoints.UsEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, ecr.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizationTokenDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "authorization_token"),
					resource.TestCheckResourceAttrSet(dataSourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(dataSourceName, "user_name"),
					resource.TestMatchResourceAttr(dataSourceName, "user_name", regexp.MustCompile(`AWS`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "password"),
				),
			},
		},
	})
}

func testAccAuthorizationTokenDataSourceConfig_basic() string {
	return `data "aws_ecrpublic_authorization_token" "repo" {}
`
}
