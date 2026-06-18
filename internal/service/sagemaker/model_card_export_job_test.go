// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerModelCardExportJob_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_card_export_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccModelCardExportJobConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelCardExportJobExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("model_card_export_job_arn"), tfknownvalue.RegionalARNExact("sagemaker", fmt.Sprintf("model-card/%[1]s-card/export-job/%[1]s", rName))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("export_artifacts"), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"s3_export_artifacts": knownvalue.NotNull(),
					})})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "model_card_export_job_arn"),
				ImportStateVerifyIdentifierAttribute: "model_card_export_job_arn",
			},
		},
	})
}

func testAccCheckModelCardExportJobExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		_, err := tfsagemaker.FindModelCardExportJobByARN(ctx, conn, rs.Primary.Attributes["model_card_export_job_arn"])

		return err
	}
}

func testAccModelCardExportJobConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model_card" "test" {
  model_card_name   = "%[1]s-card"
  model_card_status = "Draft"

  content = "{}"
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_sagemaker_model_card_export_job" "test" {
  model_card_export_job_name = %[1]q
  model_card_name            = aws_sagemaker_model_card.test.model_card_name

  output_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
  }
}
`, rName)
}
