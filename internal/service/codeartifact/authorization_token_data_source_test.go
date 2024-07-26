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

func testAccAuthorizationTokenDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_codeartifact_authorization_token.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizationTokenDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "authorization_token"),
					resource.TestCheckResourceAttrSet(dataSourceName, "expiration"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
		},
	})
}

func testAccAuthorizationTokenDataSource_owner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_codeartifact_authorization_token.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizationTokenDataSourceConfig_owner(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "authorization_token"),
					resource.TestCheckResourceAttrSet(dataSourceName, "expiration"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
		},
	})
}

func testAccAuthorizationTokenDataSource_duration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_codeartifact_authorization_token.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizationTokenDataSourceConfig_duration(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "authorization_token"),
					resource.TestCheckResourceAttrSet(dataSourceName, "expiration"),
					resource.TestCheckResourceAttr(dataSourceName, "duration_seconds", "900"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
		},
	})
}

func testAccCheckAuthorizationTokenConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}
`, rName)
}

func testAccAuthorizationTokenDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccCheckAuthorizationTokenConfig_base(rName),
		`
data "aws_codeartifact_authorization_token" "test" {
  domain = aws_codeartifact_domain.test.domain
}
`)
}

func testAccAuthorizationTokenDataSourceConfig_owner(rName string) string {
	return acctest.ConfigCompose(
		testAccCheckAuthorizationTokenConfig_base(rName),
		`
data "aws_codeartifact_authorization_token" "test" {
  domain       = aws_codeartifact_domain.test.domain
  domain_owner = aws_codeartifact_domain.test.owner
}
`)
}

func testAccAuthorizationTokenDataSourceConfig_duration(rName string) string {
	return acctest.ConfigCompose(
		testAccCheckAuthorizationTokenConfig_base(rName),
		`
data "aws_codeartifact_authorization_token" "test" {
  domain           = aws_codeartifact_domain.test.domain
  duration_seconds = 900
}
`)
}
