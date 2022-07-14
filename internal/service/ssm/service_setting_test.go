package ssm_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccSSMServiceSetting_basic(t *testing.T) {
	var setting ssm.GetServiceSettingOutput
	resourceName := "aws_ssm_service_setting.test"
	awsSession := session.New()
	stssvc := sts.New(awsSession)
	result, _ := stssvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccServiceSettingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceSettingConfig_basic(aws.StringValue(result.Account), aws.StringValue(awsSession.Config.Region), "false"),
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
				Config: testAccServiceSettingConfig_basic(aws.StringValue(result.Account), aws.StringValue(awsSession.Config.Region), "true"),
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
			return fmt.Errorf("SSM Service Setting still customized")
		}
	}

	return nil
}

func testAccServiceSettingExists(n string, res *ssm.GetServiceSettingOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSMConn

		resp, err := conn.GetServiceSetting(&ssm.GetServiceSettingInput{
			SettingId: aws.String(rs.Primary.Attributes["setting_id"]),
		})
		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

func testAccServiceSettingConfig_basic(accountID, region, value string) string {
	return fmt.Sprintf(testSettingTemplate, region, accountID, value)
}

const testSettingTemplate = `
data "aws_partition" "current" {}

resource "aws_ssm_service_setting" "test" {
  setting_id    = "arn:${data.aws_partition.current.partition}:ssm:%s:%s:servicesetting/ssm/parameter-store/high-throughput-enabled"
  setting_value = "%s"
}
`
