package securityhub_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
)

func testAccActionTarget_basic(t *testing.T) {
	resourceName := "aws_securityhub_action_target.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckActionTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccActionTargetConfig_identifier("testaction"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActionTargetExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "securityhub", "action/custom/testaction"),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a test custom action"),
					resource.TestCheckResourceAttr(resourceName, "identifier", "testaction"),
					resource.TestCheckResourceAttr(resourceName, "name", "Test action"),
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

func testAccActionTarget_disappears(t *testing.T) {
	resourceName := "aws_securityhub_action_target.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckActionTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccActionTargetConfig_identifier("testaction"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActionTargetExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfsecurityhub.ResourceActionTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccActionTarget_Description(t *testing.T) {
	resourceName := "aws_securityhub_action_target.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckActionTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccActionTargetConfig_description("description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActionTargetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccActionTargetConfig_description("description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActionTargetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func testAccActionTarget_Name(t *testing.T) {
	resourceName := "aws_securityhub_action_target.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, securityhub.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckActionTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccActionTargetConfig_name("name1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActionTargetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "name1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccActionTargetConfig_name("name2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckActionTargetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "name2"),
				),
			},
		},
	})
}

func testAccCheckActionTargetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Security Hub custom action ARN is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn

		action, err := tfsecurityhub.ActionTargetCheckExists(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if action == nil {
			return fmt.Errorf("Security Hub custom action %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckActionTargetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_action_target" {
			continue
		}

		action, err := tfsecurityhub.ActionTargetCheckExists(conn, rs.Primary.ID)

		if tfawserr.ErrMessageContains(err, securityhub.ErrCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
			continue
		}

		if err != nil {
			return err
		}

		if action != nil {
			return fmt.Errorf("Security Hub custom action %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccActionTargetConfig_description(description string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_action_target" "test" {
  depends_on  = [aws_securityhub_account.test]
  description = %[1]q
  identifier  = "testaction"
  name        = "Test action"
}
`, description)
}

func testAccActionTargetConfig_identifier(identifier string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_action_target" "test" {
  depends_on  = [aws_securityhub_account.test]
  description = "This is a test custom action"
  identifier  = %[1]q
  name        = "Test action"
}
`, identifier)
}

func testAccActionTargetConfig_name(name string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_action_target" "test" {
  depends_on  = [aws_securityhub_account.test]
  description = "This is a test custom action"
  identifier  = "testaction"
  name        = %[1]q
}
`, name)
}
