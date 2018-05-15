package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCodeBuildWebhook_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeBuildWebhookDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSCodeBuildWebhook,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeBuildWebhookExists("aws_codebuild_webhook.foo"),
				),
			},
		},
	})
}

func testAccCheckAWSCodeBuildWebhookDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codebuildconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codebuild_webhook" {
			continue
		}

		_, err := conn.CreateWebhook(&codebuild.CreateWebhookInput{
			ProjectName: aws.String(rs.Primary.Attributes["project_name"]),
		})

		if err != nil {
			return fmt.Errorf("Webhook should not exist: %s", err.Error())
		}

		_, err = conn.DeleteWebhook(&codebuild.DeleteWebhookInput{
			ProjectName: aws.String(rs.Primary.Attributes["project_name"]),
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAWSCodeBuildWebhookExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

var testAccAWSCodeBuildWebhook = `
resource "aws_codebuild_webhook" "foo" {
	project_name = "foo"
}`
