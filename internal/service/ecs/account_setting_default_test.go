// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSAccountSettingDefault_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(*testing.T){
		"containerInstanceLongARNFormat":  testAccAccountSettingDefault_containerInstanceLongARNFormat,
		"serviceLongARNFormat":            testAccAccountSettingDefault_serviceLongARNFormat,
		"taskLongARNFormat":               testAccAccountSettingDefault_taskLongARNFormat,
		"vpcTrunking":                     testAccAccountSettingDefault_vpcTrunking,
		"containerInsights":               testAccAccountSettingDefault_containerInsights,
		"fargateTaskRetirementWaitPeriod": testAccAccountSettingDefault_fargateTaskRetirementWaitPeriod,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAccountSettingDefault_containerInstanceLongARNFormat(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(awstypes.SettingNameContainerInstanceLongArnFormat)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, settingName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, names.AttrEnabled),
					acctest.MatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexache.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     settingName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAccountSettingDefault_serviceLongARNFormat(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(awstypes.SettingNameServiceLongArnFormat)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, settingName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, names.AttrEnabled),
					acctest.MatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexache.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     settingName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAccountSettingDefault_taskLongARNFormat(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(awstypes.SettingNameTaskLongArnFormat)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, settingName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, names.AttrEnabled),
					acctest.MatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexache.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     settingName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAccountSettingDefault_vpcTrunking(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(awstypes.SettingNameAwsvpcTrunking)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, settingName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, names.AttrEnabled),
					acctest.MatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexache.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     settingName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAccountSettingDefault_containerInsights(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(awstypes.SettingNameContainerInsights)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, settingName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, names.AttrEnabled),
					acctest.MatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexache.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     settingName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAccountSettingDefault_fargateTaskRetirementWaitPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(awstypes.SettingNameFargateTaskRetirementWaitPeriod)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_fargateTaskRetirementWaitPeriod(settingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "fargateTaskRetirementWaitPeriod"),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, "14"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexache.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     "fargateTaskRetirementWaitPeriod",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAccountSettingDefaultDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_account_setting_default" {
				continue
			}

			name := rs.Primary.Attributes[names.AttrName]

			input := &ecs.ListAccountSettingsInput{
				Name:              awstypes.SettingName(name),
				EffectiveSettings: true,
			}

			resp, err := conn.ListAccountSettings(ctx, input)

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return err
			}

			for _, value := range resp.Settings {
				if aws.ToString(value.Value) != "disabled" && aws.ToString(value.Value) != "7" {
					switch awstypes.SettingName(name) {
					case awstypes.SettingNameContainerInstanceLongArnFormat:
						return nil
					case awstypes.SettingNameServiceLongArnFormat:
						return nil
					case awstypes.SettingNameTaskLongArnFormat:
						return nil
					default:
						return fmt.Errorf("[Destroy Error] Account Settings (%s), still enabled", string(value.Name))
					}
				}
			}
		}

		return nil
	}
}

func testAccAccountSettingDefaultConfig_basic(settingName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_account_setting_default" "test" {
  name  = %[1]q
  value = "enabled"
}
`, settingName)
}

func testAccAccountSettingDefaultConfig_fargateTaskRetirementWaitPeriod(settingName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_account_setting_default" "test" {
  name  = %[1]q
  value = "14"
}
`, settingName)
}
