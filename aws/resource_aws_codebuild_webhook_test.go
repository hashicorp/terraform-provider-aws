package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsCodeBuildWebhook_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCodeBuildWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeBuildWebhookConfig_basic(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCodeBuildWebhookExists("aws_codebuild_webhook.test"),
					resource.TestCheckResourceAttrSet("aws_codebuild_webhook.test", "url"),
				),
			},
		},
	})
}

func testAccCheckAwsCodeBuildWebhookDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codebuildconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codebuild_webhook" {
			continue
		}

		resp, err := conn.BatchGetProjects(&codebuild.BatchGetProjectsInput{
			Names: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if err != nil {
			return err
		}

		if len(resp.Projects) == 0 {
			return nil
		}

		project := resp.Projects[0]
		if project.Webhook != nil && project.Webhook.Url != nil {
			return fmt.Errorf("Found CodeBuild Webhook: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckAwsCodeBuildWebhookExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccCodeBuildWebhookConfig_basic(rName string) string {
	return fmt.Sprintf(testAccAWSCodeBuildProjectConfig_basic(rName) + `
resource "aws_codebuild_webhook" "test" {
  name = "${aws_codebuild_project.foo.name}"
}
`)
}
