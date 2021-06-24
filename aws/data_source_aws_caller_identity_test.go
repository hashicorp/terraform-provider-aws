package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestSourceARN(t *testing.T) {
	testCases := []struct {
		Name     string
		ARN      string
		Expected string
	}{
		{
			Name:     "not an ARN",
			ARN:      "abcd",
			Expected: "abcd",
		},
		{
			Name:     "regular ARN",
			ARN:      "arn:aws:iam::111122223333:role/role_name",
			Expected: "arn:aws:iam::111122223333:role/role_name",
		},
		{
			Name:     "assumed role ARN",
			ARN:      "arn:aws:sts::444433332222:assumed-role/something_something-admin/sessionIDNotPartOfRoleARN",
			Expected: "arn:aws:iam::444433332222:role/something_something-admin",
		},
		{
			Name:     "'assumed-role' part of ARN resource",
			ARN:      "arn:aws:iam::444433332222:user/assumed-role-but-not-really",
			Expected: "arn:aws:iam::444433332222:user/assumed-role-but-not-really",
		},
		{
			Name:     "user ARN",
			ARN:      "arn:aws:iam::123456789012:user/Bob",
			Expected: "arn:aws:iam::123456789012:user/Bob",
		},
		{
			Name:     "assumed role from AWS example",
			ARN:      "arn:aws:sts::123456789012:assumed-role/example-role/AWSCLI-Session",
			Expected: "arn:aws:iam::123456789012:role/example-role",
		},
		{
			Name:     "multiple slashes in resource", // not sure this is even valid
			ARN:      "arn:aws:sts::123456789012:assumed-role/example-role/also-part-of-role-or-no/AWSCLI-Session",
			Expected: "arn:aws:iam::123456789012:role/example-role/also-part-of-role-or-no",
		},
		{
			Name:     "not an sts ARN",
			ARN:      "arn:aws:iam::123456789012:assumed-role/example-role/AWSCLI-Session",
			Expected: "arn:aws:iam::123456789012:assumed-role/example-role/AWSCLI-Session",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got := sourceARN(testCase.ARN)

			if got != testCase.Expected {
				t.Errorf("for %s: got %s, expected %s", testCase.Name, got, testCase.Expected)
			}
		})
	}
}

func TestAccAWSCallerIdentity_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, sts.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsCallerIdentityConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCallerIdentityAccountId("data.aws_caller_identity.current"),
				),
			},
		},
	})
}

func testAccCheckAwsCallerIdentityAccountId(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find AccountID resource: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Account Id resource ID not set.")
		}

		expected := testAccProvider.Meta().(*AWSClient).accountid
		if rs.Primary.Attributes["account_id"] != expected {
			return fmt.Errorf("Incorrect Account ID: expected %q, got %q", expected, rs.Primary.Attributes["account_id"])
		}

		if rs.Primary.Attributes["user_id"] == "" {
			return fmt.Errorf("UserID expected to not be nil")
		}

		if rs.Primary.Attributes["arn"] == "" {
			return fmt.Errorf("ARN expected to not be nil")
		}

		return nil
	}
}

const testAccCheckAwsCallerIdentityConfig_basic = `
data "aws_caller_identity" "current" {}
`
