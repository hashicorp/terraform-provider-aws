package codepipeline_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodepipeline "github.com/hashicorp/terraform-provider-aws/internal/service/codepipeline"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCodePipelineCustomActionType_basic(t *testing.T) {
	var v codepipeline.ActionType
	resourceName := "aws_codepipeline_custom_action_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(codestarconnections.EndpointsID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomActionTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomActionType_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomActionTypeExists(resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "codepipeline", "actiontype:Custom/Test/CodeDeploy/1"),
					resource.TestCheckResourceAttr(resourceName, "category", "Test"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.0.maximum_count", "5"),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.0.minimum_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.0.maximum_count", "4"),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.0.minimum_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "owner", "Custom"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", "CodeDeploy"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
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

func testAccCheckCustomActionTypeExists(n string, v *codepipeline.ActionType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CodePipeline Custom Action Type ID is set")
		}

		category, provider, version, err := tfcodepipeline.CustomActionTypeParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodePipelineConn

		output, err := tfcodepipeline.FindCustomActionTypeByThreePartKey(context.Background(), conn, category, provider, version)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCustomActionTypeDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodePipelineConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codepipeline_custom_action_type" {
			continue
		}

		category, provider, version, err := tfcodepipeline.CustomActionTypeParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfcodepipeline.FindCustomActionTypeByThreePartKey(context.Background(), conn, category, provider, version)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CodePipeline Custom Action Type %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCustomActionType_basic() string {
	return `
resource "aws_codepipeline_custom_action_type" "test" {
  category = "Test"

  input_artifact_details {
    maximum_count = 5
    minimum_count = 0
  }

  output_artifact_details {
    maximum_count = 4
    minimum_count = 1
  }

  provider_name = "CodeDeploy"
  version       = "1"
}
`
}
