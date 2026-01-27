// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMOutboundWebIdentityFederationDataSource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_iam_outbound_web_identity_federation.test"
	dataSourceName := "data.aws_iam_outbound_web_identity_federation.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOutboundWebIdentityFederationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOutboundWebIdentityFederationDataSourceConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEnabled, "false"),
					resource.TestCheckNoResourceAttr(dataSourceName, "issuer_identifier"),
				),
			},
			{
				Config: testAccOutboundWebIdentityFederationDataSourceConfig_enabled,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOutboundWebIdentityFederationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEnabled, "true"),
					resource.TestCheckResourceAttrSet(dataSourceName, "issuer_identifier"),
				),
			},
		},
	})
}

const testAccOutboundWebIdentityFederationDataSourceConfig_basic = `
data "aws_iam_outbound_web_identity_federation" "test" {}
`

const testAccOutboundWebIdentityFederationDataSourceConfig_enabled = `
resource "aws_iam_outbound_web_identity_federation" "test" {}

data "aws_iam_outbound_web_identity_federation" "test" {
  depends_on = [aws_iam_outbound_web_identity_federation.test]
}
`
