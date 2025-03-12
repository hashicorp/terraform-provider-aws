// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESDomainIdentityDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.RandomDomainName()

	dataSourceName := "data.aws_ses_domain_identity.test"
	resourceName := "aws_ses_domain_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainIdentityDataSourceConfig_basic(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainIdentityExists(ctx, dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDomain, resourceName, names.AttrDomain),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, dataSourceName, names.AttrDomain),
					resource.TestCheckResourceAttrPair(dataSourceName, "verification_token", resourceName, "verification_token"),
				),
			},
		},
	})
}

func testAccDomainIdentityDataSourceConfig_basic(domain string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

data "aws_ses_domain_identity" "test" {
  depends_on = [aws_ses_domain_identity.test]
  domain     = %[2]q
}
`, domain, domain)
}
