// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflakeformation "github.com/hashicorp/terraform-provider-aws/internal/service/lakeformation"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccLakeFormationIdentityCenterConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var identityCenterConfiguration lakeformation.DescribeLakeFormationIdentityCenterConfigurationOutput
	resourceName := "aws_lakeformation_identity_center_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentityCenterConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityCenterConfigurationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentityCenterConfigurationExists(ctx, t, resourceName, &identityCenterConfiguration),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("application_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrCatalogID), tfknownvalue.AccountID()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("instance_arn"), "data.aws_ssoadmin_instances.test", tfjsonpath.New(names.AttrARNs).AtSliceIndex(0), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("resource_share"), knownvalue.Null()),
				},
			},
		},
	})
}

func testAccLakeFormationIdentityCenterConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var identitycenterconfiguration lakeformation.DescribeLakeFormationIdentityCenterConfigurationOutput
	resourceName := "aws_lakeformation_identity_center_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LakeFormationEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LakeFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdentityCenterConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIdentityCenterConfigurationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIdentityCenterConfigurationExists(ctx, t, resourceName, &identitycenterconfiguration),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflakeformation.ResourceIdentityCenterConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckIdentityCenterConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LakeFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lakeformation_identity_center_configuration" {
				continue
			}

			_, err := tflakeformation.FindIdentityCenterConfigurationByID(ctx, conn, rs.Primary.Attributes[names.AttrCatalogID])
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.LakeFormation, create.ErrActionCheckingDestroyed, tflakeformation.ResNameIdentityCenterConfiguration, rs.Primary.Attributes[names.AttrCatalogID], err)
			}

			return create.Error(names.LakeFormation, create.ErrActionCheckingDestroyed, tflakeformation.ResNameIdentityCenterConfiguration, rs.Primary.Attributes[names.AttrCatalogID], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckIdentityCenterConfigurationExists(ctx context.Context, t *testing.T, name string, identitycenterconfiguration *lakeformation.DescribeLakeFormationIdentityCenterConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameIdentityCenterConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.Attributes[names.AttrCatalogID] == "" {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameIdentityCenterConfiguration, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).LakeFormationClient(ctx)

		resp, err := tflakeformation.FindIdentityCenterConfigurationByID(ctx, conn, rs.Primary.Attributes[names.AttrCatalogID])
		if err != nil {
			return create.Error(names.LakeFormation, create.ErrActionCheckingExistence, tflakeformation.ResNameIdentityCenterConfiguration, rs.Primary.Attributes[names.AttrCatalogID], err)
		}

		*identitycenterconfiguration = *resp

		return nil
	}
}

func testAccIdentityCenterConfigurationConfig_basic() string {
	return `
resource "aws_lakeformation_identity_center_configuration" "test" {
  instance_arn = local.identity_center_instance_arn
}

locals {
  identity_center_instance_arn = data.aws_ssoadmin_instances.test.arns[0]
}

data "aws_ssoadmin_instances" "test" {}
`
}
