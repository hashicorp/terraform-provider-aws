// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreHarness_model_openAIAdditionalParams(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	credentialProviderName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
			testAccPreCheckAPIKeyCredentialProviders(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_openAIAdditionalParams(rName, credentialProviderName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
					resource.TestCheckResourceAttr(resourceName, "model.0.openai_model_config.0.additional_params", `{"reasoning_effort":"high"}`),
				),
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
			},
		},
	})
}

func testAccHarnessConfig_openAIAdditionalParams(rName, credentialProviderName string) string {
	return acctest.ConfigCompose(
		testAccHarnessConfig_iamRole(rName),
		testAccAPIKeyCredentialProviderConfig_basic(credentialProviderName, "test-api-key"),
		fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    openai_model_config {
      api_key_arn = aws_bedrockagentcore_api_key_credential_provider.test.credential_provider_arn
      model_id     = "gpt-5"

      additional_params = jsonencode({
        reasoning_effort = "high"
      })
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }
}
`, rName),
	)
}
