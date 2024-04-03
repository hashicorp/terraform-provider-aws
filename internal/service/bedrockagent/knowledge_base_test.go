// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentKnowledgeBase_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// var knowledgebase types.KnowledgeBase
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_basic(rName),
				Check:  resource.ComposeAggregateTestCheckFunc(
				// testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
				// resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
				// resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
				// resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
				// 	"console_access": "false",
				// 	"groups.#":       "0",
				// 	"username":       "Test",
				// 	"password":       "TestTest1234",
				// }),
				// acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "bedrockagent", regexache.MustCompile(`knowledgebase:+.`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

// func TestAccBedrockAgentKnowledgeBase_disappears(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	if testing.Short() {
// 		t.Skip("skipping long-running test in short mode")
// 	}

// 	var knowledgebase bedrockagent.DescribeKnowledgeBaseResponse
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	resourceName := "aws_bedrockagent_knowledge_base.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(ctx, t)
// 			acctest.PreCheckPartitionHasService(t, names.BedrockAgentEndpointID)
// 			testAccPreCheck(t)
// 		},
// 		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccKnowledgeBaseConfig_basic(rName, testAccKnowledgeBaseVersionNewer),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckKnowledgeBaseExists(ctx, resourceName, &knowledgebase),
// 					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
// 					// but expects a new resource factory function as the third argument. To expose this
// 					// private function to the testing package, you may need to add a line like the following
// 					// to exports_test.go:
// 					//
// 					//   var ResourceKnowledgeBase = newResourceKnowledgeBase
// 					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagent.ResourceKnowledgeBase, resourceName),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

func testAccCheckKnowledgeBaseDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagent_knowledge_base" {
				continue
			}

			_, err := tfbedrockagent.FindKnowledgeBaseByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent knowledge base %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckKnowledgeBaseExists(ctx context.Context, n string, v *types.KnowledgeBase) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentClient(ctx)

		output, err := tfbedrockagent.FindKnowledgeBaseByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccKnowledgeBaseConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  name             = %[1]q
  role_arn    = "arn:aws:iam::182198062889:role/service-role/AmazonBedrockExecutionRoleForKnowledgeBase_22tjk"
  knowledge_base_configuration {
	vector_knowledge_base_configuration {
		embedding_model_arn = "arn:aws:bedrock:us-west-2::foundation-model/amazon.titan-embed-g1-text-02"
	}
	type = "VECTOR"
  }
  storage_configuration {
	type = "OPENSEARCH_SERVERLESS"
	opensearch_serverless_configuration {
		collection_arn = "arn:aws:aoss:us-west-2:182198062889:collection/v5p0k3ztulr1nhz8j1gg"
		vector_index_name = "bedrock-knowledge-base-default-index"
		field_mapping {
			text_field = "AMAZON_BEDROCK_TEXT_CHUNK"
			vector_field = "bedrock-knowledge-base-default-vector"
			metadata_field = "AMAZON_BEDROCK_METADATA"
		}
	}
  }
}
`, rName)
}
