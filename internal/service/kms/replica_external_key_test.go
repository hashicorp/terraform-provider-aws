package kms_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccKMSReplicaExternalKey_basic(t *testing.T) {
	var providers []*schema.Provider
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryKeyResourceName := "aws_kms_external_key.test"
	resourceName := "aws_kms_replica_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicaExternalKeyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kms", regexp.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", "false"),
					resource.TestCheckResourceAttr(resourceName, "deletion_window_in_days", "30"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "expiration_model", ""),
					resource.TestCheckNoResourceAttr(resourceName, "key_material_base64"),
					resource.TestCheckResourceAttr(resourceName, "key_state", "PendingImport"),
					resource.TestCheckResourceAttr(resourceName, "key_usage", "ENCRYPT_DECRYPT"),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(`Enable IAM User Permissions`)),
					resource.TestCheckResourceAttrPair(resourceName, "primary_key_arn", primaryKeyResourceName, "arn"),
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

func TestAccKMSReplicaExternalKey_DescriptionAndEnabled(t *testing.T) {
	var providers []*schema.Provider
	var key kms.KeyMetadata
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName4 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_replica_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicaExternalKeyDescriptionAndEnabledConfig(rName1, rName2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "description", rName2),
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
				Config: testAccReplicaExternalKeyDescriptionAndEnabledConfig(rName1, rName3, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "description", rName3),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccReplicaExternalKeyDescriptionAndEnabledConfig(rName1, rName4, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "description", rName4),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	})
}

func TestAccKMSReplicaExternalKey_Tags(t *testing.T) {
	var providers []*schema.Provider
	var key kms.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_replica_external_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, kms.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicaExternalKeyTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
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
				Config: testAccReplicaExternalKeyTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccReplicaExternalKeyTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccReplicaExternalKeyConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateRegionProvider(), fmt.Sprintf(`
# ACCEPTANCE TESTING ONLY -- NEVER EXPOSE YOUR KEY MATERIAL
resource "aws_kms_external_key" "test" {
  provider = awsalternate

  description  = %[1]q
  multi_region = true
  enabled      = true

  key_material_base64 = "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="
}

resource "aws_kms_replica_external_key" "test" {
  primary_key_arn = aws_kms_external_key.test.arn
}
`, rName))
}

func testAccReplicaExternalKeyDescriptionAndEnabledConfig(rName, description string, enabled bool) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateRegionProvider(), fmt.Sprintf(`
# ACCEPTANCE TESTING ONLY -- NEVER EXPOSE YOUR KEY MATERIAL
resource "aws_kms_external_key" "test" {
  provider = awsalternate

  description  = %[1]q
  multi_region = true
  enabled      = true

  key_material_base64 = "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="

  deletion_window_in_days = 7
}

resource "aws_kms_replica_external_key" "test" {
  description     = %[2]q
  enabled         = %[3]t
  primary_key_arn = aws_kms_external_key.test.arn

  key_material_base64 = "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="

  deletion_window_in_days = 7
}
`, rName, description, enabled))
}

func testAccReplicaExternalKeyTags1Config(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateRegionProvider(), fmt.Sprintf(`
# ACCEPTANCE TESTING ONLY -- NEVER EXPOSE YOUR KEY MATERIAL
resource "aws_kms_external_key" "test" {
  provider = awsalternate

  description  = %[1]q
  multi_region = true
  enabled      = true

  tags = {
    Name = %[1]q
  }

  key_material_base64 = "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="

  deletion_window_in_days = 7
}

resource "aws_kms_replica_external_key" "test" {
  description     = %[2]q
  enabled         = true
  primary_key_arn = aws_kms_external_key.test.arn

  key_material_base64 = "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="

  tags = {
    %[2]q = %[3]q
  }

  deletion_window_in_days = 7
}
`, rName, tagKey1, tagValue1))
}

func testAccReplicaExternalKeyTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateRegionProvider(), fmt.Sprintf(`
# ACCEPTANCE TESTING ONLY -- NEVER EXPOSE YOUR KEY MATERIAL
resource "aws_kms_external_key" "test" {
  provider = awsalternate

  description  = %[1]q
  multi_region = true
  enabled      = true

  tags = {
    Name = %[1]q
  }

  key_material_base64 = "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="

  deletion_window_in_days = 7
}

resource "aws_kms_replica_external_key" "test" {
  description     = %[2]q
  enabled         = true
  primary_key_arn = aws_kms_external_key.test.arn

  key_material_base64 = "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  deletion_window_in_days = 7
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
