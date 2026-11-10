// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccBedrockFoundationModelAgreement_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var foundationmodelagreement bedrock.GetFoundationModelAvailabilityOutput
	resourceName := "aws_bedrock_foundation_model_agreement.test"
	modelID := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_FOUNDATION_MODEL_ID")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckFoundationModelAgreement(ctx, t)
			testAccPreCheckFoundationModelUseCase(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFoundationModelAgreementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFoundationModelAgreementConfig_basic(modelID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFoundationModelAgreementExists(ctx, t, resourceName, &foundationmodelagreement),
					resource.TestCheckResourceAttrPair(resourceName, "model_id", "data.aws_bedrock_foundation_model_agreement_offers.test", "model_id"),
					resource.TestCheckResourceAttrPair(resourceName, "offer_token", "data.aws_bedrock_foundation_model_agreement_offers.test", "offers[0].offer_token"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "model_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "model_id",
				ImportStateVerifyIgnore:              []string{"offer_token"},
			},
		},
	})
}

func testAccBedrockFoundationModelAgreement_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var foundationmodelagreement bedrock.GetFoundationModelAvailabilityOutput
	resourceName := "aws_bedrock_foundation_model_agreement.test"
	modelID := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_FOUNDATION_MODEL_ID")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckFoundationModelAgreement(ctx, t)
			testAccPreCheckFoundationModelUseCase(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFoundationModelAgreementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFoundationModelAgreementConfig_basic(modelID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFoundationModelAgreementExists(ctx, t, resourceName, &foundationmodelagreement),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrock.ResourceFoundationModelAgreement, resourceName),
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

func testAccCheckFoundationModelAgreementDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_foundation_model_agreement" {
				continue
			}

			resp, err := tfbedrock.FindFoundationModelAgreementByID(ctx, conn, rs.Primary.Attributes["model_id"])
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Bedrock, create.ErrActionCheckingDestroyed, tfbedrock.ResNameFoundationModelAgreement, rs.Primary.Attributes["model_id"], err)
			}

			if resp != nil && resp.AgreementAvailability != nil && resp.AgreementAvailability.Status == awstypes.AgreementStatusNotAvailable {
				return nil
			}

			return create.Error(names.Bedrock, create.ErrActionCheckingDestroyed, tfbedrock.ResNameFoundationModelAgreement, rs.Primary.Attributes["model_id"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckFoundationModelAgreementExists(ctx context.Context, t *testing.T, name string, foundationmodelagreement *bedrock.GetFoundationModelAvailabilityOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameFoundationModelAgreement, name, errors.New("not found"))
		}

		if rs.Primary.Attributes["model_id"] == "" {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameFoundationModelAgreement, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		resp, err := tfbedrock.FindFoundationModelAgreementByID(ctx, conn, rs.Primary.Attributes["model_id"])
		if err != nil {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameFoundationModelAgreement, rs.Primary.Attributes["model_id"], err)
		}

		if resp == nil || resp.AgreementAvailability == nil || resp.AgreementAvailability.Status != awstypes.AgreementStatusAvailable {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameFoundationModelAgreement, rs.Primary.Attributes["model_id"], errors.New("not found"))
		}

		*foundationmodelagreement = *resp

		return nil
	}
}

func testAccPreCheckFoundationModelUseCase(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

	input := bedrock.GetUseCaseForModelAccessInput{}
	_, err := conn.GetUseCaseForModelAccess(ctx, &input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		t.Skipf("skipping acceptance testing due to missing use case for model access: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPreCheckFoundationModelAgreement(ctx context.Context, t *testing.T) {
	modelUsedForAgreement := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_FOUNDATION_MODEL_ID")

	conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)
	resp, err := tfbedrock.FindFoundationModelAgreementByID(ctx, conn, modelUsedForAgreement)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if resp == nil || resp.AgreementAvailability == nil || resp.AgreementAvailability.Status == awstypes.AgreementStatusAvailable {
		t.Skipf("skipping acceptance testing due to existing agreement: %s", modelUsedForAgreement)
	}
}

func testAccFoundationModelAgreementConfig_basic(modelId string) string {
	return fmt.Sprintf(`
resource "aws_bedrock_foundation_model_agreement" "test" {
  model_id    = "%[1]s"
  offer_token = data.aws_bedrock_foundation_model_agreement_offers.test.offers[0].offer_token

  lifecycle {
    ignore_changes = [offer_token]
  }
}

data "aws_bedrock_foundation_model_agreement_offers" "test" {
  model_id   = "%[1]s"
  offer_type = "PUBLIC"
}
`, modelId)
}
