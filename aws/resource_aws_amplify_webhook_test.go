package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/amplify"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/amplify/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccAWSAmplifyWebhook_basic(t *testing.T) {
	var webhook amplify.Webhook
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_webhook.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   acctest.ErrorCheck(t, amplify.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSAmplifyWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyWebhookConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAmplifyWebhookExists(resourceName, &webhook),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "amplify", regexp.MustCompile(`apps/.+/webhooks/.+`)),
					resource.TestCheckResourceAttr(resourceName, "branch_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestMatchResourceAttr(resourceName, "url", regexp.MustCompile(fmt.Sprintf(`^https://webhooks.amplify.%s.%s/.+$`, acctest.Region(), acctest.PartitionDNSSuffix()))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAWSAmplifyWebhook_disappears(t *testing.T) {
	var webhook amplify.Webhook
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_webhook.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   acctest.ErrorCheck(t, amplify.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSAmplifyWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyWebhookConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAmplifyWebhookExists(resourceName, &webhook),
					acctest.CheckResourceDisappears(acctest.Provider, resourceAwsAmplifyWebhook(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSAmplifyWebhook_update(t *testing.T) {
	var webhook amplify.Webhook
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_amplify_webhook.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSAmplify(t) },
		ErrorCheck:   acctest.ErrorCheck(t, amplify.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSAmplifyWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAmplifyWebhookConfigDescription(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAmplifyWebhookExists(resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_name", fmt.Sprintf("%s-1", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", "testdescription1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAmplifyWebhookConfigDescriptionUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAmplifyWebhookExists(resourceName, &webhook),
					resource.TestCheckResourceAttr(resourceName, "branch_name", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", "testdescription2"),
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

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Amplify Webhook ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyConn

		webhook, err := finder.WebhookByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *webhook

		return nil
	}
}

func testAccCheckAWSAmplifyWebhookDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AmplifyConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_amplify_webhook" {
			continue
		}

		_, err := finder.WebhookByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Amplify Webhook %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSAmplifyWebhookConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = %[1]q
}

resource "aws_amplify_webhook" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = aws_amplify_branch.test.branch_name
}
`, rName)
}

func testAccAWSAmplifyWebhookConfigDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test1" {
  app_id      = aws_amplify_app.test.id
  branch_name = "%[1]s-1"
}

resource "aws_amplify_branch" "test2" {
  app_id      = aws_amplify_app.test.id
  branch_name = "%[1]s-2"
}

resource "aws_amplify_webhook" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = aws_amplify_branch.test1.branch_name
  description = "testdescription1"
}
`, rName)
}

func testAccAWSAmplifyWebhookConfigDescriptionUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_amplify_app" "test" {
  name = %[1]q
}

resource "aws_amplify_branch" "test1" {
  app_id      = aws_amplify_app.test.id
  branch_name = "%[1]s-1"
}

resource "aws_amplify_branch" "test2" {
  app_id      = aws_amplify_app.test.id
  branch_name = "%[1]s-2"
}

resource "aws_amplify_webhook" "test" {
  app_id      = aws_amplify_app.test.id
  branch_name = aws_amplify_branch.test2.branch_name
  description = "testdescription2"
}
`, rName)
}
