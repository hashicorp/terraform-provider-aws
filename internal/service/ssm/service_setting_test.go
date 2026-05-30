// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMServiceSetting_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:     testAccServiceSetting_basic,
		"upgradeFromV6_5_0": testAccServiceSetting_upgradeFromV6_5_0,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccServiceSetting_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var setting awstypes.ServiceSetting
	resourceName := "aws_ssm_service_setting.test"
	settingID := "/ssm/parameter-store/high-throughput-enabled"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceSettingDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSettingConfig_basic(acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccServiceSettingExists(ctx, t, resourceName, &setting),
					resource.TestCheckResourceAttr(resourceName, "setting_id", settingID),
					resource.TestCheckResourceAttr(resourceName, "setting_value", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceSettingConfig_basic(acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccServiceSettingExists(ctx, t, resourceName, &setting),
					resource.TestCheckResourceAttr(resourceName, "setting_id", settingID),
					resource.TestCheckResourceAttr(resourceName, "setting_value", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
				),
			},
			{
				Config: testAccServiceSettingConfig_settingIDByARN(acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccServiceSettingExists(ctx, t, resourceName, &setting),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "setting_id", "ssm", "servicesetting"+settingID),
					resource.TestCheckResourceAttr(resourceName, "setting_value", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  acctest.ImportCheckResourceAttr("setting_id", settingID),
				ImportStateVerifyIgnore: []string{
					"setting_id",
				},
			},
			{
				Config: testAccServiceSettingConfig_settingIDByARN(acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccServiceSettingExists(ctx, t, resourceName, &setting),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "setting_id", "ssm", "servicesetting"+settingID),
					resource.TestCheckResourceAttr(resourceName, "setting_value", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrARN),
				),
			},
		},
	})
}

func testAccServiceSetting_upgradeFromV6_5_0(t *testing.T) {
	ctx := acctest.Context(t)
	var setting awstypes.ServiceSetting
	resourceName := "aws_ssm_service_setting.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.SSMServiceID),
		CheckDestroy: testAccCheckServiceSettingDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.5.0",
					},
				},
				Config: testAccServiceSettingConfig_settingIDByARN(acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccServiceSettingExists(ctx, t, resourceName, &setting),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccServiceSettingConfig_settingIDByARN(acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccServiceSettingExists(ctx, t, resourceName, &setting),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func testAccCheckServiceSettingDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SSMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssm_service_setting" {
				continue
			}

			output, err := tfssm.FindServiceSettingByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if aws.ToString(output.Status) == "Default" {
				continue
			}

			return fmt.Errorf("SSM Service Setting %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccServiceSettingExists(ctx context.Context, t *testing.T, n string, v *awstypes.ServiceSetting) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SSMClient(ctx)

		output, err := tfssm.FindServiceSettingByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccServiceSettingConfig_basic(settingValue string) string {
	return fmt.Sprintf(`
resource "aws_ssm_service_setting" "test" {
  setting_id    = "/ssm/parameter-store/high-throughput-enabled"
  setting_value = %[1]q
}
`, settingValue)
}

func testAccServiceSettingConfig_settingIDByARN(settingValue string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_ssm_service_setting" "test" {
  setting_id    = "arn:${data.aws_partition.current.partition}:ssm:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:servicesetting/ssm/parameter-store/high-throughput-enabled"
  setting_value = %[1]q
}
`, settingValue)
}
