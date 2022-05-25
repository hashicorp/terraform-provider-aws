package pinpointsmsvoicev2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/pinpointsmsvoicev2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"

	tfpinpointsmsvoicev2 "github.com/hashicorp/terraform-provider-aws/internal/service/pinpointsmsvoicev2"
)

func TestAccPinpointSMSVoiceV2PhoneNumber_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var phoneNumber pinpointsmsvoicev2.DescribePhoneNumbersOutput
	resourceName := "aws_pinpointsmsvoicev2_phone_number.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(pinpointsmsvoicev2.EndpointsID, t)
			testAccPreCheckPhoneNumber(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, pinpointsmsvoicev2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPhoneNumberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(resourceName, &phoneNumber),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var phoneNumber pinpointsmsvoicev2.DescribePhoneNumbersOutput
	phoneNumberName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	snsTopicName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	optOutListName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_pinpointsmsvoicev2_phone_number.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(pinpointsmsvoicev2.EndpointsID, t)
			testAccPreCheckPhoneNumber(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, pinpointsmsvoicev2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPhoneNumberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberConfigFull(phoneNumberName, snsTopicName, optOutListName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(resourceName, &phoneNumber),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var phoneNumber pinpointsmsvoicev2.DescribePhoneNumbersOutput
	resourceName := "aws_pinpointsmsvoicev2_phone_number.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(pinpointsmsvoicev2.EndpointsID, t)
			testAccPreCheckPhoneNumber(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, pinpointsmsvoicev2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPhoneNumberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNumberConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPhoneNumberExists(resourceName, &phoneNumber),
					acctest.CheckResourceDisappears(acctest.Provider, tfpinpointsmsvoicev2.ResourcePhoneNumber(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPhoneNumberDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointSMSVoiceV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_pinpointsmsvoicev2_phone_number" {
			continue
		}

		input := &pinpointsmsvoicev2.DescribePhoneNumbersInput{
			PhoneNumberIds: aws.StringSlice([]string{rs.Primary.ID}),
		}

		_, err := conn.DescribePhoneNumbers(input)
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

func testAccCheckPhoneNumberExists(name string, phoneNumber *pinpointsmsvoicev2.DescribePhoneNumbersOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no PinpointSMSVoiceV2 PhoneNumber is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointSMSVoiceV2Conn
		resp, err := conn.DescribePhoneNumbers(&pinpointsmsvoicev2.DescribePhoneNumbersInput{
			PhoneNumberIds: aws.StringSlice([]string{rs.Primary.ID}),
		})

		if err != nil {
			return fmt.Errorf("error describing PinpointSMSVoiceV2 PhoneNumber: %s", err.Error())
		}

		*phoneNumber = *resp

		return nil
	}
}

func testAccPreCheckPhoneNumber(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointSMSVoiceV2Conn

	input := &pinpointsmsvoicev2.DescribePhoneNumbersInput{}

	_, err := conn.DescribePhoneNumbers(input)

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
