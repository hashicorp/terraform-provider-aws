package aws

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/vault/helper/pgpkeys"
)

func TestAccAWSAccessKey_basic(t *testing.T) {
	var conf iam.AccessKeyMetadata
	resourceName := "aws_iam_access_key.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAccessKeyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAccessKeyExists(resourceName, &conf),
					testAccCheckAWSAccessKeyAttributes(&conf, "Active"),
					testAccCheckResourceAttrRfc3339(resourceName, "create_date"),
					resource.TestCheckResourceAttrSet(resourceName, "secret"),
					resource.TestCheckNoResourceAttr(resourceName, "encrypted_secret"),
					resource.TestCheckNoResourceAttr(resourceName, "key_fingerprint"),
					resource.TestCheckResourceAttrSet(resourceName, "ses_smtp_password_v4"),
					resource.TestCheckNoResourceAttr(resourceName, "encrypted_ses_smtp_password_v4"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"encrypted_secret", "key_fingerprint", "pgp_key", "secret", "ses_smtp_password_v4", "encrypted_ses_smtp_password_v4"},
			},
		},
	})
}

func TestAccAWSAccessKey_encrypted(t *testing.T) {
	var conf iam.AccessKeyMetadata
	resourceName := "aws_iam_access_key.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAccessKeyConfig_encrypted(rName, testPubKey1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAccessKeyExists(resourceName, &conf),
					testAccCheckAWSAccessKeyAttributes(&conf, "Active"),
					testDecryptSecretKeyAndTest(resourceName, testPrivKey1),
					resource.TestCheckNoResourceAttr(resourceName, "secret"),
					resource.TestCheckResourceAttrSet(resourceName, "encrypted_secret"),
					resource.TestCheckResourceAttrSet(resourceName, "key_fingerprint"),
					resource.TestCheckNoResourceAttr(resourceName, "ses_smtp_password_v4"),
					resource.TestCheckResourceAttrSet(resourceName, "encrypted_ses_smtp_password_v4"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"encrypted_secret", "key_fingerprint", "pgp_key", "secret", "ses_smtp_password_v4", "encrypted_ses_smtp_password_v4"},
			},
		},
	})
}

func TestAccAWSAccessKey_status(t *testing.T) {
	var conf iam.AccessKeyMetadata
	resourceName := "aws_iam_access_key.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAccessKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAccessKeyConfig_Status(rName, iam.StatusTypeInactive),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAccessKeyExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "status", iam.StatusTypeInactive),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"encrypted_secret", "key_fingerprint", "pgp_key", "secret", "ses_smtp_password_v4", "encrypted_ses_smtp_password_v4"},
			},
			{
				Config: testAccAWSAccessKeyConfig_Status(rName, iam.StatusTypeActive),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAccessKeyExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "status", iam.StatusTypeActive),
				),
			},
			{
				Config: testAccAWSAccessKeyConfig_Status(rName, iam.StatusTypeInactive),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAccessKeyExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "status", iam.StatusTypeInactive),
				),
			},
		},
	})
}

func testAccCheckAWSAccessKeyDestroy(s *terraform.State) error {
	iamconn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_access_key" {
			continue
		}

		// Try to get access key
		resp, err := iamconn.ListAccessKeys(&iam.ListAccessKeysInput{
			UserName: aws.String(rs.Primary.ID),
		})
		if err == nil {
			if len(resp.AccessKeyMetadata) > 0 {
				return fmt.Errorf("still exist.")
			}
			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "NoSuchEntity" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSAccessKeyExists(n string, res *iam.AccessKeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Role name is set")
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn
		name := rs.Primary.Attributes["user"]

		resp, err := iamconn.ListAccessKeys(&iam.ListAccessKeysInput{
			UserName: aws.String(name),
		})
		if err != nil {
			return err
		}

		if len(resp.AccessKeyMetadata) != 1 ||
			*resp.AccessKeyMetadata[0].UserName != name {
			return fmt.Errorf("User not found not found")
		}

		*res = *resp.AccessKeyMetadata[0]

		return nil
	}
}

func testAccCheckAWSAccessKeyAttributes(accessKeyMetadata *iam.AccessKeyMetadata, status string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.Contains(*accessKeyMetadata.UserName, "tf-acc-test") {
			return fmt.Errorf("Bad username: %s", *accessKeyMetadata.UserName)
		}

		if *accessKeyMetadata.Status != status {
			return fmt.Errorf("Bad status: %s", *accessKeyMetadata.Status)
		}

		return nil
	}
}

func testDecryptSecretKeyAndTest(nAccessKey, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		keyResource, ok := s.RootModule().Resources[nAccessKey]
		if !ok {
			return fmt.Errorf("Not found: %s", nAccessKey)
		}

		secret, ok := keyResource.Primary.Attributes["encrypted_secret"]
		if !ok {
			return errors.New("No secret in state")
		}

		password, ok := keyResource.Primary.Attributes["encrypted_ses_smtp_password_v4"]
		if !ok {
			return errors.New("No password in state")
		}

		// We can't verify that the decrypted secret or password is correct, because we don't
		// have it. We can verify that decrypting it does not error
		_, err := pgpkeys.DecryptBytes(secret, key)
		if err != nil {
			return fmt.Errorf("Error decrypting secret: %s", err)
		}
		_, err = pgpkeys.DecryptBytes(password, key)
		if err != nil {
			return fmt.Errorf("Error decrypting password: %s", err)
		}

		return nil
	}
}

func testAccAWSAccessKeyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_access_key" "test" {
  user = aws_iam_user.test.name
}
`, rName)
}

func testAccAWSAccessKeyConfig_encrypted(rName, key string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_access_key" "test" {
  user = aws_iam_user.test.name

  pgp_key = <<EOF
%[2]s
EOF
}
`, rName, key)
}

func testAccAWSAccessKeyConfig_Status(rName string, status string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_access_key" "test" {
  user   = aws_iam_user.test.name
  status = %[2]q
}
`, rName, status)
}

func TestSesSmtpPasswordFromSecretKeySigV4(t *testing.T) {
	cases := []struct {
		Region   string
		Input    string
		Expected string
	}{
		{endpoints.EuCentral1RegionID, "some+secret+key", "BMXhUYlu5Z3gSXVQORxlVa7XPaz91aGWdfHxvkOZdWZ2"},
		{endpoints.EuCentral1RegionID, "another+secret+key", "BBbphbrQmrKMx42d1N6+C7VINYEBGI5v9VsZeTxwskfh"},
		{endpoints.UsWest1RegionID, "some+secret+key", "BH+jbMzper5WwlwUar9E1ySBqHa9whi0GPo+sJ0mVYJj"},
		{endpoints.UsWest1RegionID, "another+secret+key", "BKVmjjMDFk/qqw8EROW99bjCS65PF8WKvK5bSr4Y6EqF"},
	}

	for _, tc := range cases {
		actual, err := sesSmtpPasswordFromSecretKeySigV4(&tc.Input, tc.Region)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if actual != tc.Expected {
			t.Fatalf("%q: expected %q, got %q", tc.Input, tc.Expected, actual)
		}
	}
}
