package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSEcsDefaultSetting_basic(t *testing.T) {
	//var provider ecs.Setting
	rName := "serviceLongArnFormat"
	resourceName := "aws_ecs_account_setting_default.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsDefaultSettingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsDefaultSettingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "value", "enabled"),
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

func testAccCheckAWSEcsDefaultSettingDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ecsconn

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

		if isAWSErr(err, ecs.ErrCodeResourceNotFoundException, "") {
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

func testAccAWSEcsDefaultSettingConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_account_setting_default" "test" {
	name = %q
    value = "enabled"
}
`, rName)
}
