package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jen20/awspolicyequivalence"
)

func TestAccAWSKmsExternalKey_basic(t *testing.T) {
	var keyBefore, keyAfter kms.KeyMetadata

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsExternalKey,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists("aws_kms_external_key.foo", &keyBefore),
				),
			},
			{
				Config: testAccAWSKmsExternalKey_removedPolicy,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists("aws_kms_external_key.foo", &keyAfter),
				),
			},
		},
	})
}

func TestAccAWSKmsExternalKey_disappears(t *testing.T) {
	var key kms.KeyMetadata

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsExternalKey,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists("aws_kms_external_key.foo", &key),
				),
			},
			{
				Config:             testAccAWSKmsExternalKey_other_region,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSKmsExternalKey_policy(t *testing.T) {
	var key kms.KeyMetadata
	expectedPolicyText := `{"Version":"2012-10-17","Id":"kms-tf-1","Statement":[{"Sid":"Enable IAM User Permissions","Effect":"Allow","Principal":{"AWS":"*"},"Action":"kms:*","Resource":"*"}]}`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsExternalKey,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists("aws_kms_external_key.foo", &key),
					testAccCheckAWSKmsExternalKeyHasPolicy("aws_kms_external_key.foo", expectedPolicyText),
				),
			},
		},
	})
}

func TestAccAWSKmsExternalKey_isEnabled_isAlwaysFalseWhenKeyStateIsPendingImport(t *testing.T) {
	var key1, key2, key3 kms.KeyMetadata

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsExternalKey_bar_removedPolicy,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists("aws_kms_external_key.bar", &key1),
					resource.TestCheckResourceAttr("aws_kms_external_key.bar", "is_enabled", "false"),
					testAccCheckAWSKmsExternalKeyIsEnabled(&key1, false),
				),
			},
			{
				Config: testAccAWSKmsExternalKey_disabled,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists("aws_kms_external_key.bar", &key2),
					resource.TestCheckResourceAttr("aws_kms_external_key.bar", "is_enabled", "false"),
					testAccCheckAWSKmsExternalKeyIsEnabled(&key2, false),
				),
			},
			{
				Config: testAccAWSKmsExternalKey_enabled,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists("aws_kms_external_key.bar", &key3),
					resource.TestCheckResourceAttr("aws_kms_external_key.bar", "is_enabled", "false"),
					testAccCheckAWSKmsExternalKeyIsEnabled(&key3, false),
				),
			},
		},
	})
}

func TestAccAWSKmsExternalKey_tags(t *testing.T) {
	var keyBefore kms.KeyMetadata

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsExternalKey_tags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists("aws_kms_external_key.foo", &keyBefore),
					resource.TestCheckResourceAttr("aws_kms_external_key.foo", "tags.%", "2"),
				),
			},
		},
	})
}

func testAccCheckAWSKmsExternalKeyHasPolicy(name string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No KMS Key ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kmsconn

		out, err := conn.GetKeyPolicy(&kms.GetKeyPolicyInput{
			KeyId:      aws.String(rs.Primary.ID),
			PolicyName: aws.String("default"),
		})
		if err != nil {
			return err
		}

		actualPolicyText := *out.Policy

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckAWSKmsExternalKeyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).kmsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kms_external_key" {
			continue
		}

		out, err := conn.DescribeKey(&kms.DescribeKeyInput{
			KeyId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *out.KeyMetadata.KeyState == "PendingDeletion" {
			return nil
		}

		return fmt.Errorf("KMS key still exists:\n%#v", out.KeyMetadata)
	}

	return nil
}

func testAccCheckAWSKmsExternalKeyExists(name string, key *kms.KeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No KMS Key ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kmsconn

		o, err := retryOnAwsCode("NotFoundException", func() (interface{}, error) {
			return conn.DescribeKey(&kms.DescribeKeyInput{
				KeyId: aws.String(rs.Primary.ID),
			})
		})
		if err != nil {
			return err
		}
		out := o.(*kms.DescribeKeyOutput)

		*key = *out.KeyMetadata

		return nil
	}
}

func testAccCheckAWSKmsExternalKeyIsEnabled(key *kms.KeyMetadata, isEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *key.Enabled != isEnabled {
			return fmt.Errorf("Expected key %q to have is_enabled=%t, given %t",
				*key.Arn, isEnabled, *key.Enabled)
		}

		return nil
	}
}

var kmsExternalTimestamp = time.Now().Format(time.RFC1123)
var testAccAWSKmsExternalKey = fmt.Sprintf(`
resource "aws_kms_external_key" "foo" {
    description = "Terraform acc test %s"
    deletion_window_in_days = 7
    policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}`, kmsExternalTimestamp)

var testAccAWSKmsExternalKey_other_region = fmt.Sprintf(`
provider "aws" { 
	region = "us-east-1"
}
resource "aws_kms_external_key" "foo" {
    description = "Terraform acc test %s"
    deletion_window_in_days = 7
    policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}`, kmsExternalTimestamp)

var testAccAWSKmsExternalKey_removedPolicy = fmt.Sprintf(`
resource "aws_kms_external_key" "foo" {
    description = "Terraform acc test %s"
    deletion_window_in_days = 7
}`, kmsExternalTimestamp)

var testAccAWSKmsExternalKey_bar_removedPolicy = fmt.Sprintf(`
resource "aws_kms_external_key" "bar" {
    description = "Terraform acc test is_enabled %s"
    deletion_window_in_days = 7
}`, kmsExternalTimestamp)

var testAccAWSKmsExternalKey_disabled = fmt.Sprintf(`
resource "aws_kms_external_key" "bar" {
    description = "Terraform acc test is_enabled %s"
    deletion_window_in_days = 7
    is_enabled = false
}`, kmsExternalTimestamp)
var testAccAWSKmsExternalKey_enabled = fmt.Sprintf(`
resource "aws_kms_external_key" "bar" {
    description = "Terraform acc test is_enabled %s"
    deletion_window_in_days = 7
    is_enabled = true
}`, kmsExternalTimestamp)

var testAccAWSKmsExternalKey_tags = fmt.Sprintf(`
resource "aws_kms_external_key" "foo" {
    description = "Terraform acc test %s"
	tags {
		Key1 = "Value One"
		Description = "Very interesting"
	}
}`, kmsExternalTimestamp)
