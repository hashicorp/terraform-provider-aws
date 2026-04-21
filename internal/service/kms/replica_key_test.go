// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfkms "github.com/hashicorp/terraform-provider-aws/internal/service/kms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccKMSReplicaKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_kms_replica_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicaKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kms", regexache.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "deletion_window_in_days", "30"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_rotation_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_spec", "SYMMETRIC_DEFAULT"),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`Enable IAM User Permissions`)),
					resource.TestCheckResourceAttrPair(resourceName, "primary_key_arn", primaryKeyResourceName, names.AttrARN),
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
		},
	})
}

func TestAccKMSReplicaKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_replica_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicaKeyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key),
					acctest.CheckSDKResourceDisappears(ctx, t, tfkms.ResourceReplicaKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKMSReplicaKey_descriptionAndEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName3 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName4 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_replica_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicaKeyConfig_descriptionAndEnabled(rName1, rName2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days", "bypass_policy_lockout_safety_check"},
			},
			{
				Config: testAccReplicaKeyConfig_descriptionAndEnabled(rName1, rName3, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName3),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
				),
			},
			{
				Config: testAccReplicaKeyConfig_descriptionAndEnabled(rName1, rName4, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName4),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccKMSReplicaKey_policy(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_replica_key.test"
	policy1 := `{"Id":"kms-tf-1","Statement":[{"Action":"kms:*","Effect":"Allow","Principal":{"AWS":"*"},"Resource":"*","Sid":"Enable IAM User Permissions 1"}],"Version":"2012-10-17"}`
	policy2 := `{"Id":"kms-tf-1","Statement":[{"Action":"kms:*","Effect":"Allow","Principal":{"AWS":"*"},"Resource":"*","Sid":"Enable IAM User Permissions 2"}],"Version":"2012-10-17"}`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicaKeyConfig_policy(rName, policy1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", acctest.CtFalse),
					testAccCheckKeyHasPolicy(ctx, t, resourceName, policy1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deletion_window_in_days", "bypass_policy_lockout_safety_check"},
			},
			{
				Config: testAccReplicaKeyConfig_policy(rName, policy2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", acctest.CtTrue),
					testAccCheckExternalKeyHasPolicy(ctx, t, resourceName, policy2),
				),
			},
		},
	})
}

func TestAccKMSReplicaKey_twoReplicas(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_kms_replica_key.test1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicaKeyConfig_two(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key),
				),
			},
		},
	})
}

func TestAccKMSReplicaKey_multipleProviders(t *testing.T) {
	ctx := acctest.Context(t)
	var key awstypes.KeyMetadata
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_kms_replica_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.KMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckKeyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicaKeyConfig_multipleProviders(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyExists(ctx, t, resourceName, &key),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "kms", regexache.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "deletion_window_in_days", "30"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "key_rotation_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "key_spec", "SYMMETRIC_DEFAULT"),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`Enable IAM User Permissions`)),
					resource.TestCheckResourceAttrPair(resourceName, "primary_key_arn", primaryKeyResourceName, names.AttrARN),
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
		},
	})
}

func testAccCheckReplicaKeyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return testAccCheckKeyDestroy(ctx, t)
}

func testAccCheckReplicaKeyExists(ctx context.Context, t *testing.T, name string, key *awstypes.KeyMetadata) resource.TestCheckFunc {
	return testAccCheckKeyExists(ctx, t, name, key)
}

func testAccReplicaKeyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  region = %[2]q

  description             = %[1]q
  multi_region            = true
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_replica_key" "test" {
  primary_key_arn = aws_kms_key.test.arn
}
`, rName, acctest.AlternateRegion())
}

func testAccReplicaKeyConfig_descriptionAndEnabled(rName, description string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  region = %[4]q

  description  = %[1]q
  multi_region = true

  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_replica_key" "test" {
  description     = %[2]q
  enabled         = %[3]t
  primary_key_arn = aws_kms_key.test.arn

  deletion_window_in_days = 7
}
`, rName, description, enabled, acctest.AlternateRegion())
}

func testAccReplicaKeyConfig_policy(rName, policy string, bypassLockoutCheck bool) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  region = %[4]q

  description  = %[1]q
  multi_region = true

  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_replica_key" "test" {
  description     = %[1]q
  primary_key_arn = aws_kms_key.test.arn

  deletion_window_in_days = 7

  bypass_policy_lockout_safety_check = %[3]t

  policy = %[2]q
}
`, rName, policy, bypassLockoutCheck, acctest.AlternateRegion())
}

func testAccReplicaKeyConfig_two(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  region = %[2]q

  description  = %[1]q
  multi_region = true

  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_kms_replica_key" "test1" {
  description     = %[1]q
  primary_key_arn = aws_kms_key.test.arn

  deletion_window_in_days = 7
}

resource "aws_kms_replica_key" "test2" {
  region = %[3]q

  description     = %[1]q
  primary_key_arn = aws_kms_key.test.arn

  deletion_window_in_days = 7
}
`, rName, acctest.AlternateRegion(), acctest.ThirdRegion())
}

func testAccReplicaKeyConfig_multipleProviders(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateRegionProvider(), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  provider = awsalternate

  description  = %[1]q
  multi_region = true
}

resource "aws_kms_replica_key" "test" {
  primary_key_arn = aws_kms_key.test.arn
}
`, rName))
}
