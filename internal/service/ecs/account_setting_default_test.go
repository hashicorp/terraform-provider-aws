// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSAccountSettingDefault_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(*testing.T){
		"containerInstanceLongARNFormat": testAccAccountSettingDefault_containerInstanceLongARNFormat,
		"serviceLongARNFormat":           testAccAccountSettingDefault_serviceLongARNFormat,
		"taskLongARNFormat":              testAccAccountSettingDefault_taskLongARNFormat,
		"vpcTrunking":                    testAccAccountSettingDefault_vpcTrunking,
		"containerInsights":              testAccAccountSettingDefault_containerInsights,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccAccountSettingDefault_containerInstanceLongARNFormat(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := ecs.SettingNameContainerInstanceLongArnFormat

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
	settingName := ecs.SettingNameServiceLongArnFormat

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
	settingName := ecs.SettingNameTaskLongArnFormat

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
	settingName := ecs.SettingNameAwsvpcTrunking

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
	settingName := ecs.SettingNameContainerInsights

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

func testAccCheckAccountSettingDefaultDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_account_setting_default" {
				continue
			}

			name := rs.Primary.Attributes[names.AttrName]

			input := &ecs.ListAccountSettingsInput{
				Name:              aws.String(name),
				EffectiveSettings: aws.Bool(true),
			}

			resp, err := conn.ListAccountSettingsWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, ecs.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return err
			}

			for _, value := range resp.Settings {
				if aws.StringValue(value.Value) != "disabled" {
					switch name {
					case ecs.SettingNameContainerInstanceLongArnFormat:
						return nil
					case ecs.SettingNameServiceLongArnFormat:
						return nil
					case ecs.SettingNameTaskLongArnFormat:
						return nil
					default:
						return fmt.Errorf("[Destroy Error] Account Settings (%s), still enabled", aws.StringValue(value.Name))
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
