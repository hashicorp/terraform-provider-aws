// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfpinpointsmsvoicev2 "github.com/hashicorp/terraform-provider-aws/internal/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointSMSVoiceV2PhoneNumber_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var phoneNumber awstypes.PhoneNumberInformation
	resourceName := "aws_pinpointsmsvoicev2_phone_number.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPhoneNumber(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(ctx, resourceName, &phoneNumber),
					resource.TestCheckResourceAttr(resourceName, "iso_country_code", "US"),
					resource.TestCheckResourceAttr(resourceName, "message_type", "TRANSACTIONAL"),
					resource.TestCheckResourceAttr(resourceName, "number_type", "TOLL_FREE"),
					resource.TestCheckResourceAttr(resourceName, "number_capabilities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "number_capabilities.0", "SMS"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2PhoneNumber_full(t *testing.T) {
	ctx := acctest.Context(t)
	var phoneNumber awstypes.PhoneNumberInformation
	phoneNumberName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	snsTopicName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	optOutListName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_phone_number.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPhoneNumber(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberConfig_full(phoneNumberName, snsTopicName, optOutListName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(ctx, resourceName, &phoneNumber),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "iso_country_code", "US"),
					resource.TestCheckResourceAttr(resourceName, "message_type", "TRANSACTIONAL"),
					resource.TestCheckResourceAttr(resourceName, "number_type", "TOLL_FREE"),
					resource.TestCheckResourceAttr(resourceName, "opt_out_list_name", optOutListName),
					resource.TestCheckResourceAttr(resourceName, "self_managed_opt_outs_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "two_way_channel_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "number_capabilities.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "number_capabilities.0", "SMS"),
					resource.TestCheckResourceAttr(resourceName, "number_capabilities.1", "VOICE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2PhoneNumber_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var phoneNumber awstypes.PhoneNumberInformation
	resourceName := "aws_pinpointsmsvoicev2_phone_number.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPhoneNumber(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(ctx, resourceName, &phoneNumber),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfpinpointsmsvoicev2.ResourcePhoneNumber(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccPinpointSMSVoiceV2PhoneNumber_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var phoneNumber awstypes.PhoneNumberInformation
	resourceName := "aws_pinpointsmsvoicev2_phone_number.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPhoneNumber(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointSMSVoiceV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPhoneNumberDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(ctx, resourceName, &phoneNumber),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPhoneNumberConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(ctx, resourceName, &phoneNumber),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccPhoneNumberConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(ctx, resourceName, &phoneNumber),
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

func testAccCheckPhoneNumberDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointSMSVoiceV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpointsmsvoicev2_phone_number" {
				continue
			}

			_, err := tfpinpointsmsvoicev2.FindPhoneNumberByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("End User Messaging Phone Number %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPhoneNumberExists(ctx context.Context, n string, v *awstypes.PhoneNumberInformation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointSMSVoiceV2Client(ctx)

		output, err := tfpinpointsmsvoicev2.FindPhoneNumberByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheckPhoneNumber(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointSMSVoiceV2Client(ctx)

	input := &pinpointsmsvoicev2.DescribePhoneNumbersInput{}

	_, err := conn.DescribePhoneNumbers(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

const testAccPhoneNumberConfig_basic = `
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code = "US"
  message_type     = "TRANSACTIONAL"
  number_type      = "TOLL_FREE"

  number_capabilities = [
    "SMS"
  ]
}
`

func testAccPhoneNumberConfig_full(phoneNumberName, snsTopicName, optOutListName string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  deletion_protection_enabled   = false
  iso_country_code              = "US"
  message_type                  = "TRANSACTIONAL"
  number_type                   = "TOLL_FREE"
  opt_out_list_name             = aws_pinpointsmsvoicev2_opt_out_list.test.name
  self_managed_opt_outs_enabled = false
  two_way_channel_arn           = aws_sns_topic.test.arn
  two_way_channel_enabled       = true

  number_capabilities = [
    "SMS",
    "VOICE",
  ]

  tags = {
    Name = %[1]q
  }
}

resource "aws_sns_topic" "test" {
  name = %[2]q
}

resource "aws_pinpointsmsvoicev2_opt_out_list" "test" {
  name = %[3]q
}
`, phoneNumberName, snsTopicName, optOutListName)
}

func testAccPhoneNumberConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code = "US"
  message_type     = "TRANSACTIONAL"
  number_type      = "TOLL_FREE"

  number_capabilities = [
    "SMS"
  ]

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccPhoneNumberConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code = "US"
  message_type     = "TRANSACTIONAL"
  number_type      = "TOLL_FREE"

  number_capabilities = [
    "SMS"
  ]

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
