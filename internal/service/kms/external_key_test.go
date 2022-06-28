package kms_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

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

func TestAccKMSExternalKey_basic(t *testing.T) {
	var key kms.KeyMetadata
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kms", regexp.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", "false"),
					resource.TestCheckResourceAttr(resourceName, "deletion_window_in_days", "30"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "expiration_model", ""),
					resource.TestCheckNoResourceAttr(resourceName, "key_material_base64"),
					resource.TestCheckResourceAttr(resourceName, "key_state", "PendingImport"),
					resource.TestCheckResourceAttr(resourceName, "key_usage", "ENCRYPT_DECRYPT"),
					resource.TestCheckResourceAttr(resourceName, "multi_region", "false"),
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

func TestAccKMSExternalKey_disappears(t *testing.T) {
	var key kms.KeyMetadata
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key),
					acctest.CheckResourceDisappears(acctest.Provider, tfkms.ResourceExternalKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKMSExternalKey_multiRegion(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_multiRegion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "multi_region", "true"),
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

func TestAccKMSExternalKey_deletionWindowInDays(t *testing.T) {
	var key1, key2 kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_deletionWindowInDays(rName, 8),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key1),
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
				Config: testAccExternalKeyConfig_deletionWindowInDays(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key2),
					testAccCheckExternalKeyNotRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, "deletion_window_in_days", "7"),
				),
			},
		},
	})
}

func TestAccKMSExternalKey_description(t *testing.T) {
	var key1, key2 kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_description(rName + "-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key1),
					resource.TestCheckResourceAttr(resourceName, "description", rName+"-1"),
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
				Config: testAccExternalKeyConfig_description(rName + "-2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key2),
					testAccCheckExternalKeyNotRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, "description", rName+"-2"),
				),
			},
		},
	})
}

func TestAccKMSExternalKey_enabled(t *testing.T) {
	var key1, key2, key3 kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key1),
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
				Config: testAccExternalKeyConfig_enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key2),
					testAccCheckExternalKeyNotRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccExternalKeyConfig_enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key3),
					testAccCheckExternalKeyNotRecreated(&key2, &key3),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	})
}

func TestAccKMSExternalKey_keyMaterialBase64(t *testing.T) {
	var key1, key2 kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				// ACCEPTANCE TESTING ONLY -- NEVER EXPOSE YOUR KEY MATERIAL
				Config: testAccExternalKeyConfig_materialBase64(rName, "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key1),
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
				Config: testAccExternalKeyConfig_materialBase64(rName, "O1zsg06cKRCsZnoT5oizMlwHEtnk0HoOmBLkFtwh2Vw="),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key2),
					testAccCheckExternalKeyRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, "key_material_base64", "O1zsg06cKRCsZnoT5oizMlwHEtnk0HoOmBLkFtwh2Vw="),
				),
			},
		},
	})
}

func TestAccKMSExternalKey_policy(t *testing.T) {
	var key1, key2 kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policy1 := `{"Id":"kms-tf-1","Statement":[{"Action":"kms:*","Effect":"Allow","Principal":{"AWS":"*"},"Resource":"*","Sid":"Enable IAM User Permissions 1"}],"Version":"2012-10-17"}`
	policy2 := `{"Id":"kms-tf-1","Statement":[{"Action":"kms:*","Effect":"Allow","Principal":{"AWS":"*"},"Resource":"*","Sid":"Enable IAM User Permissions 2"}],"Version":"2012-10-17"}`
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_policy(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key1),
					testAccCheckExternalKeyHasPolicy(resourceName, policy1),
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
				Config: testAccExternalKeyConfig_policy(rName, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key2),
					testAccCheckExternalKeyNotRecreated(&key1, &key2),
					testAccCheckExternalKeyHasPolicy(resourceName, policy2),
				),
			},
		},
	})
}

func TestAccKMSExternalKey_policyBypass(t *testing.T) {
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policy := `{"Id":"kms-tf-1","Statement":[{"Action":"kms:*","Effect":"Allow","Principal":{"AWS":"*"},"Resource":"*","Sid":"Enable IAM User Permissions 1"}],"Version":"2012-10-17"}`
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_policyBypass(rName, policy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key),
					testAccCheckExternalKeyHasPolicy(resourceName, policy),
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

func TestAccKMSExternalKey_tags(t *testing.T) {
	var key1, key2, key3 kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key1),
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
				Config: testAccExternalKeyConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key2),
					testAccCheckExternalKeyNotRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccExternalKeyConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key3),
					testAccCheckExternalKeyNotRecreated(&key2, &key3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccKMSExternalKey_validTo(t *testing.T) {
	var key1, key2, key3, key4 kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_external_key.test"
	validTo1 := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	validTo2 := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckExternalKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_validTo(rName, validTo1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key1),
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
				Config: testAccExternalKeyConfig_enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key2),
					testAccCheckExternalKeyNotRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, "expiration_model", "KEY_MATERIAL_DOES_NOT_EXPIRE"),
					resource.TestCheckResourceAttr(resourceName, "valid_to", ""),
				),
			},
			{
				Config: testAccExternalKeyConfig_validTo(rName, validTo1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key3),
					testAccCheckExternalKeyNotRecreated(&key2, &key3),
					resource.TestCheckResourceAttr(resourceName, "expiration_model", "KEY_MATERIAL_EXPIRES"),
					resource.TestCheckResourceAttr(resourceName, "valid_to", validTo1),
				),
			},
			{
				Config: testAccExternalKeyConfig_validTo(rName, validTo2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(resourceName, &key4),
					testAccCheckExternalKeyNotRecreated(&key3, &key4),
					resource.TestCheckResourceAttr(resourceName, "expiration_model", "KEY_MATERIAL_EXPIRES"),
					resource.TestCheckResourceAttr(resourceName, "valid_to", validTo2),
				),
			},
		},
	})
}

func testAccCheckExternalKeyHasPolicy(name string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No KMS External Key ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn

		output, err := tfkms.FindKeyPolicyByKeyIDAndPolicyName(conn, rs.Primary.ID, tfkms.PolicyNameDefault)

		if err != nil {
			return err
		}

		actualPolicyText := aws.StringValue(output)

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

func testAccCheckExternalKeyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KMSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kms_external_key" {
			continue
		}

		_, err := tfkms.FindKeyByID(conn, rs.Primary.ID)

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

func testAccCheckExternalKeyExists(name string, key *kms.KeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No KMS External Key ID is set")
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

func testAccCheckExternalKeyNotRecreated(i, j *kms.KeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreationDate).Equal(aws.TimeValue(j.CreationDate)) {
			return fmt.Errorf("KMS External Key recreated")
		}

		return nil
	}
}

func testAccCheckExternalKeyRecreated(i, j *kms.KeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreationDate).Equal(aws.TimeValue(j.CreationDate)) {
			return fmt.Errorf("KMS External Key not recreated")
		}

		return nil
	}
}

func testAccExternalKeyConfig_basic() string {
	return `
resource "aws_kms_external_key" "test" {}
`
}

func testAccExternalKeyConfig_multiRegion(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  description  = %[1]q
  multi_region = true

  deletion_window_in_days = 7
}
`, rName)
}

func testAccExternalKeyConfig_deletionWindowInDays(rName string, deletionWindowInDays int) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  description             = %[1]q
  deletion_window_in_days = %[2]d
}
`, rName, deletionWindowInDays)
}

func testAccExternalKeyConfig_description(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}
`, rName)
}

func testAccExternalKeyConfig_enabled(rName string, enabled bool) string {
	return fmt.Sprintf(`
# ACCEPTANCE TESTING ONLY -- NEVER EXPOSE YOUR KEY MATERIAL
resource "aws_kms_external_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enabled                 = %[2]t
  key_material_base64     = "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="
}
`, rName, enabled)
}

func testAccExternalKeyConfig_materialBase64(rName, keyMaterialBase64 string) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  key_material_base64     = %[2]q
}
`, rName, keyMaterialBase64)
}

func testAccExternalKeyConfig_policy(rName, policy string) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  policy = %[2]q
}
`, rName, policy)
}

func testAccExternalKeyConfig_policyBypass(rName, policy string) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  bypass_policy_lockout_safety_check = true

  policy = %[2]q
}
`, rName, policy)
}

func testAccExternalKeyConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccExternalKeyConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccExternalKeyConfig_validTo(rName, validTo string) string {
	return fmt.Sprintf(`
# ACCEPTANCE TESTING ONLY -- NEVER EXPOSE YOUR KEY MATERIAL
resource "aws_kms_external_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  key_material_base64     = "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="
  valid_to                = %[2]q
}
`, rName, validTo)
}
