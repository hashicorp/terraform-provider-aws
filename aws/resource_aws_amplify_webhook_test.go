package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/amplify"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSAmplifyWebhook_basic(t *testing.T) {
	var webhook amplify.Webhook
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_webhook.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAmplifyWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyWebhookConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAmplifyWebhookExists(resourceName, &webhook),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile("^arn:[^:]+:amplify:[^:]+:[^:]+:apps/[^/]+/webhooks/[^/]+$")),
					resource.TestMatchResourceAttr(resourceName, "url", regexp.MustCompile("^https://webhooks.amplify.")),
					resource.TestCheckResourceAttr(resourceName, "branch_name", "master"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyWebhookConfigAll(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "triggermaster"),
				),
			},
		},
	})
}

func testAccCheckAWSAmplifyWebhookExists(resourceName string, v *amplify.Webhook) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).amplifyconn

		output, err := conn.GetWebhook(&amplify.GetWebhookInput{
			WebhookId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if output == nil || output.Webhook == nil {
			return fmt.Errorf("Amplify Webhook (%s) not found", rs.Primary.ID)
		}

		*v = *output.Webhook

		return nil
	}
}

func testAccCheckAWSAmplifyWebhookDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_amplify_webhook" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).amplifyconn

		_, err := conn.GetWebhook(&amplify.GetWebhookInput{
			WebhookId: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, amplify.ErrCodeNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccAWSAmplifyWebhookConfig_Required(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = "%s"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = "master"
}

resource "aws_amplify_webhook" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = aws_amplify_branch.test.branch_name
}
`, rName)
}

func testAccAWSAmplifyWebhookConfigAll(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = "%s"
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = "master"
}

resource "aws_amplify_webhook" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = aws_amplify_branch.test.branch_name
  description = "triggermaster"
}
`, rName)
}
