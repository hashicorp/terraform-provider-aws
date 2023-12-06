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
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func TestAccECSAccountSettingDefault_containerInstanceLongARNFormat(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(types.SettingNameContainerInstanceLongArnFormat)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", settingName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
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

func TestAccECSAccountSettingDefault_serviceLongARNFormat(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(types.SettingNameServiceLongArnFormat)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", settingName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
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

func TestAccECSAccountSettingDefault_taskLongARNFormat(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(types.SettingNameTaskLongArnFormat)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", settingName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
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

func TestAccECSAccountSettingDefault_vpcTrunking(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(types.SettingNameAwsvpcTrunking)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", settingName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
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

func TestAccECSAccountSettingDefault_containerInsights(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := string(types.SettingNameContainerInsights)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingDefaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig_basic(settingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", settingName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_account_setting_default" {
				continue
			}

			name := rs.Primary.Attributes["name"]

			input := &ecs.ListAccountSettingsInput{
				Name:              types.SettingName(name),
				EffectiveSettings: true,
			}

			resp, err := conn.ListAccountSettings(ctx, input)

			if errs.IsA[*types.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return err
			}

			for _, value := range resp.Settings {
				if aws.ToString(value.Value) != "disabled" {
					switch types.SettingName(name) {
					case types.SettingNameContainerInstanceLongArnFormat:
						return nil
					case types.SettingNameServiceLongArnFormat:
						return nil
					case types.SettingNameTaskLongArnFormat:
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
