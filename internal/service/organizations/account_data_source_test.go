// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccountDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	orgsEmailDomain := acctest.SkipIfEnvVarNotSet(t, "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")

	dataSourceName := "data.aws_organizations_account.test"

	rInt := acctest.RandInt(t)
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountDataSourceConfig_basic(name, email),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrGlobalARN(ctx, dataSourceName, names.AttrARN, "organizations", regexache.MustCompile(`account/`+organizationIDRegexPattern+`/\d{12}$`)),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEmail, email),
					resource.TestCheckResourceAttrSet(dataSourceName, "joined_method"),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "joined_timestamp"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, name),
					resource.TestCheckResourceAttrSet(dataSourceName, "parent_id"),
				),
			},
		},
	})
}

func testAccAccountDataSourceConfig_basic(name string, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "test" {
  name              = %[1]q
  email             = %[2]q
  close_on_deletion = true

  timeouts {
    delete = "30m"
  }
}

data "aws_organizations_account" "test" {
  account_id = aws_organizations_account.test.id
}


`, name, email)
}
