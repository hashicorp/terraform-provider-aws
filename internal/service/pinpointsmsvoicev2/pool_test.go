// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfpinpointsmsvoicev2 "github.com/hashicorp/terraform-provider-aws/internal/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.PinpointSMSVoiceV2ServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"ServiceQuotaExceededException",
	)
}

func TestPoolValidatePhoneIdentity(t *testing.T) {
	t.Parallel()

	const identityARN = "arn:aws:sms-voice:us-east-1:111122223333:phone-number/abc" // lintignore:AWSAT003,AWSAT005

	testCases := []struct {
		TestName      string
		Info          awstypes.PhoneNumberInformation
		Intended      tfpinpointsmsvoicev2.IntendedIdentityConfig
		WantSummaries []string
	}{
		{
			TestName: "active_match",
			Info: awstypes.PhoneNumberInformation{
				Status:         awstypes.NumberStatusActive,
				MessageType:    awstypes.MessageTypeTransactional,
				IsoCountryCode: aws.String("US"),
			},
			Intended: tfpinpointsmsvoicev2.IntendedIdentityConfig{
				MessageType:    awstypes.MessageTypeTransactional,
				IsoCountryCode: "US",
			},
			WantSummaries: nil,
		},
		{
			TestName: "status_pending_short_circuits",
			Info: awstypes.PhoneNumberInformation{
				Status:         awstypes.NumberStatusPending,
				MessageType:    awstypes.MessageTypeTransactional,
				IsoCountryCode: aws.String("US"),
			},
			Intended: tfpinpointsmsvoicev2.IntendedIdentityConfig{
				MessageType:    awstypes.MessageTypeTransactional,
				IsoCountryCode: "US",
			},
			WantSummaries: []string{"is not ACTIVE"},
		},
		{
			TestName: "message_type_mismatch",
			Info: awstypes.PhoneNumberInformation{
				Status:         awstypes.NumberStatusActive,
				MessageType:    awstypes.MessageTypePromotional,
				IsoCountryCode: aws.String("US"),
			},
			Intended: tfpinpointsmsvoicev2.IntendedIdentityConfig{
				MessageType:    awstypes.MessageTypeTransactional,
				IsoCountryCode: "US",
			},
			WantSummaries: []string{"mismatched message_type"},
		},
		{
			TestName: "iso_country_code_mismatch",
			Info: awstypes.PhoneNumberInformation{
				Status:         awstypes.NumberStatusActive,
				MessageType:    awstypes.MessageTypeTransactional,
				IsoCountryCode: aws.String("GB"),
			},
			Intended: tfpinpointsmsvoicev2.IntendedIdentityConfig{
				MessageType:    awstypes.MessageTypeTransactional,
				IsoCountryCode: "US",
			},
			WantSummaries: []string{"mismatched iso_country_code"},
		},
		{
			TestName: "both_mismatch",
			Info: awstypes.PhoneNumberInformation{
				Status:         awstypes.NumberStatusActive,
				MessageType:    awstypes.MessageTypePromotional,
				IsoCountryCode: aws.String("GB"),
			},
			Intended: tfpinpointsmsvoicev2.IntendedIdentityConfig{
				MessageType:    awstypes.MessageTypeTransactional,
				IsoCountryCode: "US",
			},
			WantSummaries: []string{"mismatched message_type", "mismatched iso_country_code"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			diags := tfpinpointsmsvoicev2.ValidatePhoneIdentity(identityARN, testCase.Info, testCase.Intended)

			if got, want := len(diags), len(testCase.WantSummaries); got != want {
				t.Fatalf("diag count: got %d, want %d (%v)", got, want, diags)
			}
			for i, want := range testCase.WantSummaries {
				if !strings.Contains(diags[i].Summary(), want) {
					t.Errorf("diag %d summary %q does not contain %q", i, diags[i].Summary(), want)
				}
			}
		})
	}
}

func TestPoolValidateSenderIdentity(t *testing.T) {
	t.Parallel()

	const identityARN = "arn:aws:sms-voice:us-east-1:111122223333:sender-id/EXAMPLE/US" // lintignore:AWSAT003,AWSAT005

	testCases := []struct {
		TestName      string
		Info          awstypes.SenderIdInformation
		Intended      tfpinpointsmsvoicev2.IntendedIdentityConfig
		WantSummaries []string
	}{
		{
			TestName: "supported_match",
			Info: awstypes.SenderIdInformation{
				MessageTypes:   []awstypes.MessageType{awstypes.MessageTypeTransactional},
				IsoCountryCode: aws.String("US"),
			},
			Intended: tfpinpointsmsvoicev2.IntendedIdentityConfig{
				MessageType:    awstypes.MessageTypeTransactional,
				IsoCountryCode: "US",
			},
			WantSummaries: nil,
		},
		{
			TestName: "supported_multi_type_match",
			Info: awstypes.SenderIdInformation{
				MessageTypes:   []awstypes.MessageType{awstypes.MessageTypeTransactional, awstypes.MessageTypePromotional},
				IsoCountryCode: aws.String("US"),
			},
			Intended: tfpinpointsmsvoicev2.IntendedIdentityConfig{
				MessageType:    awstypes.MessageTypePromotional,
				IsoCountryCode: "US",
			},
			WantSummaries: nil,
		},
		{
			TestName: "unsupported_type",
			Info: awstypes.SenderIdInformation{
				MessageTypes:   []awstypes.MessageType{awstypes.MessageTypePromotional},
				IsoCountryCode: aws.String("US"),
			},
			Intended: tfpinpointsmsvoicev2.IntendedIdentityConfig{
				MessageType:    awstypes.MessageTypeTransactional,
				IsoCountryCode: "US",
			},
			WantSummaries: []string{"does not support message_type="},
		},
		{
			TestName: "iso_country_code_mismatch",
			Info: awstypes.SenderIdInformation{
				MessageTypes:   []awstypes.MessageType{awstypes.MessageTypeTransactional},
				IsoCountryCode: aws.String("GB"),
			},
			Intended: tfpinpointsmsvoicev2.IntendedIdentityConfig{
				MessageType:    awstypes.MessageTypeTransactional,
				IsoCountryCode: "US",
			},
			WantSummaries: []string{"mismatched iso_country_code"},
		},
		{
			TestName: "both_mismatch",
			Info: awstypes.SenderIdInformation{
				MessageTypes:   []awstypes.MessageType{awstypes.MessageTypePromotional},
				IsoCountryCode: aws.String("GB"),
			},
			Intended: tfpinpointsmsvoicev2.IntendedIdentityConfig{
				MessageType:    awstypes.MessageTypeTransactional,
				IsoCountryCode: "US",
			},
			WantSummaries: []string{"does not support message_type=", "mismatched iso_country_code"},
		},
		{
			TestName: "empty_message_types",
			Info: awstypes.SenderIdInformation{
				MessageTypes:   []awstypes.MessageType{},
				IsoCountryCode: aws.String("US"),
			},
			Intended: tfpinpointsmsvoicev2.IntendedIdentityConfig{
				MessageType:    awstypes.MessageTypeTransactional,
				IsoCountryCode: "US",
			},
			WantSummaries: []string{"does not support message_type="},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			diags := tfpinpointsmsvoicev2.ValidateSenderIdentity(identityARN, testCase.Info, testCase.Intended)

			if got, want := len(diags), len(testCase.WantSummaries); got != want {
				t.Fatalf("diag count: got %d, want %d (%v)", got, want, diags)
			}
			for i, want := range testCase.WantSummaries {
				if !strings.Contains(diags[i].Summary(), want) {
					t.Errorf("diag %d summary %q does not contain %q", i, diags[i].Summary(), want)
				}
			}
		})
	}
}

func TestAccPinpointSMSVoiceV2Pool_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.PoolInformation
	resourceName := "aws_pinpointsmsvoicev2_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, t, resourceName, &pool),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.StringRegexp(
						regexache.MustCompile(`arn:aws:sms-voice:[a-z0-9-]+:[0-9]{12}:pool/pool-.+$`))), // lintignore:AWSAT005
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("iso_country_code"), knownvalue.StringExact("US")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("message_type"), knownvalue.StringExact("TRANSACTIONAL")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("origination_identities"), knownvalue.SetSizeExact(1)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"iso_country_code"},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Pool_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.PoolInformation
	resourceName := "aws_pinpointsmsvoicev2_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, t, resourceName, &pool),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfpinpointsmsvoicev2.ResourcePool, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Pool_DeletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.PoolInformation
	resourceName := "aws_pinpointsmsvoicev2_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_DeletionProtection(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, t, resourceName, &pool),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(true)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"iso_country_code"},
			},
			{
				Config: testAccPoolConfig_DeletionProtection(true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: testAccPoolConfig_DeletionProtection(false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Pool_ISOCountryCode_mismatch(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccPoolConfig_ISOCountryCode_mismatch(),
				ExpectError: regexache.MustCompile(`mismatched iso_country_code`),
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Pool_MessageType(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.PoolInformation
	resourceName := "aws_pinpointsmsvoicev2_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, t, resourceName, &pool),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("message_type"), knownvalue.StringExact("TRANSACTIONAL")),
				},
			},
			{
				Config: testAccPoolConfig_basic(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: testAccPoolConfig_MessageTypePromotional(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				// SIMULATOR phone numbers only support TRANSACTIONAL message types.
				// ExpectError usage exercises a documented compatibility promise risk
				ExpectError: regexache.MustCompile(`ResourceNotFoundException`),
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Pool_MessageType_mismatch(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccPoolConfig_MessageType_mismatchOnCreate(),
				ExpectError: regexache.MustCompile(`mismatched message_type`),
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Pool_OptOutListName(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.PoolInformation
	var optOutList awstypes.OptOutListInformation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_pool.test"
	optOutListResourceName := "aws_pinpointsmsvoicev2_opt_out_list.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, t, resourceName, &pool),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("opt_out_list_name"), knownvalue.StringExact("Default")),
				},
			},
			{
				Config: testAccPoolConfig_OptOutListName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, t, resourceName, &pool),
					testAccCheckOptOutListExists(ctx, t, optOutListResourceName, &optOutList),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("opt_out_list_name"), optOutListResourceName, tfjsonpath.New(names.AttrName), compare.ValuesSame()),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"iso_country_code"},
			},
			{
				Config: testAccPoolConfig_OptOutListName(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: testAccPoolConfig_OptOutListNameDefault(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("opt_out_list_name"), knownvalue.StringExact("Default")),
				},
			},
			{
				Config: testAccPoolConfig_basic(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("opt_out_list_name"), knownvalue.StringExact("Default")),
				},
			},
		},
	})
}

// Once a phone is in a pool, the pool drives the shared mutable fields; planning a simultaneous
// write to the same logical field on both the phone and the pool is rejected by AWS.
func TestAccPinpointSMSVoiceV2Pool_OptOutListName_dualUpdateRejected(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.PoolInformation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_OptOutListName_dualUpdate(rName, `"Default"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, t, resourceName, &pool),
				),
			},
			{
				Config:      testAccPoolConfig_OptOutListName_dualUpdate(rName, "aws_pinpointsmsvoicev2_opt_out_list.test.name"),
				ExpectError: regexache.MustCompile(`(?i)pool|associated`),
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Pool_OptOutListName_mismatch(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.PoolInformation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdate := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_pool.test"
	optOutListResourceName := "aws_pinpointsmsvoicev2_opt_out_list.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_OptOutListName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, t, resourceName, &pool),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("opt_out_list_name"), optOutListResourceName, tfjsonpath.New(names.AttrName), compare.ValuesSame()),
				},
			},
			{
				Config:      testAccPoolConfig_OptOutListName_mismatchOnUpdate(rName, rNameUpdate),
				ExpectError: regexache.MustCompile(`(?i)match|configuration`),
			},
			{
				Config: testAccPoolConfig_OptOutListName(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Pool_OriginationIdentities(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.PoolInformation
	resourceName := "aws_pinpointsmsvoicev2_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_OriginationIdentities_two(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("origination_identities"), knownvalue.SetSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("origination_identities").AtSliceIndex(0), knownvalue.StringRegexp(regexache.MustCompile(`arn:aws:sms-voice:[a-z0-9-]+:[0-9]{12}:(phone-number|sender-id).+$`))), // lintignore:AWSAT005
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"iso_country_code"},
			},
			// Crosses the 5-ARN DescribePhoneNumbers per-request cap; exercises chunking.
			{
				Config: testAccPoolConfig_OriginationIdentities_six(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, t, resourceName, &pool),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("origination_identities"), knownvalue.SetSizeExact(6)),
				},
			},
			{
				Config: testAccPoolConfig_OriginationIdentities_six(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: testAccPoolConfig_OriginationIdentities_one(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, t, resourceName, &pool),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("origination_identities"), knownvalue.SetSizeExact(1)),
				},
			},
			{
				Config: testAccPoolConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, t, resourceName, &pool),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("origination_identities"), knownvalue.SetSizeExact(1)),
				},
			},
		},
	})
}

// Sorting origination identities lexicographically by ARN is predictable for sender-ids only.
// phone-number ARNs are fully computed by AWS. Therefore, the ordering of phone number resources
// are NOT anchored across test runs. For deterministic ordering we should use sender ids
// instead (once the resource is available).
func TestAccPinpointSMSVoiceV2Pool_OriginationIdentities_replaceSeed(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.PoolInformation
	resourceName := "aws_pinpointsmsvoicev2_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, t, resourceName, &pool),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("origination_identities"), knownvalue.SetSizeExact(1)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("origination_identities").AtSliceIndex(0), "aws_pinpointsmsvoicev2_phone_number.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				Config: testAccPoolConfig_OriginationIdentities_replaceSeed(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("origination_identities"), knownvalue.SetSizeExact(1)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("origination_identities").AtSliceIndex(0), "aws_pinpointsmsvoicev2_phone_number.update", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Pool_SharedRoutesEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.PoolInformation
	resourceName := "aws_pinpointsmsvoicev2_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_SharedRoutesEnabled(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, t, resourceName, &pool),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("shared_routes_enabled"), knownvalue.Bool(true)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"iso_country_code"},
			},
			{
				Config: testAccPoolConfig_SharedRoutesEnabled(true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: testAccPoolConfig_SharedRoutesEnabled(false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("shared_routes_enabled"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Pool_TwoWayChannelConnect(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.PoolInformation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_TwoWayChannelConnect(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, t, resourceName, &pool),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("two_way_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("two_way_channel_arn"), knownvalue.StringRegexp(regexache.MustCompile(`^connect\.[a-z0-9-]+\.amazonaws\.com`))),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("two_way_channel_role"), "aws_iam_role.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"iso_country_code"},
			},
			{
				Config: testAccPoolConfig_TwoWayChannelConnect(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Pool_TwoWayChannelSNS(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.PoolInformation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_TwoWayChannelSNS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPoolExists(ctx, t, resourceName, &pool),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("two_way_enabled"), knownvalue.Bool(true)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("two_way_channel_arn"), "aws_sns_topic.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("two_way_channel_role"), "aws_iam_role.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"iso_country_code"},
			},
			{
				Config: testAccPoolConfig_TwoWayChannelSNS(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func testAccCheckPoolDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpointsmsvoicev2_pool" {
				continue
			}

			_, err := tfpinpointsmsvoicev2.FindPoolByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("End User Messaging SMS Pool %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPoolExists(ctx context.Context, t *testing.T, n string, v *awstypes.PoolInformation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		output, err := tfpinpointsmsvoicev2.FindPoolByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output
		return nil
	}
}

func testAccPreCheckPool(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

	_, err := conn.DescribePools(ctx, &pinpointsmsvoicev2.DescribePoolsInput{})

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPoolConfig_basic() string {
	return `
resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}
`
}

func testAccPoolConfig_DeletionProtection(enabled bool) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_pool" "test" {
  deletion_protection_enabled = %[1]t
  iso_country_code            = "US"
  message_type                = "TRANSACTIONAL"
  origination_identities      = [aws_pinpointsmsvoicev2_phone_number.test.arn]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}
`, enabled)
}

func testAccPoolConfig_ISOCountryCode_mismatch() string {
	return `
resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code       = "CA"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}
`
}

func testAccPoolConfig_MessageTypePromotional() string {
	return `
resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code       = "US"
  message_type           = "PROMOTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "PROMOTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}
`
}

func testAccPoolConfig_MessageType_mismatchOnCreate() string {
	return `
resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code       = "US"
  message_type           = "PROMOTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}
`
}

func testAccPoolConfig_OptOutListName(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  opt_out_list_name      = aws_pinpointsmsvoicev2_opt_out_list.test.name
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_opt_out_list" "test" {
  name = %[1]q
}
`, rName)
}

func testAccPoolConfig_OptOutListNameDefault(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  opt_out_list_name      = "Default"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_opt_out_list" "test" {
  name = %[1]q
}
`, rName)
}

func testAccPoolConfig_OptOutListName_dualUpdate(rName, value string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]
  opt_out_list_name      = %[2]s
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
  opt_out_list_name   = %[2]s
}

resource "aws_pinpointsmsvoicev2_opt_out_list" "test" {
  name = %[1]q
}
`, rName, value)
}

func testAccPoolConfig_OptOutListName_mismatchOnUpdate(rName, rNameUpdate string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code  = "US"
  message_type      = "TRANSACTIONAL"
  opt_out_list_name = aws_pinpointsmsvoicev2_opt_out_list.test.name
  origination_identities = [
    aws_pinpointsmsvoicev2_phone_number.test.arn,
    aws_pinpointsmsvoicev2_phone_number.mismatched.arn,
  ]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_phone_number" "mismatched" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
  opt_out_list_name   = aws_pinpointsmsvoicev2_opt_out_list.mismatched.name
}

resource "aws_pinpointsmsvoicev2_opt_out_list" "test" {
  name = %[1]q
}

resource "aws_pinpointsmsvoicev2_opt_out_list" "mismatched" {
  name = %[2]q
}
`, rName, rNameUpdate)
}

func testAccPoolConfig_OriginationIdentities_one() string {
	return `
resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code = "US"
  message_type     = "TRANSACTIONAL"
  origination_identities = [
    aws_pinpointsmsvoicev2_phone_number.test.arn,
  ]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test_with_count" {
  count = 4

  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
  force_disassociate  = true
}
`
}

func testAccPoolConfig_OriginationIdentities_two() string {
	return `
resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code = "US"
  message_type     = "TRANSACTIONAL"
  origination_identities = [
    aws_pinpointsmsvoicev2_phone_number.test.arn,
    aws_pinpointsmsvoicev2_phone_number.test_with_count[0].arn
  ]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test_with_count" {
  count = 1

  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
  force_disassociate  = true
}
`
}

func testAccPoolConfig_OriginationIdentities_six() string {
	return `
resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code = "US"
  message_type     = "TRANSACTIONAL"
  origination_identities = [
    aws_pinpointsmsvoicev2_phone_number.test.arn,
    aws_pinpointsmsvoicev2_phone_number.test_with_count[0].arn,
    aws_pinpointsmsvoicev2_phone_number.test_with_count[1].arn,
    aws_pinpointsmsvoicev2_phone_number.test_with_count[2].arn,
    aws_pinpointsmsvoicev2_phone_number.test_with_count[3].arn,
    aws_pinpointsmsvoicev2_phone_number.test_with_count[4].arn
  ]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test_with_count" {
  count = 5

  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
  force_disassociate  = true
}
`
}

func testAccPoolConfig_OriginationIdentities_replaceSeed() string {
	return `
resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.update.arn]
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_phone_number" "update" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}
`
}

func testAccPoolConfig_SharedRoutesEnabled(enabled bool) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]
  shared_routes_enabled  = %[1]t
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}
`, enabled)
}

func testAccPoolConfig_TwoWayChannelConnect(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]
  two_way_enabled        = true
  two_way_channel_arn    = "connect.${data.aws_region.current.region}.amazonaws.com"
  two_way_channel_role   = aws_iam_role.test.arn
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS", "VOICE"]
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "sms-voice.amazonaws.com"
      }
      Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [{
      Effect   = "Allow"
      Action   = ["connect:SendChatIntegrationEvent"]
      Resource = ["*"]
    }]
  })
}
`, rName)
}

func testAccPoolConfig_TwoWayChannelSNS(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]
  two_way_enabled        = true
  two_way_channel_arn    = aws_sns_topic.test.arn
  two_way_channel_role   = aws_iam_role.test.arn
}

resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "sms-voice.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["sns:Publish"]
      Resource = aws_sns_topic.test.arn
    }]
  })
}
`, rName)
}
