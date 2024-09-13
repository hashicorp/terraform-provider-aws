// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpointsmsvoicev2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfpinpointsmsvoicev2 "github.com/hashicorp/terraform-provider-aws/internal/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointSMSVoiceV2PhoneNumber_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var phoneNumber pinpointsmsvoicev2.DescribePhoneNumbersOutput
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
				Config: testAccPhoneNumberConfigBasic,
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
	var phoneNumber pinpointsmsvoicev2.DescribePhoneNumbersOutput
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
				Config: testAccPhoneNumberConfigFull(phoneNumberName, snsTopicName, optOutListName),
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
	var phoneNumber pinpointsmsvoicev2.DescribePhoneNumbersOutput
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
				Config: testAccPhoneNumberConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(ctx, resourceName, &phoneNumber),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfpinpointsmsvoicev2.ResourcePhoneNumber(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPhoneNumberDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointSMSVoiceV2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpointsmsvoicev2_phone_number" {
				continue
			}

			input := &pinpointsmsvoicev2.DescribePhoneNumbersInput{
				PhoneNumberIds: aws.StringSlice([]string{rs.Primary.ID}),
			}

			_, err := conn.DescribePhoneNumbersWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, pinpointsmsvoicev2.ErrCodeResourceNotFoundException) {
					return nil
				}
				return err
			}

			return fmt.Errorf("expected PinpointSMSVoiceV2 PhoneNumber to be destroyed, %s found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPhoneNumberExists(ctx context.Context, n string, v *pinpointsmsvoicev2.DescribePhoneNumbersOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointSMSVoiceV2Conn(ctx)

		resp, err := conn.DescribePhoneNumbersWithContext(ctx, &pinpointsmsvoicev2.DescribePhoneNumbersInput{
			PhoneNumberIds: aws.StringSlice([]string{rs.Primary.ID}),
		})

		if err != nil {
			return fmt.Errorf("error describing PinpointSMSVoiceV2 PhoneNumber: %s", err.Error())
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckPhoneNumber(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointSMSVoiceV2Conn(ctx)

	input := &pinpointsmsvoicev2.DescribePhoneNumbersInput{}

	_, err := conn.DescribePhoneNumbersWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

const testAccPhoneNumberConfigBasic = `
resource "aws_pinpointsmsvoicev2_phone_number" "test" {
  iso_country_code = "US"
  message_type     = "TRANSACTIONAL"
  number_type      = "TOLL_FREE"

  number_capabilities = [
    "SMS"
  ]
}
`

func testAccPhoneNumberConfigFull(phoneNumberName, snsTopicName, optOutListName string) string {
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
