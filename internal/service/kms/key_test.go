package kms_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccKMSKey_basic(t *testing.T) {
	var key kms.KeyMetadata
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "customer_master_key_spec", "SYMMETRIC_DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "key_usage", "ENCRYPT_DECRYPT"),
					resource.TestCheckResourceAttr(resourceName, "multi_region", "false"),
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

func TestAccKMSKey_disappears(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					acctest.CheckResourceDisappears(acctest.Provider, tfkms.ResourceKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKMSKey_multiRegion(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_multiRegion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "multi_region", "true"),
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

func TestAccKMSKey_asymmetricKey(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_asymmetric(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "customer_master_key_spec", "ECC_NIST_P384"),
					resource.TestCheckResourceAttr(resourceName, "key_usage", "SIGN_VERIFY"),
				),
			},
		},
	})
}

func TestAccKMSKey_Policy_basic(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"
	expectedPolicyText := fmt.Sprintf(`{"Version":"2012-10-17","Id":%[1]q,"Statement":[{"Sid":"Enable IAM User Permissions","Effect":"Allow","Principal":{"AWS":"*"},"Action":"kms:*","Resource":"*"}]}`, rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_policy(rName),
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
				Config: testAccKeyConfig_removedPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
				),
			},
		},
	})
}

func TestAccKMSKey_Policy_bypass(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccKeyConfig_policyBypass(rName, false),
				ExpectError: regexp.MustCompile(`The new key policy will not allow you to update the key policy in the future`),
			},
			{
				Config: testAccKeyConfig_policyBypass(rName, true),
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

func TestAccKMSKey_Policy_bypassUpdate(t *testing.T) {
	var before, after kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", "false"),
				),
			},
			{
				Config: testAccKeyConfig_policyBypass(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", "true"),
				),
			},
		},
	})
}

func TestAccKMSKey_Policy_iamRole(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_policyIAMRole(rName),
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

func TestAccKMSKey_Policy_iamRoleUpdate(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
				),
			},
			{
				Config: testAccKeyConfig_policyIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11801
func TestAccKMSKey_Policy_iamRoleOrder(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_policyIAMMultiRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
				),
			},
			{
				Config: testAccKeyConfig_policyIAMMultiRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
				),
				PlanOnly: true,
			},
			{
				Config: testAccKeyConfig_policyIAMMultiRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
				),
				PlanOnly: true,
			},
			{
				Config: testAccKeyConfig_policyIAMMultiRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
				),
				PlanOnly: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7646
func TestAccKMSKey_Policy_iamServiceLinkedRole(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_policyIAMServiceLinkedRole(rName),
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

func TestAccKMSKey_Policy_booleanCondition(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_policyBooleanCondition(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
				),
			},
		},
	})
}

func TestAccKMSKey_isEnabled(t *testing.T) {
	var key1, key2, key3 kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_enabledRotation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key1),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_key_rotation", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days", "bypass_policy_lockout_safety_check"},
			},
			{
				Config: testAccKeyConfig_disabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key2),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_key_rotation", "false"),
				),
			},
			{
				Config: testAccKeyConfig_enabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key3),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "enable_key_rotation", "true"),
				),
			},
		},
	})
}

func TestAccKMSKey_tags(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_tags1(rName, "key1", "value1"),
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
				Config: testAccKeyConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccKeyConfig_tags1(rName, "key2", "value2"),
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

		actualPolicyText := aws.StringValue(out.Policy)

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

func testAccKeyConfig_basic() string {
	return `
resource "aws_kms_key" "test" {}
`
}

func testAccKeyConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
`, rName)
}

func testAccKeyConfig_multiRegion(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = false
  is_enabled              = true
  multi_region            = true
}
`, rName)
}

func testAccKeyConfig_asymmetric(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  key_usage                = "SIGN_VERIFY"
  customer_master_key_spec = "ECC_NIST_P384"
}
`, rName)
}

func testAccKeyConfig_policy(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  policy = jsonencode({
    Id = %[1]q
    Statement = [{
      Sid    = "Enable IAM User Permissions"
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action   = "kms:*"
      Resource = "*"
    }]
    Version = "2012-10-17"
  })
}
`, rName)
}

func testAccKeyConfig_policyBypass(rName string, bypassFlag bool) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  bypass_policy_lockout_safety_check = %[2]t

  policy = jsonencode({
    Id = %[1]q
    Statement = [
      {
        Action = [
          "kms:CreateKey",
          "kms:DescribeKey",
          "kms:ScheduleKeyDeletion",
          "kms:Describe*",
          "kms:Get*",
          "kms:List*",
          "kms:TagResource",
          "kms:UntagResource",
        ]
        Effect = "Allow"
        Principal = {
          AWS = data.aws_caller_identity.current.arn
        }
        Resource = "*"
        Sid      = "Enable IAM User Permissions"
      },
    ]
    Version = "2012-10-17"
  })
}
`, rName, bypassFlag)
}

func testAccKeyConfig_policyIAMRole(rName string) string {
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
    Id = %[1]q
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
          AWS = aws_iam_role.test.arn
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

func testAccKeyConfig_policyIAMMultiRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test1" {
  name = "%[1]s-sultan"

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

resource "aws_iam_role" "test2" {
  name = "%[1]s-shepard"

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

resource "aws_iam_role" "test3" {
  name = "%[1]s-tritonal"

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

resource "aws_iam_role" "test4" {
  name = "%[1]s-artlec"

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

resource "aws_iam_role" "test5" {
  name = "%[1]s-cazzette"

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

data "aws_iam_policy_document" "test" {
  policy_id = %[1]q
  statement {
    actions = [
      "kms:*",
    ]
    effect = "Allow"
    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
    resources = ["*"]
  }

  statement {
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey",
    ]
    effect = "Allow"
    principals {
      identifiers = [
        aws_iam_role.test2.arn,
        aws_iam_role.test1.arn,
        aws_iam_role.test4.arn,
        aws_iam_role.test3.arn,
        aws_iam_role.test5.arn,
      ]
      type = "AWS"
    }
    resources = ["*"]
  }
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  policy = data.aws_iam_policy_document.test.json
}
`, rName)
}

func testAccKeyConfig_policyIAMServiceLinkedRole(rName string) string {
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
    Id = %[1]q
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
          AWS = aws_iam_service_linked_role.test.arn
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

func testAccKeyConfig_policyBooleanCondition(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  policy = jsonencode({
    Id = %[1]q
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
          AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        }

        Resource = "*"
        Sid      = "Enable IAM User Permissions"

        Condition = {
          Bool = {
            "kms:GrantIsForAWSResource" = true
          }
        }
      },
    ]
    Version = "2012-10-17"
  })
}
`, rName)
}

func testAccKeyConfig_removedPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
`, rName)
}

func testAccKeyConfig_enabledRotation(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}
`, rName)
}

func testAccKeyConfig_disabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = false
  is_enabled              = false
}
`, rName)
}

func testAccKeyConfig_enabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
  is_enabled              = true
}
`, rName)
}

func testAccKeyConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccKeyConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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
