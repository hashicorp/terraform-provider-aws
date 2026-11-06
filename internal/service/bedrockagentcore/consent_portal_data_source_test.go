// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreConsentPortalDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_consent_portal.test"
	var checks []resource.TestCheckFunc
	for _, ds := range []string{"data.aws_bedrockagentcore_consent_portal.by_id", "data.aws_bedrockagentcore_consent_portal.by_arn"} {
		for _, attr := range []string{"consent_portal_id", "consent_portal_arn", "portal_url", names.AttrExecutionRoleARN, names.AttrName, "idp_config.#", "idp_config.0.credential_provider_arn", "idp_config.0.scopes.#", "sources.0.identifier", "sources.0.type"} {
			checks = append(checks, resource.TestCheckResourceAttrPair(ds, attr, resourceName, attr))
		}
	}
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckConsentPortals(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConsentPortalDestroy(ctx, t),
		Steps: []resource.TestStep{{
			Config: testAccConsentPortalDataSourceConfig_basic(rName),
			Check:  resource.ComposeAggregateTestCheckFunc(checks...),
		}},
	})
}

func testAccConsentPortalDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccConsentPortalConfig_basic(rName), `
data "aws_bedrockagentcore_consent_portal" "by_id" {
  consent_portal_identifier = aws_bedrockagentcore_consent_portal.test.consent_portal_id
}
data "aws_bedrockagentcore_consent_portal" "by_arn" {
  consent_portal_identifier = aws_bedrockagentcore_consent_portal.test.consent_portal_arn
}
`)
}
