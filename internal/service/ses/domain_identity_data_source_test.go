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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainIdentityDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainIdentityDataSourceConfig_basic(domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainIdentityExists(ctx, "aws_ses_domain_identity.test"),
					testAccCheckDomainIdentityARN("data.aws_ses_domain_identity.test", domain),
				),
			},
		},
	})
}

func testAccDomainIdentityDataSourceConfig_basic(domain string) string {
	return fmt.Sprintf(`
resource "aws_ses_domain_identity" "test" {
  domain = "%s"
}

data "aws_ses_domain_identity" "test" {
  depends_on = [aws_ses_domain_identity.test]
  domain     = "%s"
}
`, domain, domain)
}
