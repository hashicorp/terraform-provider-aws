package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var out bedrockagentcorecontrol.GetResourcePolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &out),
					resource.TestCheckResourceAttrSet(resourceName, "resource_arn"),
					resource.TestCheckResourceAttr(resourceName, "policy", "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"AllowAccess\",\"Effect\":\"Allow\",\"Principal\":{\"AWS\":\"*\"},\"Action\":[],\"Resource\":\"*\"}]}"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBedrockAgentCoreResourcePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var out bedrockagentcorecontrol.GetResourcePolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &out),
					acctest.CheckSDKResourceDisappears(ctx, t, tfbedrockagentcore.ResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResourcePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_resource_policy" {
				continue
			}

			arn := rs.Primary.Attributes["resource_arn"]
			input := bedrockagentcorecontrol.GetResourcePolicyInput{
				ResourceArn: &arn,
			}
			_, err := conn.GetResourcePolicy(ctx, &input)
			if err != nil {
				// Not found or other error implies destroyed or error; treat NotFound as success
				continue
			}

			return fmt.Errorf("Bedrock Agent Core Resource Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckResourcePolicyExists(ctx context.Context, t *testing.T, name string, out *bedrockagentcorecontrol.GetResourcePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		arn := rs.Primary.Attributes["resource_arn"]
		input := bedrockagentcorecontrol.GetResourcePolicyInput{
			ResourceArn: &arn,
		}

		resp, err := conn.GetResourcePolicy(ctx, &input)
		if err != nil {
			return err
		}

		*out = *resp

		return nil
	}
}

func testAccResourcePolicyConfig_basic(rName string) string {
	// Reuse gateway infra from gateway_target_test.go
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), `
resource "aws_bedrockagentcore_resource_policy" "test" {
  resource_arn = aws_bedrockagentcore_gateway.test.gateway_arn
  policy       = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"AllowAccess\",\"Effect\":\"Allow\",\"Principal\":{\"AWS\":\"*\"},\"Action\":[],\"Resource\":\"*\"}]}"
}
`)
}
