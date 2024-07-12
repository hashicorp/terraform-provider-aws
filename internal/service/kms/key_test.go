// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKMSKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "customer_master_key_spec", "SYMMETRIC_DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "key_usage", "ENCRYPT_DECRYPT"),
					resource.TestCheckResourceAttr(resourceName, "multi_region", acctest.CtFalse),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days", "bypass_policy_lockout_safety_check"},
			},
			{
				// Set deletion window to 7 days
				Config: testAccKeyConfig_basicDeletionWindow(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
				),
			},
		},
	})
}

func TestAccKMSKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfkms.ResourceKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKMSKey_multiRegion(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_multiRegion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "multi_region", acctest.CtTrue),
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
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_asymmetric(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "customer_master_key_spec", "ECC_NIST_P384"),
					resource.TestCheckResourceAttr(resourceName, "key_usage", "SIGN_VERIFY"),
				),
			},
		},
	})
}

func TestAccKMSKey_hmacKey(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_hmac(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "customer_master_key_spec", "HMAC_256"),
					resource.TestCheckResourceAttr(resourceName, "key_usage", "GENERATE_VERIFY_MAC"),
				),
			},
		},
	})
}

func TestAccKMSKey_Policy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"
	expectedPolicyText := fmt.Sprintf(`{"Version":"2012-10-17","Id":%[1]q,"Statement":[{"Sid":"Enable IAM User Permissions","Effect":"Allow","Principal":{"AWS":"*"},"Action":"kms:*","Resource":"*"}]}`, rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					testAccCheckKeyHasPolicy(ctx, resourceName, expectedPolicyText),
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
					testAccCheckKeyExists(ctx, resourceName, &key),
				),
			},
		},
	})
}

func TestAccKMSKey_Policy_bypass(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccKeyConfig_policyBypass(rName, false),
				ExpectError: regexache.MustCompile(`The new key policy will not allow you to update the key policy in the future`),
			},
			{
				Config: testAccKeyConfig_policyBypass(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", acctest.CtTrue),
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
	ctx := acctest.Context(t)
	var before, after awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", acctest.CtFalse),
				),
			},
			{
				Config: testAccKeyConfig_policyBypass(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccKMSKey_Policy_iamRole(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_policyIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
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
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
				),
			},
			{
				Config: testAccKeyConfig_policyIAMRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11801
func TestAccKMSKey_Policy_iamRoleOrder(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_policyIAMMultiRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
				),
			},
			{
				Config: testAccKeyConfig_policyIAMMultiRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
				),
				PlanOnly: true,
			},
			{
				Config: testAccKeyConfig_policyIAMMultiRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
				),
				PlanOnly: true,
			},
			{
				Config: testAccKeyConfig_policyIAMMultiRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
				),
				PlanOnly: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7646
func TestAccKMSKey_Policy_iamServiceLinkedRole(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_policyIAMServiceLinkedRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
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
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_policyBooleanCondition(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
				),
			},
		},
	})
}

func TestAccKMSKey_isEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var key1, key2, key3 awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_enabledRotation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key1),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_key_rotation", acctest.CtTrue),
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
					testAccCheckKeyExists(ctx, resourceName, &key2),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_key_rotation", acctest.CtFalse),
				),
			},
			{
				Config: testAccKeyConfig_enabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key3),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_key_rotation", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccKMSKey_rotation(t *testing.T) {
	ctx := acctest.Context(t)
	var key1, key2, key3 awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyConfig_enabledRotation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key1),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_key_rotation", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days", "bypass_policy_lockout_safety_check"},
			},
			{
				Config: testAccKeyConfig_enabledRotationPeriod(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key2),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_key_rotation", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "rotation_period_in_days", "91"),
				),
			},
			{
				Config: testAccKeyConfig_enabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key3),
					resource.TestCheckResourceAttr(resourceName, "is_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "enable_key_rotation", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "rotation_period_in_days", "91"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/26174.
func TestAccKMSKey_tags_IgnoreTags_ModifyOutOfBand(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.KMSServiceID),
		CheckDestroy: testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Key/tags_ignore/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtProviderTags: config.MapVariable(map[string]config.Variable{
						acctest.CtProviderKey1: config.StringVariable(acctest.CtProviderValue1),
					}),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
						acctest.CtKey2: config.StringVariable(acctest.CtValue2),
					}),
					"ignore_tag_keys": config.SetVariable(
						config.StringVariable("ignkey1"),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1:         knownvalue.StringExact(acctest.CtValue1),
						acctest.CtKey2:         knownvalue.StringExact(acctest.CtValue2),
						acctest.CtProviderKey1: knownvalue.StringExact(acctest.CtProviderValue1),
					})),
					expectFullTags(resourceName, knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1:         knownvalue.StringExact(acctest.CtValue1),
						acctest.CtKey2:         knownvalue.StringExact(acctest.CtValue2),
						acctest.CtProviderKey1: knownvalue.StringExact(acctest.CtProviderValue1),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
							acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1:         knownvalue.StringExact(acctest.CtValue1),
							acctest.CtKey2:         knownvalue.StringExact(acctest.CtValue2),
							acctest.CtProviderKey1: knownvalue.StringExact(acctest.CtProviderValue1),
						})),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			// Add an ignored tag.
			{
				PreConfig: func() {
					testAccKeyAddTag(ctx, t, aws.ToString(key.KeyId), "ignkey1", "ignvalue1")
				},
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Key/tags_ignore/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtProviderTags: config.MapVariable(map[string]config.Variable{
						acctest.CtProviderKey1: config.StringVariable(acctest.CtProviderValue1),
					}),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
						acctest.CtKey2: config.StringVariable(acctest.CtValue2),
					}),
					"ignore_tag_keys": config.SetVariable(
						config.StringVariable("ignkey1"),
					),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1:         knownvalue.StringExact(acctest.CtValue1),
						acctest.CtKey2:         knownvalue.StringExact(acctest.CtValue2),
						acctest.CtProviderKey1: knownvalue.StringExact(acctest.CtProviderValue1),
					})),
					expectFullTags(resourceName, knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1:         knownvalue.StringExact(acctest.CtValue1),
						acctest.CtKey2:         knownvalue.StringExact(acctest.CtValue2),
						acctest.CtProviderKey1: knownvalue.StringExact(acctest.CtProviderValue1),
						"ignkey1":              knownvalue.StringExact("ignvalue1"),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			// Modify tags
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Key/tags_ignore/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtProviderTags: config.MapVariable(map[string]config.Variable{
						acctest.CtProviderKey1: config.StringVariable(acctest.CtProviderValue1),
					}),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1Updated),
						"key3":         config.StringVariable("value3"),
					}),
					"ignore_tag_keys": config.SetVariable(
						config.StringVariable("ignkey1"),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &key),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						"key3":         knownvalue.StringExact("value3"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1:         knownvalue.StringExact(acctest.CtValue1Updated),
						"key3":                 knownvalue.StringExact("value3"),
						acctest.CtProviderKey1: knownvalue.StringExact(acctest.CtProviderValue1),
					})),
					expectFullTags(resourceName, knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1:         knownvalue.StringExact(acctest.CtValue1Updated),
						"key3":                 knownvalue.StringExact("value3"),
						acctest.CtProviderKey1: knownvalue.StringExact(acctest.CtProviderValue1),
						"ignkey1":              knownvalue.StringExact("ignvalue1"),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
							"key3":         knownvalue.StringExact("value3"),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1:         knownvalue.StringExact(acctest.CtValue1Updated),
							"key3":                 knownvalue.StringExact("value3"),
							acctest.CtProviderKey1: knownvalue.StringExact(acctest.CtProviderValue1),
						})),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func testAccCheckKeyHasPolicy(ctx context.Context, name string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No KMS Key ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSClient(ctx)

		output, err := tfkms.FindKeyPolicyByTwoPartKey(ctx, conn, rs.Primary.ID, tfkms.PolicyNameDefault)

		if err != nil {
			return err
		}

		actualPolicyText := aws.ToString(output)

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

func testAccCheckKeyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kms_key" {
				continue
			}

			_, err := tfkms.FindKeyByID(ctx, conn, rs.Primary.ID)

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
}

func testAccCheckKeyExists(ctx context.Context, name string, key *awstypes.KeyMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KMSClient(ctx)

		outputRaw, err := tfresource.RetryWhenNotFound(ctx, tfkms.PropagationTimeout, func() (interface{}, error) {
			return tfkms.FindKeyByID(ctx, conn, rs.Primary.ID)
		})

		if err != nil {
			return err
		}

		*key = *(outputRaw.(*awstypes.KeyMetadata))

		return nil
	}
}

func testAccKeyAddTag(ctx context.Context, t *testing.T, identifier, key, value string) {
	t.Helper()

	conn := acctest.Provider.Meta().(*conns.AWSClient).KMSClient(ctx)

	input := &kms.TagResourceInput{
		KeyId: aws.String(identifier),
		Tags: []awstypes.Tag{{
			TagKey:   aws.String(key),
			TagValue: aws.String(value),
		}},
	}

	_, err := conn.TagResource(ctx, input)

	if err != nil {
		t.Fatalf("adding tag: %s", err)
	}
}

func testAccKeyConfig_basic() string {
	return `
resource "aws_kms_key" "test" {}
`
}

func testAccKeyConfig_basicDeletionWindow() string {
	return `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}
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

func testAccKeyConfig_hmac(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7

  key_usage                = "GENERATE_VERIFY_MAC"
  customer_master_key_spec = "HMAC_256"
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
        "AWS" : "*"
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

data "aws_caller_identity" "current" {}

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
          AWS = data.aws_caller_identity.current.arn
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

data "aws_caller_identity" "current" {}

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
          AWS = data.aws_caller_identity.current.arn
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
          AWS = data.aws_caller_identity.current.arn
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

func testAccKeyConfig_enabledRotationPeriod(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
  rotation_period_in_days = "91"
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

// The following tests will eventually be generated

func TestAccKMSKey_tags_IgnoreTags_Overlap_DefaultTag(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.KeyMetadata
	resourceName := "aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.KMSServiceID),
		CheckDestroy: testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Key/tags_ignore/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtProviderTags: config.MapVariable(map[string]config.Variable{
						acctest.CtProviderKey1: config.StringVariable(acctest.CtProviderValue1),
					}),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtResourceKey1: config.StringVariable(acctest.CtResourceValue1),
					}),
					"ignore_tag_keys": config.SetVariable(
						config.StringVariable(acctest.CtProviderKey1),
					),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1),
					})),
					expectFullTags(resourceName, knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtProviderKey1: knownvalue.StringExact(acctest.CtProviderValue1),
						acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1),
						})),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Key/tags_ignore/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtProviderTags: config.MapVariable(map[string]config.Variable{
						acctest.CtProviderKey1: config.StringVariable("providervalue1updated"),
					}),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtResourceKey1: config.StringVariable(acctest.CtResourceValue1),
					}),
					"ignore_tag_keys": config.SetVariable(
						config.StringVariable(acctest.CtProviderKey1),
					),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1),
					})),
					expectFullTags(resourceName, knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtProviderKey1: knownvalue.StringExact(acctest.CtProviderValue1),
						acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1),
						})),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccKMSKey_tags_IgnoreTags_Overlap_ResourceTag(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.KeyMetadata
	resourceName := "aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.KMSServiceID),
		CheckDestroy: testAccCheckKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Key/tags_ignore/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtResourceKey1: config.StringVariable(acctest.CtResourceValue1),
					}),
					"ignore_tag_keys": config.SetVariable(
						config.StringVariable(acctest.CtResourceKey1),
					),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
					expectFullTags(resourceName, knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1),
						})),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New(names.AttrTagsAll)),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
					},
				},
				ExpectNonEmptyPlan: true,
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Key/tags_ignore/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtResourceKey1: config.StringVariable(acctest.CtResourceValue1Updated),
					}),
					"ignore_tag_keys": config.SetVariable(
						config.StringVariable(acctest.CtResourceKey1),
					),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
					expectFullTags(resourceName, knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1),
					})),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1Updated),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1Updated),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtResourceKey1: knownvalue.StringExact(acctest.CtResourceValue1Updated),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
