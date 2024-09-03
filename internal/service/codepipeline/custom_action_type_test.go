// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codepipeline_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/codepipeline/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodepipeline "github.com/hashicorp/terraform-provider-aws/internal/service/codepipeline"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodePipelineCustomActionType_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.ActionType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_custom_action_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomActionTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomActionTypeConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomActionTypeExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "codepipeline", fmt.Sprintf("actiontype:Custom/Test/%s/1", rName)),
					resource.TestCheckResourceAttr(resourceName, "category", "Test"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.0.maximum_count", "5"),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.0.minimum_count", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.0.maximum_count", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.0.minimum_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrOwner, "Custom"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProviderName, rName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
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
	var v types.ActionType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_custom_action_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
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
	var v types.ActionType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_custom_action_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomActionTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomActionTypeConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomActionTypeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomActionTypeConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomActionTypeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCustomActionTypeConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomActionTypeExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccCodePipelineCustomActionType_allAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.ActionType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codepipeline_custom_action_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodePipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomActionTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomActionTypeConfig_allAttributes(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomActionTypeExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "codepipeline", fmt.Sprintf("actiontype:Custom/Test/%s/1", rName)),
					resource.TestCheckResourceAttr(resourceName, "category", "Test"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.0.description", ""),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.0.key", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.0.name", "pk"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.0.queryable", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.0.required", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.0.secret", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.0.type", "Number"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.1.description", "Date of birth"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.1.key", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.1.name", "dob"),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.1.queryable", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.1.required", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.1.secret", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "configuration_property.1.type", "String"),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.0.maximum_count", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "input_artifact_details.0.minimum_count", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.0.maximum_count", "5"),
					resource.TestCheckResourceAttr(resourceName, "output_artifact_details.0.minimum_count", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, names.AttrOwner, "Custom"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProviderName, rName),
					resource.TestCheckResourceAttr(resourceName, "settings.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "settings.0.entity_url_template", "https://example.com/entity"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.execution_url_template", ""),
					resource.TestCheckResourceAttr(resourceName, "settings.0.revision_url_template", "https://example.com/configuration"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.third_party_configuration_url", ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
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

func testAccCheckCustomActionTypeExists(ctx context.Context, n string, v *types.ActionType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		category, provider, version, err := tfcodepipeline.CustomActionTypeParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodePipelineClient(ctx)

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
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodePipelineClient(ctx)

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
