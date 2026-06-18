// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSAccountSettingDefault_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(*testing.T){
		"containerInstanceLongARNFormat":  testAccAccountSettingDefault_containerInstanceLongARNFormat,
		"defaultLogDriverMode":            testAccAccountSettingDefault_defaultLogDriverMode,
		"serviceLongARNFormat":            testAccAccountSettingDefault_serviceLongARNFormat,
		"taskLongARNFormat":               testAccAccountSettingDefault_taskLongARNFormat,
		"vpcTrunking":                     testAccAccountSettingDefault_vpcTrunking,
		"containerInsights":               testAccAccountSettingDefault_containerInsights,
		"fargateTaskRetirementWaitPeriod": testAccAccountSettingDefault_fargateTaskRetirementWaitPeriod,
		"dualStackIPv6":                   testAccAccountSettingDefault_dualStackIPv6,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAccountSettingDefault_containerInstanceLongARNFormat(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(awstypes.SettingNameContainerInstanceLongArnFormat)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName, names.AttrEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSettingDefaultExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, settingName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, names.AttrEnabled),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, "principal_arn", "iam", regexache.MustCompile("root")),
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

func testAccAccountSettingDefault_defaultLogDriverMode(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(awstypes.SettingNameDefaultLogDriverMode)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName, "blocking"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSettingDefaultExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, settingName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, "blocking"),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, "principal_arn", "iam", regexache.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     settingName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName, "non-blocking"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSettingDefaultExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, settingName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, "non-blocking"),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, "principal_arn", "iam", regexache.MustCompile("root")),
				),
			},
		},
	})
}

func testAccAccountSettingDefault_dualStackIPv6(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := "dualStackIPv6"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName, names.AttrEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSettingDefaultExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, settingName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, names.AttrEnabled),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, "principal_arn", "iam", regexache.MustCompile("root")),
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

func testAccAccountSettingDefault_serviceLongARNFormat(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(awstypes.SettingNameServiceLongArnFormat)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName, names.AttrEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSettingDefaultExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, settingName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, names.AttrEnabled),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, "principal_arn", "iam", regexache.MustCompile("root")),
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

func testAccAccountSettingDefault_taskLongARNFormat(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(awstypes.SettingNameTaskLongArnFormat)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName, names.AttrEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSettingDefaultExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, settingName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, names.AttrEnabled),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, "principal_arn", "iam", regexache.MustCompile("root")),
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

func testAccAccountSettingDefault_vpcTrunking(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(awstypes.SettingNameAwsvpcTrunking)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName, names.AttrEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSettingDefaultExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, settingName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, names.AttrEnabled),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, "principal_arn", "iam", regexache.MustCompile("root")),
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

func testAccAccountSettingDefault_containerInsights(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(awstypes.SettingNameContainerInsights)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName, names.AttrEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSettingDefaultExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, settingName),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, names.AttrEnabled),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, "principal_arn", "iam", regexache.MustCompile("root")),
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

func testAccAccountSettingDefault_fargateTaskRetirementWaitPeriod(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(awstypes.SettingNameFargateTaskRetirementWaitPeriod)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_fargateTaskRetirementWaitPeriod(settingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSettingDefaultExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, "fargateTaskRetirementWaitPeriod"),
					resource.TestCheckResourceAttr(resourceName, names.AttrValue, "14"),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, "principal_arn", "iam", regexache.MustCompile("root")),
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

func testAccCheckAccountSettingDefaultExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		settingName := awstypes.SettingName(rs.Primary.Attributes[names.AttrName])
		_, err := tfecs.FindEffectiveAccountSettingByName(ctx, conn, settingName)

		return err
	}
}

func testAccCheckAccountSettingDefaultDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_account_setting_default" {
				continue
			}

			settingName := awstypes.SettingName(rs.Primary.Attributes[names.AttrName])
			output, err := tfecs.FindEffectiveAccountSettingByName(ctx, conn, settingName)

			if retry.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			switch value := aws.ToString(output.Value); settingName {
			case awstypes.SettingNameContainerInstanceLongArnFormat, awstypes.SettingNameServiceLongArnFormat, awstypes.SettingNameTaskLongArnFormat:
				return nil
			case awstypes.SettingNameFargateTaskRetirementWaitPeriod:
				if value == "7" {
					return nil
				}
			case awstypes.SettingNameDefaultLogDriverMode:
				if value == "non-blocking" {
					return nil
				}
			default:
				if value == "disabled" {
					return nil
				}
			}

			return fmt.Errorf("ECS Account Setting Default %s still exists", settingName)
		}

		return nil
	}
}

func testAccAccountSettingDefaultConfig_basic(settingName, value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_account_setting_default" "test" {
  name  = %[1]q
  value = %[2]q
}
`, settingName, value)
}

func testAccAccountSettingDefaultConfig_fargateTaskRetirementWaitPeriod(settingName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_account_setting_default" "test" {
  name  = %[1]q
  value = "14"
}
`, settingName)
}
