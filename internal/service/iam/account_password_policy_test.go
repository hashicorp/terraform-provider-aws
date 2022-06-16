package iam_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccIAMAccountPasswordPolicy_basic(t *testing.T) {
	var policy iam.GetAccountPasswordPolicyOutput
	resourceName := "aws_iam_account_password_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, iam.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccountPasswordPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountPasswordPolicyConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPasswordPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "minimum_password_length", "8"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountPasswordPolicyConfig_modified,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountPasswordPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "minimum_password_length", "7"),
				),
			},
		},
	})
}

func testAccCheckAccountPasswordPolicyDestroy(s *terraform.State) error {
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

func testAccCheckAccountPasswordPolicyExists(n string, res *iam.GetAccountPasswordPolicyOutput) resource.TestCheckFunc {
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

const testAccAccountPasswordPolicyConfig_basic = `
resource "aws_iam_account_password_policy" "test" {
  allow_users_to_change_password = true
  minimum_password_length        = 8
  require_numbers                = true
}
`

const testAccAccountPasswordPolicyConfig_modified = `
resource "aws_iam_account_password_policy" "test" {
  allow_users_to_change_password = true
  minimum_password_length        = 7
  require_numbers                = false
  require_symbols                = true
  require_uppercase_characters   = true
}
`
