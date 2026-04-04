// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfpinpointsmsvoicev2 "github.com/hashicorp/terraform-provider-aws/internal/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointSMSVoiceV2SenderID_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var senderID awstypes.SenderIdInformation
	senderIDName := "TfBasic"
	isoCountryCode := "GB"
	resourceName := "aws_pinpointsmsvoicev2_sender_id.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSenderID(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSenderIDDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSenderIDConfig_basic(senderIDName, isoCountryCode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSenderIDExists(ctx, t, resourceName, &senderID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("iso_country_code"), knownvalue.StringExact(isoCountryCode)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"sender_id"},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2SenderID_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var senderID awstypes.SenderIdInformation
	senderIDName := "TfDisappr"
	isoCountryCode := "GB"
	resourceName := "aws_pinpointsmsvoicev2_sender_id.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSenderID(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSenderIDDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSenderIDConfig_basic(senderIDName, isoCountryCode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSenderIDExists(ctx, t, resourceName, &senderID),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfpinpointsmsvoicev2.ResourceSenderID, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2SenderID_deletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	var senderID awstypes.SenderIdInformation
	senderIDName := "TfDelProt"
	isoCountryCode := "GB"
	resourceName := "aws_pinpointsmsvoicev2_sender_id.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSenderID(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSenderIDDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSenderIDConfig_deletionProtection(senderIDName, isoCountryCode, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSenderIDExists(ctx, t, resourceName, &senderID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(true)),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"sender_id"},
			},
			{
				Config: testAccSenderIDConfig_deletionProtection(senderIDName, isoCountryCode, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSenderIDExists(ctx, t, resourceName, &senderID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("deletion_protection_enabled"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2SenderID_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var senderID awstypes.SenderIdInformation
	senderIDName := "TfTags"
	isoCountryCode := "GB"
	resourceName := "aws_pinpointsmsvoicev2_sender_id.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckSenderID(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSenderIDDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSenderIDConfig_tags1(senderIDName, isoCountryCode, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSenderIDExists(ctx, t, resourceName, &senderID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"sender_id"},
			},
			{
				Config: testAccSenderIDConfig_tags2(senderIDName, isoCountryCode, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSenderIDExists(ctx, t, resourceName, &senderID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccSenderIDConfig_tags1(senderIDName, isoCountryCode, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSenderIDExists(ctx, t, resourceName, &senderID),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckSenderIDDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpointsmsvoicev2_sender_id" {
				continue
			}

			senderID := rs.Primary.Attributes["sender_id"]
			isoCountryCode := rs.Primary.Attributes["iso_country_code"]

			_, err := tfpinpointsmsvoicev2.FindSenderIDByTwoPartKey(ctx, conn, senderID, isoCountryCode)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("End User Messaging SMS Sender ID %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSenderIDExists(ctx context.Context, t *testing.T, n string, v *awstypes.SenderIdInformation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		senderID := rs.Primary.Attributes["sender_id"]
		isoCountryCode := rs.Primary.Attributes["iso_country_code"]

		conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

		output, err := tfpinpointsmsvoicev2.FindSenderIDByTwoPartKey(ctx, conn, senderID, isoCountryCode)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheckSenderID(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).PinpointSMSVoiceV2Client(ctx)

	input := &pinpointsmsvoicev2.DescribeSenderIdsInput{}

	_, err := conn.DescribeSenderIds(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSenderIDConfig_basic(senderID, isoCountryCode string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_sender_id" "test" {
  sender_id        = %[1]q
  iso_country_code = %[2]q
}
`, senderID, isoCountryCode)
}

func testAccSenderIDConfig_deletionProtection(senderID, isoCountryCode string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_sender_id" "test" {
  sender_id                   = %[1]q
  iso_country_code            = %[2]q
  deletion_protection_enabled = %[3]t
}
`, senderID, isoCountryCode, deletionProtection)
}

func testAccSenderIDConfig_tags1(senderID, isoCountryCode, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_sender_id" "test" {
  sender_id        = %[1]q
  iso_country_code = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, senderID, isoCountryCode, tagKey1, tagValue1)
}

func testAccSenderIDConfig_tags2(senderID, isoCountryCode, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_sender_id" "test" {
  sender_id        = %[1]q
  iso_country_code = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, senderID, isoCountryCode, tagKey1, tagValue1, tagKey2, tagValue2)
}
