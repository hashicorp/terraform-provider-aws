// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfobservabilityadmin "github.com/hashicorp/terraform-provider-aws/internal/service/observabilityadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// testAccDatasetIntegrationPreCheck skips when the account already has a dataset integration
// in the Region, as only one can exist per account and Region.
func testAccDatasetIntegrationPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

	input := observabilityadmin.ListDatasetIntegrationsInput{}
	output, err := conn.ListDatasetIntegrations(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if len(output.DatasetIntegrationSummaries) > 0 {
		t.Skipf("skipping acceptance testing: dataset integration %s already exists", aws.ToString(output.DatasetIntegrationSummaries[0].Arn))
	}
}

func testAccDatasetIntegration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_observabilityadmin_dataset_integration.test"
	var v observabilityadmin.GetDatasetIntegrationOutput

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccDatasetIntegrationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatasetIntegrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/DatasetIntegration/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasetIntegrationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("observabilityadmin", "dataset-integration/default")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCreatedAt), knownvalue.NotNull()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrRoleARN), "aws_iam_role.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("updated_at"), knownvalue.NotNull()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/DatasetIntegration/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func testAccDatasetIntegration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_observabilityadmin_dataset_integration.test"
	var v observabilityadmin.GetDatasetIntegrationOutput

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccDatasetIntegrationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatasetIntegrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/DatasetIntegration/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasetIntegrationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfobservabilityadmin.ResourceDatasetIntegration, resourceName),
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

func testAccDatasetIntegration_roleARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_observabilityadmin_dataset_integration.test"
	var v1, v2 observabilityadmin.GetDatasetIntegrationOutput

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccDatasetIntegrationPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatasetIntegrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/DatasetIntegration/update_role/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"role_index":    config.IntegerVariable(0),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasetIntegrationExists(ctx, t, resourceName, &v1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrRoleARN), "aws_iam_role.test[0]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/DatasetIntegration/update_role/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"role_index":    config.IntegerVariable(0),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/DatasetIntegration/update_role/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"role_index":    config.IntegerVariable(1),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDatasetIntegrationExists(ctx, t, resourceName, &v2),
					testAccCheckDatasetIntegrationNotRecreated(&v1, &v2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrRoleARN), "aws_iam_role.test[1]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
		},
	})
}

func testAccCheckDatasetIntegrationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_observabilityadmin_dataset_integration" {
				continue
			}

			_, err := tfobservabilityadmin.FindDatasetIntegrationByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Observability Admin Dataset Integration %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckDatasetIntegrationExists(ctx context.Context, t *testing.T, n string, v *observabilityadmin.GetDatasetIntegrationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

		output, err := tfobservabilityadmin.FindDatasetIntegrationByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckDatasetIntegrationNotRecreated(before, after *observabilityadmin.GetDatasetIntegrationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToTime(before.CreatedAt), aws.ToTime(after.CreatedAt); !before.Equal(after) {
			return fmt.Errorf("Observability Admin Dataset Integration recreated: created at %s, then %s", before, after)
		}

		return nil
	}
}
