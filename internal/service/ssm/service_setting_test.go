package ssm_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMServiceSetting_basic(t *testing.T) {
	var setting ssm.ServiceSetting
	resourceName := "aws_ssm_service_setting.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ssm.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccServiceSettingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSettingConfig_basic("false"),
				Check: resource.ComposeTestCheckFunc(
					testAccServiceSettingExists(resourceName, &setting),
					resource.TestCheckResourceAttr(resourceName, "setting_value", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceSettingConfig_basic("true"),
				Check: resource.ComposeTestCheckFunc(
					testAccServiceSettingExists(resourceName, &setting),
					resource.TestCheckResourceAttr(resourceName, "setting_value", "true"),
				),
			},
		},
	})
}

func testAccServiceSettingDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_service_setting" {
			continue
		}

		output, err := conn.GetServiceSetting(&ssm.GetServiceSettingInput{
			SettingId: aws.String(rs.Primary.Attributes["setting_id"]),
		})
		_, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if output.ServiceSetting.Status != aws.String("default") {
			return names.Error(names.SSM, names.ErrActionCheckingDestroyed, tfssm.ResNameServiceSetting, rs.Primary.Attributes["setting_id"], err)
		}
	}

	return nil
}

func testAccServiceSettingExists(n string, res *ssm.ServiceSetting) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn

		output, err := tfssm.FindServiceSettingByARN(conn, rs.Primary.Attributes["setting_id"])

		if err != nil {
			return names.Error(names.SSM, names.ErrActionReading, tfssm.ResNameServiceSetting, rs.Primary.Attributes["setting_id"], err)
		}

		*res = *output

		return nil
	}
}

func testAccServiceSettingConfig_basic(settingValue string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_ssm_service_setting" "test" {
  setting_id    = "arn:${data.aws_partition.current.partition}:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:servicesetting/ssm/parameter-store/high-throughput-enabled"
  setting_value = %[1]q
}
`, settingValue)
}
