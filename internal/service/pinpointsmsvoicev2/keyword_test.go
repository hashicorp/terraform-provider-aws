// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfpinpointsmsvoicev2 "github.com/hashicorp/terraform-provider-aws/internal/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointSMSVoiceV2Keyword_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var keyword awstypes.KeywordInformation
	rName := randomKeywordName(t)
	resourceName := "aws_pinpointsmsvoicev2_keyword.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckKeyword(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeywordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeywordConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeywordExists(ctx, t, resourceName, &keyword),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword_action"), knownvalue.StringExact("AUTOMATIC_RESPONSE")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword_message"), knownvalue.StringExact("test keyword message")),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("origination_identity"), "aws_pinpointsmsvoicev2_phone_number.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("origination_identity_arn"), "aws_pinpointsmsvoicev2_phone_number.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "origination_identity", "keyword"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "origination_identity",
			},
			{
				Config: testAccKeywordConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Keyword_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var keyword awstypes.KeywordInformation
	rName := randomKeywordName(t)
	resourceName := "aws_pinpointsmsvoicev2_keyword.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckKeyword(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeywordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeywordConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeywordExists(ctx, t, resourceName, &keyword),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfpinpointsmsvoicev2.ResourceKeyword, resourceName),
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

func TestAccPinpointSMSVoiceV2Keyword_KeywordHELP(t *testing.T) {
	ctx := acctest.Context(t)
	var keyword awstypes.KeywordInformation
	resourceName := "aws_pinpointsmsvoicev2_keyword.test"
	originationName := "aws_pinpointsmsvoicev2_phone_number.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckKeyword(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeywordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeywordConfig_Keyword_mandatory("HELP", "mandatory help message"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeywordExists(ctx, t, resourceName, &keyword),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword"), knownvalue.StringExact("HELP")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword_action"), knownvalue.StringExact("AUTOMATIC_RESPONSE")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword_message"), knownvalue.StringExact("mandatory help message")),
				},
			},
			{
				Config: testAccKeywordConfig_Keyword_mandatory("HELP", "updated help message"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword_message"), knownvalue.StringExact("updated help message")),
				},
			},
			{
				Config: testAccKeywordConfig_Keyword_mandatoryBase(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroy),
						plancheck.ExpectResourceAction(originationName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Keyword_KeywordMessage(t *testing.T) {
	ctx := acctest.Context(t)
	var keyword awstypes.KeywordInformation
	rName := randomKeywordName(t)
	resourceName := "aws_pinpointsmsvoicev2_keyword.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckKeyword(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeywordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeywordConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeywordExists(ctx, t, resourceName, &keyword),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword_message"), knownvalue.StringExact("test keyword message")),
				},
			},
			{
				Config: testAccKeywordConfig_KeywordMessage_updated(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword_message"), knownvalue.StringExact("updated keyword message")),
				},
			},
			{
				Config: testAccKeywordConfig_KeywordMessage_updated(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Keyword_KeywordAction(t *testing.T) {
	ctx := acctest.Context(t)
	var keyword awstypes.KeywordInformation
	rName := randomKeywordName(t)
	resourceName := "aws_pinpointsmsvoicev2_keyword.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckKeyword(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeywordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeywordConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeywordExists(ctx, t, resourceName, &keyword),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword_action"), knownvalue.StringExact("AUTOMATIC_RESPONSE")),
				},
			},
			{
				Config: testAccKeywordConfig_KeywordActionOptOut(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword_action"), knownvalue.StringExact("OPT_OUT")),
				},
			},
			{
				Config: testAccKeywordConfig_KeywordActionOptOut(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Keyword_Keyword_mandatoryActionError(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckKeyword(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeywordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccKeywordConfig_Keyword_mandatoryWithAction("HELP", "AUTOMATIC_RESPONSE"),
				ExpectError: regexache.MustCompile(`keyword_action is managed by AWS for the mandatory keyword "HELP"`),
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Keyword_KeywordSTOP(t *testing.T) {
	ctx := acctest.Context(t)
	var keyword awstypes.KeywordInformation
	resourceName := "aws_pinpointsmsvoicev2_keyword.test"
	originationName := "aws_pinpointsmsvoicev2_phone_number.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckKeyword(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeywordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeywordConfig_Keyword_mandatory("STOP", "mandatory stop message"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeywordExists(ctx, t, resourceName, &keyword),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword"), knownvalue.StringExact("STOP")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword_action"), knownvalue.StringExact("OPT_OUT")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword_message"), knownvalue.StringExact("mandatory stop message")),
				},
			},
			{
				Config: testAccKeywordConfig_Keyword_mandatory("STOP", "updated stop message"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword_message"), knownvalue.StringExact("updated stop message")),
				},
			},
			{
				Config: testAccKeywordConfig_Keyword_mandatoryBase(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroy),
						plancheck.ExpectResourceAction(originationName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Keyword_OriginationIdentityPhoneNumberARN(t *testing.T) {
	ctx := acctest.Context(t)
	var keyword awstypes.KeywordInformation
	rName := randomKeywordName(t)
	resourceName := "aws_pinpointsmsvoicev2_keyword.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckKeyword(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeywordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeywordConfig_OriginationIdentityPhoneNumberARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeywordExists(ctx, t, resourceName, &keyword),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword"), knownvalue.StringExact(rName)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("origination_identity"), "aws_pinpointsmsvoicev2_phone_number.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("origination_identity_arn"), "aws_pinpointsmsvoicev2_phone_number.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				Config: testAccKeywordConfig_OriginationIdentityPhoneNumberARN(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Keyword_OriginationIdentityPoolID(t *testing.T) {
	ctx := acctest.Context(t)
	var keyword awstypes.KeywordInformation
	rName := randomKeywordName(t)
	resourceName := "aws_pinpointsmsvoicev2_keyword.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckKeyword(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeywordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeywordConfig_OriginationIdentityPoolID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeywordExists(ctx, t, resourceName, &keyword),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword"), knownvalue.StringExact(rName)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("origination_identity"), "aws_pinpointsmsvoicev2_pool.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("origination_identity_arn"), "aws_pinpointsmsvoicev2_pool.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				Config: testAccKeywordConfig_OriginationIdentityPoolID(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2Keyword_OriginationIdentityPoolARN(t *testing.T) {
	ctx := acctest.Context(t)
	var keyword awstypes.KeywordInformation
	rName := randomKeywordName(t)
	resourceName := "aws_pinpointsmsvoicev2_keyword.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckKeyword(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeywordDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeywordConfig_OriginationIdentityPoolARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeywordExists(ctx, t, resourceName, &keyword),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("keyword"), knownvalue.StringExact(rName)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("origination_identity"), "aws_pinpointsmsvoicev2_pool.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("origination_identity_arn"), "aws_pinpointsmsvoicev2_pool.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				Config: testAccKeywordConfig_OriginationIdentityPoolARN(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

// randomKeywordName returns a name that fits the keyword's 30-character limit
func randomKeywordName(t *testing.T) string {
	return fmt.Sprintf("%s-%s", acctest.ResourcePrefix, acctest.RandStringFromCharSet(t, 8, acctest.CharSetAlpha))
}

func testAccCheckKeywordDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpointsmsvoicev2_keyword" {
				continue
			}

			_, _, err := tfpinpointsmsvoicev2.FindKeywordByTwoPartKey(ctx, conn, rs.Primary.Attributes["origination_identity"], rs.Primary.Attributes["keyword"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("End User Messaging SMS Keyword %s still exists", rs.Primary.Attributes["keyword"])
		}

		return nil
	}
}

func testAccCheckKeywordExists(ctx context.Context, t *testing.T, n string, v *awstypes.KeywordInformation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		output, _, err := tfpinpointsmsvoicev2.FindKeywordByTwoPartKey(ctx, conn, rs.Primary.Attributes["origination_identity"], rs.Primary.Attributes["keyword"])

		*v = *output
		return err
	}
}

func testAccPreCheckKeyword(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

	// no arg-free DescribeKeywords (needs origination identity)
	// probe a phone number instead to verify service availability
	input := pinpointsmsvoicev2.DescribePhoneNumbersInput{}
	_, err := conn.DescribePhoneNumbers(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccKeywordConfig_basic(keyword string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_keyword" "test" {
  origination_identity = aws_pinpointsmsvoicev2_phone_number.test.id
  keyword              = %[1]q
  keyword_message      = "test keyword message"
}
`, keyword)
}

func testAccKeywordConfig_KeywordActionOptOut(keyword string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_keyword" "test" {
  origination_identity = aws_pinpointsmsvoicev2_phone_number.test.id
  keyword              = %[1]q
  keyword_message      = "test keyword message"
  keyword_action       = "OPT_OUT"
}
`, keyword)
}

func testAccKeywordConfig_Keyword_mandatoryBase() string {
	return `
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}
`
}

func testAccKeywordConfig_Keyword_mandatory(keyword, message string) string {
	return acctest.ConfigCompose(testAccKeywordConfig_Keyword_mandatoryBase(), fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_keyword" "test" {
  origination_identity = aws_pinpointsmsvoicev2_phone_number.test.id
  keyword              = %[1]q
  keyword_message      = %[2]q
}
`, keyword, message))
}

func testAccKeywordConfig_Keyword_mandatoryWithAction(keyword, action string) string {
	return acctest.ConfigCompose(testAccKeywordConfig_Keyword_mandatoryBase(), fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_keyword" "test" {
  origination_identity = aws_pinpointsmsvoicev2_phone_number.test.id
  keyword              = %[1]q
  keyword_message      = "mandatory keyword message"
  keyword_action       = %[2]q
}
`, keyword, action))
}

func testAccKeywordConfig_KeywordMessage_updated(keyword string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_keyword" "test" {
  origination_identity = aws_pinpointsmsvoicev2_phone_number.test.id
  keyword              = %[1]q
  keyword_message      = "updated keyword message"
}
`, keyword)
}

func testAccKeywordConfig_OriginationIdentityPhoneNumberARN(keyword string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_keyword" "test" {
  origination_identity = aws_pinpointsmsvoicev2_phone_number.test.arn
  keyword              = %[1]q
  keyword_message      = "test keyword message"
}
`, keyword)
}

func testAccKeywordConfig_OriginationIdentityPoolID(keyword string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  force_disassociate  = true
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]
}

resource "aws_pinpointsmsvoicev2_keyword" "test" {
  origination_identity = aws_pinpointsmsvoicev2_pool.test.id
  keyword              = %[1]q
  keyword_message      = "test keyword message"
}
`, keyword)
}

func testAccKeywordConfig_OriginationIdentityPoolARN(keyword string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  force_disassociate  = true
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

resource "aws_pinpointsmsvoicev2_pool" "test" {
  iso_country_code       = "US"
  message_type           = "TRANSACTIONAL"
  origination_identities = [aws_pinpointsmsvoicev2_phone_number.test.arn]
}

resource "aws_pinpointsmsvoicev2_keyword" "test" {
  origination_identity = aws_pinpointsmsvoicev2_pool.test.arn
  keyword              = %[1]q
  keyword_message      = "test keyword message"
}
`, keyword)
}
