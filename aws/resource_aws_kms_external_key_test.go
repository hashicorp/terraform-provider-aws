package aws

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	awspolicy "github.com/jen20/awspolicyequivalence"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kms/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kms/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func TestAccAWSKmsExternalKey_basic(t *testing.T) {
	var key1 kms.KeyMetadata
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, kms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsExternalKeyConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "kms", regexp.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", "false"),
					resource.TestCheckResourceAttr(resourceName, "deletion_window_in_days", "30"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "expiration_model", ""),
					resource.TestCheckNoResourceAttr(resourceName, "key_material_base64"),
					resource.TestCheckResourceAttr(resourceName, "key_state", "PendingImport"),
					resource.TestCheckResourceAttr(resourceName, "key_usage", "ENCRYPT_DECRYPT"),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`Enable IAM User Permissions`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "valid_to", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_policy_lockout_safety_check",
					"deletion_window_in_days",
					"key_material_base64",
				},
			},
		},
	})
}

func TestAccAWSKmsExternalKey_disappears(t *testing.T) {
	var key1 kms.KeyMetadata
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, kms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsExternalKeyConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key1),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsKmsExternalKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSKmsExternalKey_DeletionWindowInDays(t *testing.T) {
	var key1, key2 kms.KeyMetadata
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, kms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsExternalKeyConfigDeletionWindowInDays(8),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key1),
					resource.TestCheckResourceAttr(resourceName, "deletion_window_in_days", "8"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_policy_lockout_safety_check",
					"deletion_window_in_days",
					"key_material_base64",
				},
			},
			{
				Config: testAccAWSKmsExternalKeyConfigDeletionWindowInDays(7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key2),
					testAccCheckAWSKmsExternalKeyNotRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, "deletion_window_in_days", "7"),
				),
			},
		},
	})
}

func TestAccAWSKmsExternalKey_Description(t *testing.T) {
	var key1, key2 kms.KeyMetadata
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, kms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsExternalKeyConfigDescription("description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key1),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_policy_lockout_safety_check",
					"deletion_window_in_days",
					"key_material_base64",
				},
			},
			{
				Config: testAccAWSKmsExternalKeyConfigDescription("description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key2),
					testAccCheckAWSKmsExternalKeyNotRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccAWSKmsExternalKey_Enabled(t *testing.T) {
	var key1, key2, key3 kms.KeyMetadata
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, kms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsExternalKeyConfigEnabled(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key1),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_policy_lockout_safety_check",
					"deletion_window_in_days",
					"key_material_base64",
				},
			},
			{
				Config: testAccAWSKmsExternalKeyConfigEnabled(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key2),
					testAccCheckAWSKmsExternalKeyNotRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccAWSKmsExternalKeyConfigEnabled(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key3),
					testAccCheckAWSKmsExternalKeyNotRecreated(&key2, &key3),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	})
}

func TestAccAWSKmsExternalKey_KeyMaterialBase64(t *testing.T) {
	var key1, key2 kms.KeyMetadata
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, kms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				// ACCEPTANCE TESTING ONLY -- NEVER EXPOSE YOUR KEY MATERIAL
				Config: testAccAWSKmsExternalKeyConfigKeyMaterialBase64("Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key1),
					resource.TestCheckResourceAttr(resourceName, "key_material_base64", "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_policy_lockout_safety_check",
					"deletion_window_in_days",
					"key_material_base64",
				},
			},
			{
				// ACCEPTANCE TESTING ONLY -- NEVER EXPOSE YOUR KEY MATERIAL
				Config: testAccAWSKmsExternalKeyConfigKeyMaterialBase64("O1zsg06cKRCsZnoT5oizMlwHEtnk0HoOmBLkFtwh2Vw="),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key2),
					testAccCheckAWSKmsExternalKeyRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, "key_material_base64", "O1zsg06cKRCsZnoT5oizMlwHEtnk0HoOmBLkFtwh2Vw="),
				),
			},
		},
	})
}

func TestAccAWSKmsExternalKey_Policy(t *testing.T) {
	var key1, key2 kms.KeyMetadata
	rName := acctest.RandomWithPrefix("tf-acc-test")
	policy1 := `{"Version":"2012-10-17","Id":"kms-tf-1","Statement":[{"Sid":"Enable IAM User Permissions 1","Effect":"Allow","Principal":{"AWS":"*"},"Action":"kms:*","Resource":"*"}]}`
	policy2 := `{"Version":"2012-10-17","Id":"kms-tf-1","Statement":[{"Sid":"Enable IAM User Permissions 2","Effect":"Allow","Principal":{"AWS":"*"},"Action":"kms:*","Resource":"*"}]}`
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, kms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsExternalKeyConfigPolicy(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key1),
					testAccCheckAWSKmsExternalKeyHasPolicy(resourceName, policy1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_policy_lockout_safety_check",
					"deletion_window_in_days",
					"key_material_base64",
				},
			},
			{
				Config: testAccAWSKmsExternalKeyConfigPolicy(rName, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key2),
					testAccCheckAWSKmsExternalKeyNotRecreated(&key1, &key2),
					testAccCheckAWSKmsExternalKeyHasPolicy(resourceName, policy2),
				),
			},
		},
	})
}

func TestAccAWSKmsExternalKey_PolicyBypass(t *testing.T) {
	var key kms.KeyMetadata
	rName := acctest.RandomWithPrefix("tf-acc-test")
	policy := `{"Version":"2012-10-17","Id":"kms-tf-1","Statement":[{"Sid":"Enable IAM User Permissions 1","Effect":"Allow","Principal":{"AWS":"*"},"Action":"kms:*","Resource":"*"}]}`
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, kms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsExternalKeyConfigPolicyBypass(rName, policy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key),
					testAccCheckAWSKmsExternalKeyHasPolicy(resourceName, policy),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_policy_lockout_safety_check",
					"deletion_window_in_days",
					"key_material_base64",
				},
			},
		},
	})
}

func TestAccAWSKmsExternalKey_Tags(t *testing.T) {
	var key1, key2, key3 kms.KeyMetadata
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, kms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsExternalKeyConfigTags1("value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_policy_lockout_safety_check",
					"deletion_window_in_days",
					"key_material_base64",
				},
			},
			{
				Config: testAccAWSKmsExternalKeyConfigTags2("value1updated", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key2),
					testAccCheckAWSKmsExternalKeyNotRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSKmsExternalKeyConfigTags1("value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key3),
					testAccCheckAWSKmsExternalKeyNotRecreated(&key2, &key3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
				),
			},
		},
	})
}

func TestAccAWSKmsExternalKey_ValidTo(t *testing.T) {
	var key1, key2, key3, key4 kms.KeyMetadata
	resourceName := "aws_kms_external_key.test"
	validTo1 := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	validTo2 := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, kms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSKmsExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSKmsExternalKeyConfigValidTo(validTo1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key1),
					resource.TestCheckResourceAttr(resourceName, "expiration_model", "KEY_MATERIAL_EXPIRES"),
					resource.TestCheckResourceAttr(resourceName, "valid_to", validTo1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_policy_lockout_safety_check",
					"deletion_window_in_days",
					"key_material_base64",
				},
			},
			{
				Config: testAccAWSKmsExternalKeyConfigEnabled(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key2),
					testAccCheckAWSKmsExternalKeyNotRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, "expiration_model", "KEY_MATERIAL_DOES_NOT_EXPIRE"),
					resource.TestCheckResourceAttr(resourceName, "valid_to", ""),
				),
			},
			{
				Config: testAccAWSKmsExternalKeyConfigValidTo(validTo1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key3),
					testAccCheckAWSKmsExternalKeyNotRecreated(&key2, &key3),
					resource.TestCheckResourceAttr(resourceName, "expiration_model", "KEY_MATERIAL_EXPIRES"),
					resource.TestCheckResourceAttr(resourceName, "valid_to", validTo1),
				),
			},
			{
				Config: testAccAWSKmsExternalKeyConfigValidTo(validTo2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSKmsExternalKeyExists(resourceName, &key4),
					testAccCheckAWSKmsExternalKeyNotRecreated(&key3, &key4),
					resource.TestCheckResourceAttr(resourceName, "expiration_model", "KEY_MATERIAL_EXPIRES"),
					resource.TestCheckResourceAttr(resourceName, "valid_to", validTo2),
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

		_, err := finder.KeyByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("KMS External Key %s still exists", rs.Primary.ID)
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
			return fmt.Errorf("No KMS External Key ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kmsconn

		outputRaw, err := tfresource.RetryWhenNotFound(waiter.PropagationTimeout, func() (interface{}, error) {
			return finder.KeyByID(conn, rs.Primary.ID)
		})

		if err != nil {
			return err
		}

		*key = *(outputRaw.(*kms.KeyMetadata))

		return nil
	}
}

func testAccCheckAWSKmsExternalKeyNotRecreated(i, j *kms.KeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreationDate).Equal(aws.TimeValue(j.CreationDate)) {
			return fmt.Errorf("KMS External Key recreated")
		}

		return nil
	}
}

func testAccCheckAWSKmsExternalKeyRecreated(i, j *kms.KeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreationDate).Equal(aws.TimeValue(j.CreationDate)) {
			return fmt.Errorf("KMS External Key not recreated")
		}

		return nil
	}
}

func testAccAWSKmsExternalKeyConfig() string {
	return `
resource "aws_kms_external_key" "test" {}
`
}

func testAccAWSKmsExternalKeyConfigDeletionWindowInDays(deletionWindowInDays int) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  deletion_window_in_days = %[1]d
}
`, deletionWindowInDays)
}

func testAccAWSKmsExternalKeyConfigDescription(description string) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
`, description)
}

func testAccAWSKmsExternalKeyConfigEnabled(enabled bool) string {
	return fmt.Sprintf(`
# ACCEPTANCE TESTING ONLY -- NEVER EXPOSE YOUR KEY MATERIAL
resource "aws_kms_external_key" "test" {
  deletion_window_in_days = 7
  enabled                 = %[1]t
  key_material_base64     = "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="
}
`, enabled)
}

func testAccAWSKmsExternalKeyConfigKeyMaterialBase64(keyMaterialBase64 string) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  deletion_window_in_days = 7
  key_material_base64     = %[1]q
}
`, keyMaterialBase64)
}

func testAccAWSKmsExternalKeyConfigPolicy(rName, policy string) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  policy = <<POLICY
%[2]s
POLICY
}
`, rName, policy)
}

func testAccAWSKmsExternalKeyConfigPolicyBypass(rName, policy string) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  bypass_policy_lockout_safety_check = true

  policy = <<POLICY
%[2]s
POLICY
}
`, rName, policy)
}

func testAccAWSKmsExternalKeyConfigTags1(value1 string) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  deletion_window_in_days = 7

  tags = {
    key1 = %[1]q
  }
}
`, value1)
}

func testAccAWSKmsExternalKeyConfigTags2(value1, value2 string) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  deletion_window_in_days = 7

  tags = {
    key1 = %[1]q
    key2 = %[2]q
  }
}
`, value1, value2)
}

func testAccAWSKmsExternalKeyConfigValidTo(validTo string) string {
	return fmt.Sprintf(`
# ACCEPTANCE TESTING ONLY -- NEVER EXPOSE YOUR KEY MATERIAL
resource "aws_kms_external_key" "test" {
  deletion_window_in_days = 7
  key_material_base64     = "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="
  valid_to                = %[1]q
}
`, validTo)
}
