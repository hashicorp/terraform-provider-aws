package ecs_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccECSAccountSettingDefault_containerInstanceLongArnFormat(t *testing.T) {
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := ecs.SettingNameContainerInstanceLongArnFormat

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAccountSettingDefaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig(settingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", settingName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexp.MustCompile("root")),
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

func TestAccECSAccountSettingDefault_serviceLongArnFormat(t *testing.T) {
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := ecs.SettingNameServiceLongArnFormat

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAccountSettingDefaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig(settingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", settingName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexp.MustCompile("root")),
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

func TestAccECSAccountSettingDefault_taskLongArnFormat(t *testing.T) {
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := ecs.SettingNameTaskLongArnFormat

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAccountSettingDefaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig(settingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", settingName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexp.MustCompile("root")),
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

func TestAccECSAccountSettingDefault_awsvpcTrunking(t *testing.T) {
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := ecs.SettingNameAwsvpcTrunking

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAccountSettingDefaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig(settingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", settingName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexp.MustCompile("root")),
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
	resourceName := "aws_ecs_account_setting_default.test"
	settingName := ecs.SettingNameContainerInsights

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAccountSettingDefaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingDefaultConfig(settingName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", settingName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexp.MustCompile("root")),
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

func testAccCheckAccountSettingDefaultDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ECSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecs_account_setting_default" {
			continue
		}

		name := rs.Primary.Attributes["name"]

		input := &ecs.ListAccountSettingsInput{
			Name:              aws.String(name),
			EffectiveSettings: aws.Bool(true),
		}

		resp, err := conn.ListAccountSettings(input)

		if tfawserr.ErrCodeEquals(err, ecs.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		for _, value := range resp.Settings {
			if aws.StringValue(value.Value) != "disabled" {
				return fmt.Errorf("[Destroy Error] Account Settings (%s) Still enabled", aws.StringValue(value.Name))
			}
		}

	}

	return nil
}

func testAccAccountSettingDefaultConfig(settingName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_account_setting_default" "test" {
  name  = %[1]q
  value = "enabled"
}
`, settingName)
}
