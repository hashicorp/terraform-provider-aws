// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kms_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKMSExternalKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	resourceName := "aws_kms_external_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExternalKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kms", regexache.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "deletion_window_in_days", "30"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "expiration_model", ""),
					resource.TestCheckNoResourceAttr(resourceName, "key_material_base64"),
					resource.TestCheckResourceAttr(resourceName, "key_state", "PendingImport"),
					resource.TestCheckResourceAttr(resourceName, "key_usage", "ENCRYPT_DECRYPT"),
					resource.TestCheckResourceAttr(resourceName, "multi_region", acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`Enable IAM User Permissions`)),
					resource.TestCheckResourceAttr(resourceName, "valid_to", ""),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_policy_lockout_safety_check",
					"deletion_window_in_days",
				},
			},
		},
	})
}

func TestAccKMSExternalKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	resourceName := "aws_kms_external_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExternalKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key),
					acctest.CheckSDKResourceDisappears(ctx, t, tfkms.ResourceExternalKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKMSExternalKey_multiRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_external_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExternalKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_multiRegion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "multi_region", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_policy_lockout_safety_check",
					"deletion_window_in_days",
				},
			},
		},
	})
}

func TestAccKMSExternalKey_deletionWindowInDays(t *testing.T) {
	ctx := acctest.Context(t)
	var key1, key2 awstypes.KeyMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_external_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExternalKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_deletionWindowInDays(rName, 8),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key1),
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
				},
			},
			{
				Config: testAccExternalKeyConfig_deletionWindowInDays(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key2),
					testAccCheckExternalKeyNotRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, "deletion_window_in_days", "7"),
				),
			},
		},
	})
}

func TestAccKMSExternalKey_description(t *testing.T) {
	ctx := acctest.Context(t)
	var key1, key2 awstypes.KeyMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_external_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExternalKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_description(rName + "-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName+"-1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_policy_lockout_safety_check",
					"deletion_window_in_days",
				},
			},
			{
				Config: testAccExternalKeyConfig_description(rName + "-2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key2),
					testAccCheckExternalKeyNotRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName+"-2"),
				),
			},
		},
	})
}

func TestAccKMSExternalKey_enabled(t *testing.T) {
	ctx := acctest.Context(t)
	var key1, key2, key3 awstypes.KeyMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_external_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExternalKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key1),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
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
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key2),
					testAccCheckExternalKeyNotRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
			{
				Config: testAccExternalKeyConfig_enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key3),
					testAccCheckExternalKeyNotRecreated(&key2, &key3),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccKMSExternalKey_keyMaterialBase64(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_external_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExternalKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// ACCEPTANCE TESTING ONLY -- NEVER EXPOSE YOUR KEY MATERIAL
				Config: testAccExternalKeyConfig_materialBase64(rName, "Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY="),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_material_base64"), knownvalue.StringExact("Wblj06fduthWggmsT0cLVoIMOkeLbc2kVfMud77i/JY=")),
				},
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
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "key_material_base64", "O1zsg06cKRCsZnoT5oizMlwHEtnk0HoOmBLkFtwh2Vw="),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_material_base64"), knownvalue.StringExact("O1zsg06cKRCsZnoT5oizMlwHEtnk0HoOmBLkFtwh2Vw=")),
				},
			},
		},
	})
}

func TestAccKMSExternalKey_policy(t *testing.T) {
	ctx := acctest.Context(t)
	var key1, key2 awstypes.KeyMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	policy1 := `{"Id":"kms-tf-1","Statement":[{"Action":"kms:*","Effect":"Allow","Principal":{"AWS":"*"},"Resource":"*","Sid":"Enable IAM User Permissions 1"}],"Version":"2012-10-17"}`
	policy2 := `{"Id":"kms-tf-1","Statement":[{"Action":"kms:*","Effect":"Allow","Principal":{"AWS":"*"},"Resource":"*","Sid":"Enable IAM User Permissions 2"}],"Version":"2012-10-17"}`
	resourceName := "aws_kms_external_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExternalKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_policy(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key1),
					testAccCheckExternalKeyHasPolicy(ctx, t, resourceName, policy1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_policy_lockout_safety_check",
					"deletion_window_in_days",
				},
			},
			{
				Config: testAccExternalKeyConfig_policy(rName, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key2),
					testAccCheckExternalKeyNotRecreated(&key1, &key2),
					testAccCheckExternalKeyHasPolicy(ctx, t, resourceName, policy2),
				),
			},
		},
	})
}

func TestAccKMSExternalKey_policyBypass(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	policy := `{"Id":"kms-tf-1","Statement":[{"Action":"kms:*","Effect":"Allow","Principal":{"AWS":"*"},"Resource":"*","Sid":"Enable IAM User Permissions 1"}],"Version":"2012-10-17"}`
	resourceName := "aws_kms_external_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExternalKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_policyBypass(rName, policy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key),
					testAccCheckExternalKeyHasPolicy(ctx, t, resourceName, policy),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"bypass_policy_lockout_safety_check",
					"deletion_window_in_days",
				},
			},
		},
	})
}

func TestAccKMSExternalKey_validTo(t *testing.T) {
	ctx := acctest.Context(t)
	var key1, key2, key3, key4 awstypes.KeyMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_external_key.test"
	validTo1 := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	validTo2 := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExternalKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_validTo(rName, validTo1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key1),
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
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key2),
					testAccCheckExternalKeyNotRecreated(&key1, &key2),
					resource.TestCheckResourceAttr(resourceName, "expiration_model", "KEY_MATERIAL_DOES_NOT_EXPIRE"),
					resource.TestCheckResourceAttr(resourceName, "valid_to", ""),
				),
			},
			{
				Config: testAccExternalKeyConfig_validTo(rName, validTo1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key3),
					testAccCheckExternalKeyNotRecreated(&key2, &key3),
					resource.TestCheckResourceAttr(resourceName, "expiration_model", "KEY_MATERIAL_EXPIRES"),
					resource.TestCheckResourceAttr(resourceName, "valid_to", validTo1),
				),
			},
			{
				Config: testAccExternalKeyConfig_validTo(rName, validTo2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key4),
					testAccCheckExternalKeyNotRecreated(&key3, &key4),
					resource.TestCheckResourceAttr(resourceName, "expiration_model", "KEY_MATERIAL_EXPIRES"),
					resource.TestCheckResourceAttr(resourceName, "valid_to", validTo2),
				),
			},
		},
	})
}

func TestAccKMSExternalKey_keyUsage(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_external_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExternalKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_keyUsage(rName, "RSA_4096", "ENCRYPT_DECRYPT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_usage"), tfknownvalue.StringExact(awstypes.KeyUsageTypeEncryptDecrypt)),
				},
			},
			{
				Config: testAccExternalKeyConfig_keyUsage(rName, "RSA_4096", "SIGN_VERIFY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_usage"), tfknownvalue.StringExact(awstypes.KeyUsageTypeSignVerify)),
				},
			},
		}})
}

func TestAccKMSExternalKey_keySpec(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_external_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExternalKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExternalKeyConfig_keySpec(rName, "RSA_2048"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_spec"), tfknownvalue.StringExact(awstypes.KeySpecRsa2048)),
				},
			},
			{
				Config: testAccExternalKeyConfig_keySpec(rName, "SYMMETRIC_DEFAULT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExternalKeyExists(ctx, t, resourceName, &key),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_spec"), tfknownvalue.StringExact(awstypes.KeySpecSymmetricDefault)),
				},
			},
		}})
}

func testAccCheckExternalKeyHasPolicy(ctx context.Context, t *testing.T, name string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.ProviderMeta(ctx, t).KMSClient(ctx)

		output, err := tfkms.FindKeyPolicyByTwoPartKey(ctx, conn, rs.Primary.ID, tfkms.PolicyNameDefault)

		if err != nil {
			return err
		}

		actualPolicyText := aws.ToString(output)

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %w", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckExternalKeyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).KMSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kms_external_key" {
				continue
			}

			_, err := tfkms.FindKeyByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("KMS External Key %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckExternalKeyExists(ctx context.Context, t *testing.T, n string, v *awstypes.KeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).KMSClient(ctx)

		output, err := tfresource.RetryWhenNotFound(ctx, tfkms.PropagationTimeout, func(ctx context.Context) (*awstypes.KeyMetadata, error) {
			return tfkms.FindKeyByID(ctx, conn, rs.Primary.ID)
		})

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckExternalKeyNotRecreated(i, j *awstypes.KeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.ToTime(i.CreationDate).Equal(aws.ToTime(j.CreationDate)) {
			return fmt.Errorf("KMS External Key recreated")
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

func testAccExternalKeyConfig_keyUsage(rName, spec, usage string) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  key_spec                = %[2]q
  key_usage               = %[3]q
}
`, rName, spec, usage)
}

func testAccExternalKeyConfig_keySpec(rName, spec string) string {
	return fmt.Sprintf(`
resource "aws_kms_external_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  key_spec                = %[2]q
  key_usage               = "ENCRYPT_DECRYPT"
}
`, rName, spec)
}
