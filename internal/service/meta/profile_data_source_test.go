// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfmeta "github.com/hashicorp/terraform-provider-aws/internal/service/meta"
)

func TestAccMetaProfileDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_profile.test"
	profileName := acctest.SkipIfEnvVarNotSet(t, "AWS_PROFILE")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProfileDataSourceConfig_ProviderProfile(profileName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", profileName),
					resource.TestCheckResourceAttr(dataSourceName, "id", profileName),
				),
			},
		},
	})
}

func TestAccMetaProfileDataSource_multiProvider(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_profile.test"
	profileName := acctest.SkipIfEnvVarNotSet(t, "AWS_ALTERNATE_PROFILE")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, tfmeta.PseudoServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProfileDataSourceConfig_multiProvider(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", profileName),
					resource.TestCheckResourceAttr(dataSourceName, "id", profileName),
				),
			},
		},
	})
}

func testAccProfileDataSourceConfig_ProviderProfile(name string) string {
	return fmt.Sprintf(`
provider "aws" {
	profile = %[1]q
}

data "aws_profile" "test" {}
`, name)
}

func testAccProfileDataSourceConfig_multiProvider() string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprint(`
data "aws_profile" "test" {
  provider = "awsalternate"
}
`))

}
