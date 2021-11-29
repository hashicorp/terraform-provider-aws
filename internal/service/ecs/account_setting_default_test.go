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

func TestAccECSAccountDefaultSetting_containerInstanceLongArnFormat(t *testing.T) {
	resourceName := "aws_ecs_account_setting_default.test"
	rName := "containerInstanceLongArnFormat"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultSettingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultSettingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexp.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccECSAccountDefaultSetting_serviceLongArnFormat(t *testing.T) {
	resourceName := "aws_ecs_account_setting_default.test"
	rName := "serviceLongArnFormat"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultSettingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultSettingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexp.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccECSAccountDefaultSetting_taskLongArnFormat(t *testing.T) {
	resourceName := "aws_ecs_account_setting_default.test"
	rName := "taskLongArnFormat"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultSettingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultSettingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexp.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccECSAccountDefaultSetting_awsvpcTrunking(t *testing.T) {
	resourceName := "aws_ecs_account_setting_default.test"
	rName := "awsvpcTrunking"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultSettingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultSettingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexp.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccECSAccountDefaultSetting_containerInsights(t *testing.T) {
	resourceName := "aws_ecs_account_setting_default.test"
	rName := "containerInsights"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultSettingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultSettingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "principal_arn", "iam", regexp.MustCompile("root")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDefaultSettingDestroy(s *terraform.State) error {
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

func testAccDefaultSettingConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_account_setting_default" "test" {
	name = %q
    value = "enabled"
}
`, rName)
}
