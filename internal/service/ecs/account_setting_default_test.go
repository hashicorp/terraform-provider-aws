// Copyright (c) HashiCorp, Inc.
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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
					testAccCheckAccountSettingDefaultExists(ctx, resourceName),
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
					testAccCheckAccountSettingDefaultExists(ctx, resourceName),
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
					testAccCheckAccountSettingDefaultExists(ctx, resourceName),
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
					testAccCheckAccountSettingDefaultExists(ctx, resourceName),
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
					testAccCheckAccountSettingDefaultExists(ctx, resourceName),
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
					testAccCheckAccountSettingDefaultExists(ctx, resourceName),
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

func testAccCheckAccountSettingDefaultExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSClient(ctx)

		settingName := awstypes.SettingName(rs.Primary.Attributes[names.AttrName])
		_, err := tfecs.FindEffectiveAccountSettingByName(ctx, conn, settingName)

		return err
	}
}

func testAccCheckAccountSettingDefaultDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_account_setting_default" {
				continue
			}

			settingName := awstypes.SettingName(rs.Primary.Attributes[names.AttrName])
			output, err := tfecs.FindEffectiveAccountSettingByName(ctx, conn, settingName)

			if tfresource.NotFound(err) {
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
