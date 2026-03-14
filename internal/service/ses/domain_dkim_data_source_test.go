// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESDomainDKIMDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()

	dataSourceName := "data.aws_ses_domain_dkim.test"
	resourceName := "aws_ses_domain_dkim.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SESEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop, //???
		Steps: []resource.TestStep{
			{
				Config: testAccDomainDKIMDataSourceConfig_basic(domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDomain, resourceName, names.AttrDomain),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, dataSourceName, names.AttrDomain),
					resource.TestCheckResourceAttr(dataSourceName, "dkim_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(dataSourceName, "dkim_tokens.0"),
					resource.TestCheckResourceAttrSet(dataSourceName, "dkim_tokens.1"),
					resource.TestCheckResourceAttrSet(dataSourceName, "dkim_tokens.2"),
					resource.TestCheckResourceAttr(dataSourceName, "dkim_verification_status", "Pending"),
				),
			},
			{
				Config:      testAccDomainDKIMDataSourceConfig_nonExistent,
				ExpectError: regexache.MustCompile(`couldn't find resource`),
			},
		},
	})
}

func testAccDomainDKIMDataSourceConfig_basic(domain string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_ses_domain_dkim" "test" {
  domain = aws_ses_domain_identity.test.domain
}

data "aws_ses_domain_dkim" "test" {
  domain = aws_ses_domain_dkim.test.domain
}
`, domain)
}

const testAccDomainDKIMDataSourceConfig_nonExistent = `
data "aws_ses_domain_dkim" "test" {
  domain = "tf-acc-test-does-not-exist"
}
`
