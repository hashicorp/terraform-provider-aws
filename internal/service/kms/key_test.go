package kms_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	awspolicy "github.com/jen20/awspolicyequivalence"
)

func init() {
	resource.AddTestSweepers("aws_kms_key", &resource.Sweeper{
		Name: "aws_kms_key",
		F:    sweepKeys,
	})
}

func sweepKeys(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).KMSConn

	err = conn.ListKeysPages(&kms.ListKeysInput{Limit: aws.Int64(int64(1000))}, func(out *kms.ListKeysOutput, lastPage bool) bool {
		for _, k := range out.Keys {
			kKeyId := aws.StringValue(k.KeyId)
			kOut, err := conn.DescribeKey(&kms.DescribeKeyInput{
				KeyId: k.KeyId,
			})
			if err != nil {
				log.Printf("Error: Failed to describe key %q: %s", kKeyId, err)
				return false
			}
			if aws.StringValue(kOut.KeyMetadata.KeyManager) == kms.KeyManagerTypeAws {
				// Skip (default) keys which are managed by AWS
				continue
			}
			if aws.StringValue(kOut.KeyMetadata.KeyState) == kms.KeyStatePendingDeletion {
				// Skip keys which are already scheduled for deletion
				continue
			}

			r := tfkms.ResourceKey()
			d := r.Data(nil)
			d.SetId(kKeyId)
			d.Set("key_id", kKeyId)
			d.Set("deletion_window_in_days", "7")
			err = r.Delete(d, client)
			if err != nil {
				log.Printf("Error: Failed to schedule key %q for deletion: %s", kKeyId, err)
				return false
			}
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping KMS Key sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing KMS keys: %w", err)
	}

	return nil
}

func TestAccAWSKmsKey_basic(t *testing.T) {
	var key kms.KeyMetadata
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "customer_master_key_spec", "SYMMETRIC_DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "key_usage", "ENCRYPT_DECRYPT"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days", "bypass_policy_lockout_safety_check"},
			},
		},
	})
}

func TestAccAWSKmsKey_asymmetricKey(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKey_asymmetric(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "customer_master_key_spec", "ECC_NIST_P384"),
					resource.TestCheckResourceAttr(resourceName, "key_usage", "SIGN_VERIFY"),
				),
			},
		},
	})
}

func TestAccAWSKmsKey_disappears(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					acctest.CheckResourceDisappears(acctest.Provider, tfkms.ResourceKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSKmsKey_policy(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"
	expectedPolicyText := `{"Version":"2012-10-17","Id":"kms-tf-1","Statement":[{"Sid":"Enable IAM User Permissions","Effect":"Allow","Principal":{"AWS":"*"},"Action":"kms:*","Resource":"*"}]}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKey_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					testAccCheckKeyHasPolicy(resourceName, expectedPolicyText),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days", "bypass_policy_lockout_safety_check"},
			},
			{
				Config: testAccKey_removedPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
				),
			},
		},
	})
}

func TestAccAWSKmsKey_policyBypass(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccKey_policyBypass(rName, false),
				ExpectError: regexp.MustCompile(`The new key policy will not allow you to update the key policy in the future`),
			},
			{
				Config: testAccKey_policyBypass(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days", "bypass_policy_lockout_safety_check"},
			},
		},
	})
}

func TestAccAWSKmsKey_policyBypassUpdate(t *testing.T) {
	var before, after kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", "false"),
				),
			},
			{
				Config: testAccKey_policyBypass(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", "true"),
				),
			},
		},
	})
}

func TestAccAWSKmsKey_Policy_IamRole(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyIAMRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days", "bypass_policy_lockout_safety_check"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7646
func TestAccAWSKmsKey_Policy_IamServiceLinkedRole(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyPolicyIAMServiceLinkedRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days", "bypass_policy_lockout_safety_check"},
			},
		},
	})
}

func TestAccAWSKmsKey_isEnabled(t *testing.T) {
	var key1, key2, key3 kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKey_enabledRotation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key1),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					testAccCheckKeyIsEnabled(&key1, true),
					resource.TestCheckResourceAttr("aws_kms_key.test", "enable_key_rotation", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days", "bypass_policy_lockout_safety_check"},
			},
			{
				Config: testAccKey_disabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key2),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "false"),
					testAccCheckKeyIsEnabled(&key2, false),
					resource.TestCheckResourceAttr(resourceName, "enable_key_rotation", "false"),
				),
			},
			{
				Config: testAccKey_enabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key3),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					testAccCheckKeyIsEnabled(&key3, true),
					resource.TestCheckResourceAttr(resourceName, "enable_key_rotation", "true"),
				),
			},
		},
	})
}

func TestAccAWSKmsKey_tags(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days", "bypass_policy_lockout_safety_check"},
			},
			{
				Config: testAccKeyTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccKeyTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckKeyHasPolicy(name string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No KMS Key ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn

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

func testAccCheckKeyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kms_key" {
			continue
		}

		_, err := tfkms.FindKeyByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("KMS Key %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckKeyExists(name string, key *kms.KeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No KMS Key ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn

		outputRaw, err := tfresource.RetryWhenNotFound(tfkms.PropagationTimeout, func() (interface{}, error) {
			return tfkms.FindKeyByID(conn, rs.Primary.ID)
		})

		if err != nil {
			return err
		}

		*key = *(outputRaw.(*kms.KeyMetadata))

		return nil
	}
}

func testAccCheckKeyIsEnabled(key *kms.KeyMetadata, isEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if got, want := aws.BoolValue(key.Enabled), isEnabled; got != want {
			return fmt.Errorf("Expected key %q to have is_enabled=%t, given %t", aws.StringValue(key.Arn), want, got)
		}

		return nil
	}
}

func testAccKeyConfig() string {
	return `
resource "aws_kms_key" "test" {}
`
}

func testAccKeyNameConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
`, rName)
}

func testAccKey_asymmetric(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  key_usage                = "SIGN_VERIFY"
  customer_master_key_spec = "ECC_NIST_P384"
}
`, rName)
}

func testAccKey_policy(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
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
}
`, rName)
}

func testAccKey_policyBypass(rName string, bypassFlag bool) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  bypass_policy_lockout_safety_check = %[2]t

  policy = <<-POLICY
    {
      "Version": "2012-10-17",
      "Id": "kms-tf-1",
      "Statement": [
        {
          "Sid": "Enable IAM User Permissions",
          "Effect": "Allow",
          "Principal": {
            "AWS": "${data.aws_caller_identity.current.arn}"
          },
          "Action": [
            "kms:CreateKey",
            "kms:DescribeKey",
            "kms:ScheduleKeyDeletion",
            "kms:Describe*",
            "kms:Get*",
            "kms:List*",
            "kms:TagResource",
            "kms:UntagResource"
          ],
          "Resource": "*"
        }
      ]
    }
  POLICY
}
`, rName, bypassFlag)
}

func testAccKeyPolicyIAMRoleConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  policy = jsonencode({
    Id = "kms-tf-1"
    Statement = [
      {
        Action = "kms:*"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }

        Resource = "*"
        Sid      = "Enable IAM User Permissions"
      },
      {
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncrypt*",
          "kms:GenerateDataKey*",
          "kms:DescribeKey",
        ]
        Effect = "Allow"
        Principal = {
          AWS = [aws_iam_role.test.arn]
        }

        Resource = "*"
        Sid      = "Enable IAM User Permissions"
      },
    ]
    Version = "2012-10-17"
  })
}
`, rName)
}

func testAccKeyPolicyIAMServiceLinkedRoleConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_service_linked_role" "test" {
  aws_service_name = "autoscaling.${data.aws_partition.current.dns_suffix}"
  custom_suffix    = %[1]q
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  policy = jsonencode({
    Id = "kms-tf-1"
    Statement = [
      {
        Action = "kms:*"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }

        Resource = "*"
        Sid      = "Enable IAM User Permissions"
      },
      {
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncrypt*",
          "kms:GenerateDataKey*",
          "kms:DescribeKey",
        ]
        Effect = "Allow"
        Principal = {
          AWS = [aws_iam_service_linked_role.test.arn]
        }

        Resource = "*"
        Sid      = "Enable IAM User Permissions"
      },
    ]
    Version = "2012-10-17"
  })
}
`, rName)
}

func testAccKey_removedPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
`, rName)
}

func testAccKey_enabledRotation(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}
`, rName)
}

func testAccKey_disabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = false
  is_enabled              = false
}
`, rName)
}

func testAccKey_enabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
  is_enabled              = true
}
`, rName)
}

func testAccKeyTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccKeyTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
