package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func TestAccAWSIAMAccountPasswordPolicy_basic(t *testing.T) {
	var policy iam.GetAccountPasswordPolicyOutput
	resourceName := "aws_iam_account_password_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSIAMAccountPasswordPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMAccountPasswordPolicy,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIAMAccountPasswordPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "minimum_password_length", "8"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSIAMAccountPasswordPolicy_modified,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIAMAccountPasswordPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "minimum_password_length", "7"),
				),
			},
		},
	})
}

func testAccCheckAWSIAMAccountPasswordPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_account_password_policy" {
			continue
		}

		// Try to get policy
		_, err := conn.GetAccountPasswordPolicy(&iam.GetAccountPasswordPolicyInput{})
		if err == nil {
			return fmt.Errorf("still exist.")
		}

		// Verify the error is what we want
		awsErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if awsErr.Code() != "NoSuchEntity" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSIAMAccountPasswordPolicyExists(n string, res *iam.GetAccountPasswordPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		resp, err := conn.GetAccountPasswordPolicy(&iam.GetAccountPasswordPolicyInput{})
		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

const testAccAWSIAMAccountPasswordPolicy = `
resource "aws_iam_account_password_policy" "test" {
  allow_users_to_change_password = true
  minimum_password_length        = 8
  require_numbers                = true
}
`

const testAccAWSIAMAccountPasswordPolicy_modified = `
resource "aws_iam_account_password_policy" "test" {
  allow_users_to_change_password = true
  minimum_password_length        = 7
  require_numbers                = false
  require_symbols                = true
  require_uppercase_characters   = true
}
`
