// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package controltower_test

import (
	"context"
	"fmt"
	"testing"

	types "github.com/aws/aws-sdk-go-v2/service/controltower/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcontroltower "github.com/hashicorp/terraform-provider-aws/internal/service/controltower"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccControl_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var control types.EnabledControlSummary
	resourceName := "aws_controltower_control.test"
	ouDataSourceName := "data.aws_organizations_organizational_unit.test"
	ouName := acctest.SkipIfEnvVarNotSet(t, "TF_AWS_CONTROLTOWER_CONTROL_OU_NAME")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ControlTowerServiceID),
		CheckDestroy:             testAccCheckControlDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccControlConfig_basic(ouName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(ctx, t, resourceName, &control),
					resource.TestCheckResourceAttrSet(resourceName, "control_identifier"),
					resource.TestCheckResourceAttrPair(resourceName, "target_identifier", ouDataSourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "0"),
				),
			},
		},
	})
}

func testAccControl_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var control types.EnabledControlSummary
	resourceName := "aws_controltower_control.test"
	ouName := acctest.SkipIfEnvVarNotSet(t, "TF_AWS_CONTROLTOWER_CONTROL_OU_NAME")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ControlTowerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckControlDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccControlConfig_basic(ouName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(ctx, t, resourceName, &control),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcontroltower.ResourceControl(), resourceName),
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

func testAccControl_parameters(t *testing.T) {
	ctx := acctest.Context(t)
	var control types.EnabledControlSummary
	resourceName := "aws_controltower_control.test"
	ouName := acctest.SkipIfEnvVarNotSet(t, "TF_AWS_CONTROLTOWER_CONTROL_OU_NAME")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ControlTowerServiceID),
		CheckDestroy:             testAccCheckControlDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccControlConfig_parameters(ouName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(ctx, t, resourceName, &control),
					resource.TestCheckResourceAttrSet(resourceName, "control_identifier"),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.key", "ExemptedPrincipalArns"),
				),
			},
			{
				Config: testAccControlConfig_basic(ouName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(ctx, t, resourceName, &control),
					resource.TestCheckResourceAttrSet(resourceName, "control_identifier"),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccControlConfig_parameters(ouName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckControlExists(ctx, t, resourceName, &control),
					resource.TestCheckResourceAttrSet(resourceName, "control_identifier"),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.key", "ExemptedPrincipalArns"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccCheckControlExists(ctx context.Context, t *testing.T, n string, v *types.EnabledControlSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ControlTowerClient(ctx)

		output, err := tfcontroltower.FindEnabledControlByTwoPartKey(ctx, conn, rs.Primary.Attributes["target_identifier"], rs.Primary.Attributes["control_identifier"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckControlDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ControlTowerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_controltower_control" {
				continue
			}

			_, err := tfcontroltower.FindEnabledControlByTwoPartKey(ctx, conn, rs.Primary.Attributes["target_identifier"], rs.Primary.Attributes["control_identifier"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ControlTower Control %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccControlConfigBase(ouName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

data "aws_organizations_organization" "test" {}

data "aws_organizations_organizational_unit" "test" {
  parent_id = data.aws_organizations_organization.test.roots[0].id
  name      = %[1]q
}
`, ouName)
}

func testAccControlConfig_basic(ouName string) string {
	return acctest.ConfigCompose(
		testAccControlConfigBase(ouName),
		`
resource "aws_controltower_control" "test" {
  control_identifier = "arn:${data.aws_partition.current.partition}:controltower:${data.aws_region.current.region}::control/AWS-GR_DISALLOW_CROSS_REGION_NETWORKING"
  target_identifier  = data.aws_organizations_organizational_unit.test.arn
}
`)
}

// See the AWS documentation for a list of parameterized controls.
//
// Ref:
// - https://docs.aws.amazon.com/controltower/latest/controlreference/control-parameter-concepts.html
// - https://docs.aws.amazon.com/controltower/latest/controlreference/elective-preventive-controls.html
func testAccControlConfig_parameters(ouName string) string {
	return acctest.ConfigCompose(
		testAccControlConfigBase(ouName),
		`
resource "aws_controltower_control" "test" {
  control_identifier = "arn:${data.aws_partition.current.partition}:controltower:${data.aws_region.current.region}::control/AWS-GR_DISALLOW_CROSS_REGION_NETWORKING"
  target_identifier  = data.aws_organizations_organizational_unit.test.arn

  parameters {
    key   = "ExemptedPrincipalArns"
    value = jsonencode(["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/tf-acctest-example"])
  }
}
`)
}
