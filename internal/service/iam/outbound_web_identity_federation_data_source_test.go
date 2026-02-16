// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMOutboundWebIdentityFederationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_iam_outbound_web_identity_federation.test"
	resourceName := "aws_iam_outbound_web_identity_federation.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutboundWebIdentityFederationDataSourceConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "issuer_identifier", resourceName, "issuer_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "issuer_identifier"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("issuer_identifier"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("id"), knownvalue.NotNull()),
				},
			},
		},
	})
}

const testAccOutboundWebIdentityFederationDataSourceConfig_basic = `
resource "aws_iam_outbound_web_identity_federation" "test" {}

data "aws_iam_outbound_web_identity_federation" "test" {
  depends_on = [aws_iam_outbound_web_identity_federation.test]
}
`
