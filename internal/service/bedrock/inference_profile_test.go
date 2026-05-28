// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Regions are hard coded due to limited availability of Bedrock service
const (
	foundationModelARN = "arn:aws:bedrock:us-west-2::foundation-model/anthropic.claude-3-5-sonnet-20240620-v1:0" // lintignore:AWSAT003,AWSAT005
)

func TestAccBedrockInferenceProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var inferenceprofile bedrock.GetInferenceProfileOutput

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_inference_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceProfileConfig_basic(rName, foundationModelARN),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInferenceProfileExists(ctx, t, resourceName, &inferenceprofile),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock", regexache.MustCompile(`application-inference-profile/[a-z0-9]+$`)),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "models.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "models.0.model_arn", foundationModelARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.InferenceProfileStatusActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.InferenceProfileTypeApplication)),
					acctest.CheckResourceAttrRFC3339(resourceName, "updated_at"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"model_source",
				},
			},
			// Validate a replacement is not planned following import.
			// Ref: https://github.com/hashicorp/terraform-provider-aws/issues/45705
			{
				Config: testAccInferenceProfileConfig_basic(rName, foundationModelARN),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccBedrockInferenceProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var inferenceprofile bedrock.GetInferenceProfileOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_inference_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceProfileConfig_basic(rName, foundationModelARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceProfileExists(ctx, t, resourceName, &inferenceprofile),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrock.ResourceInferenceProfile, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccBedrockInferenceProfile_description(t *testing.T) {
	ctx := acctest.Context(t)
	var inferenceprofile bedrock.GetInferenceProfileOutput

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_inference_profile.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceProfileConfig_description(rName, foundationModelARN, names.AttrDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInferenceProfileExists(ctx, t, resourceName, &inferenceprofile),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"model_source",
				},
			},
			{
				Config: testAccInferenceProfileConfig_description(rName, foundationModelARN, "updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInferenceProfileExists(ctx, t, resourceName, &inferenceprofile),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"model_source",
				},
			},
		},
	})
}

func testAccCheckInferenceProfileDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_inference_profile" {
				continue
			}

			input := &bedrock.GetInferenceProfileInput{
				InferenceProfileIdentifier: aws.String(rs.Primary.ID),
			}

			_, err := conn.GetInferenceProfile(ctx, input)

			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}

			if err != nil {
				return create.Error(names.Bedrock, create.ErrActionCheckingDestroyed, tfbedrock.ResNameInferenceProfile, rs.Primary.ID, err)
			}

			return create.Error(names.Bedrock, create.ErrActionCheckingDestroyed, tfbedrock.ResNameInferenceProfile, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckInferenceProfileExists(ctx context.Context, t *testing.T, name string, inferenceprofile *bedrock.GetInferenceProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameInferenceProfile, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameInferenceProfile, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		input := bedrock.GetInferenceProfileInput{
			InferenceProfileIdentifier: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetInferenceProfile(ctx, &input)
		if err != nil {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameInferenceProfile, rs.Primary.ID, err)
		}

		*inferenceprofile = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

	input := &bedrock.ListInferenceProfilesInput{}

	_, err := conn.ListInferenceProfiles(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccInferenceProfileConfig_basic(rName, source string) string {
	return fmt.Sprintf(`
resource "aws_bedrock_inference_profile" "test" {
  name = %[1]q

  model_source {
    copy_from = %[2]q
  }
}
`, rName, source)
}

func testAccInferenceProfileConfig_description(rName, source, description string) string {
	return fmt.Sprintf(`
resource "aws_bedrock_inference_profile" "test" {
  name        = %[1]q
  description = %[3]q

  model_source {
    copy_from = %[2]q
  }
}
`, rName, source, description)
}
