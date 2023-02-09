package codepipeline_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodepipeline "github.com/hashicorp/terraform-provider-aws/internal/service/codepipeline"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccCodePipelineCustomActionType_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v codepipeline.ActionType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_custom_action_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(codestarconnections.EndpointsID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomActionTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomActionTypeConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomActionTypeExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "codepipeline", fmt.Sprintf("actiontype:Custom/Test/%s/1", rName)),
					resource.TestCheckResourceAttr(resourceName, "category", "Test"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.0.maximum_count", "5"),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.0.minimum_count", "0"),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.0.maximum_count", "4"),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.0.minimum_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "owner", "Custom"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", rName),
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

func TestAccCodePipelineCustomActionType_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v codepipeline.ActionType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_custom_action_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(codestarconnections.EndpointsID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomActionTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomActionTypeConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomActionTypeExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodepipeline.ResourceCustomActionType(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodePipelineCustomActionType_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v codepipeline.ActionType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_custom_action_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(codestarconnections.EndpointsID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomActionTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomActionTypeConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomActionTypeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomActionTypeConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomActionTypeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCustomActionTypeConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomActionTypeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccCodePipelineCustomActionType_allAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	var v codepipeline.ActionType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_custom_action_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(codestarconnections.EndpointsID, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codepipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomActionTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomActionTypeConfig_allAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomActionTypeExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "codepipeline", fmt.Sprintf("actiontype:Custom/Test/%s/1", rName)),
					resource.TestCheckResourceAttr(resourceName, "category", "Test"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.0.description", ""),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.0.key", "true"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.0.name", "pk"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.0.queryable", "true"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.0.required", "true"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.0.secret", "false"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.0.type", "Number"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.1.description", "Date of birth"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.1.key", "false"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.1.name", "dob"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.1.queryable", "false"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.1.required", "false"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.1.secret", "true"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.1.type", "String"),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.0.maximum_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.0.minimum_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.0.maximum_count", "5"),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.0.minimum_count", "4"),
					resource.TestCheckResourceAttr(resourceName, "owner", "Custom"),
					resource.TestCheckResourceAttr(resourceName, "provider_name", rName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.entity_url_template", "https://example.com/entity"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.execution_url_template", ""),
					resource.TestCheckResourceAttr(resourceName, "settings.0.revision_url_template", "https://example.com/configuration"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.third_party_configuration_url", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configuration_property.0.type",
					"configuration_property.1.type",
				},
			},
		},
	})
}

func testAccCheckCustomActionTypeExists(ctx context.Context, n string, v *codepipeline.ActionType) resource.TestCheckFunc {
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodePipelineConn()

		output, err := tfcodepipeline.FindCustomActionTypeByThreePartKey(ctx, conn, category, provider, version)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCustomActionTypeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodePipelineConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codepipeline_custom_action_type" {
				continue
			}

			category, provider, version, err := tfcodepipeline.CustomActionTypeParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfcodepipeline.FindCustomActionTypeByThreePartKey(ctx, conn, category, provider, version)

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
}

func testAccCustomActionTypeConfig_basic(rName string) string {
	return fmt.Sprintf(`
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

  provider_name = %[1]q
  version       = "1"
}
`, rName)
}

func testAccCustomActionTypeConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
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

  provider_name = %[1]q
  version       = "1"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccCustomActionTypeConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
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

  provider_name = %[1]q
  version       = "1"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccCustomActionTypeConfig_allAttributes(rName string) string {
	return fmt.Sprintf(`
resource "aws_codepipeline_custom_action_type" "test" {
  category = "Test"

  configuration_property {
    key       = true
    name      = "pk"
    queryable = true
    required  = true
    secret    = false
    type      = "Number"
  }

  configuration_property {
    description = "Date of birth"
    key         = false
    name        = "dob"
    queryable   = false
    required    = false
    secret      = true
    type        = "String"
  }

  input_artifact_details {
    maximum_count = 3
    minimum_count = 2
  }

  output_artifact_details {
    maximum_count = 5
    minimum_count = 4
  }

  provider_name = %[1]q
  version       = "1"

  settings {
    entity_url_template   = "https://example.com/entity"
    revision_url_template = "https://example.com/configuration"
  }
}
`, rName)
}
