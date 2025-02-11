// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpoint/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfpinpoint "github.com/hashicorp/terraform-provider-aws/internal/service/pinpoint"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointEmailChannel_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var channel awstypes.EmailChannelResponse
	resourceName := "aws_pinpoint_email_channel.test"

	domain := acctest.RandomDomainName()
	address1 := acctest.RandomEmailAddress(domain)
	address2 := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailChannelConfig_fromAddress(domain, address1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "from_address", address1),
					resource.TestCheckResourceAttrSet(resourceName, "messages_per_second"),
					resource.TestCheckResourceAttrPair(resourceName, "identity", "aws_ses_domain_identity.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "orchestration_sending_role_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEmailChannelConfig_fromAddress(domain, address2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttr(resourceName, names.AttrEnabled, acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "from_address", address2),
					resource.TestCheckResourceAttrSet(resourceName, "messages_per_second"),
				),
			},
		},
	})
}

func TestAccPinpointEmailChannel_set(t *testing.T) {
	ctx := acctest.Context(t)
	var channel awstypes.EmailChannelResponse
	resourceName := "aws_pinpoint_email_channel.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	domain := acctest.RandomDomainName()
	address := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailChannelConfig_configurationSet(domain, address, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_set", "aws_ses_configuration_set.test", names.AttrName),
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

func TestAccPinpointEmailChannel_noRole(t *testing.T) {
	ctx := acctest.Context(t)
	var channel awstypes.EmailChannelResponse
	resourceName := "aws_pinpoint_email_channel.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	domain := acctest.RandomDomainName()
	address := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailChannelConfig_noRole(domain, address, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_set", "aws_ses_configuration_set.test", names.AttrARN),
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

func TestAccPinpointEmailChannel_orchestrationSendingRoleARN(t *testing.T) {
	ctx := acctest.Context(t)
	var channel awstypes.EmailChannelResponse
	resourceName := "aws_pinpoint_email_channel.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	domain := acctest.RandomDomainName()
	address := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailChannelConfig_orchestrationSendingRoleARN(domain, address, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailChannelExists(ctx, resourceName, &channel),
					resource.TestCheckResourceAttrPair(resourceName, "orchestration_sending_role_arn", "aws_iam_role.test", names.AttrARN),
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

func TestAccPinpointEmailChannel_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var channel awstypes.EmailChannelResponse
	resourceName := "aws_pinpoint_email_channel.test"

	domain := acctest.RandomDomainName()
	address := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailChannelDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailChannelConfig_fromAddress(domain, address),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailChannelExists(ctx, resourceName, &channel),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfpinpoint.ResourceEmailChannel(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEmailChannelExists(ctx context.Context, n string, channel *awstypes.EmailChannelResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Pinpoint Email Channel with that application ID exists")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointClient(ctx)

		output, err := tfpinpoint.FindEmailChannelByApplicationId(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*channel = *output

		return nil
	}
}

func testAccCheckEmailChannelDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpoint_email_channel" {
				continue
			}

			_, err := tfpinpoint.FindEmailChannelByApplicationId(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Pinpoint Email Channel %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccEmailChannelConfig_fromAddress(domain, fromAddress string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {}

resource "aws_pinpoint_email_channel" "test" {
  application_id = aws_pinpoint_app.test.application_id
  enabled        = "false"
  from_address   = %[2]q
  identity       = aws_ses_domain_identity.test.arn
}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "pinpoint.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "test"
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Action": [
      "mobileanalytics:PutEvents",
      "mobileanalytics:PutItems"
    ],
    "Effect": "Allow",
    "Resource": [
      "*"
    ]
  }
}
EOF
}
`, domain, fromAddress)
}

func testAccEmailChannelConfig_configurationSet(domain, fromAddress, rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {}

resource "aws_ses_configuration_set" "test" {
  name = %[3]q
}

resource "aws_pinpoint_email_channel" "test" {
  application_id    = aws_pinpoint_app.test.application_id
  enabled           = "false"
  from_address      = %[2]q
  identity          = aws_ses_domain_identity.test.arn
  configuration_set = aws_ses_configuration_set.test.name
}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "pinpoint.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "test"
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Action": [
      "mobileanalytics:PutEvents",
      "mobileanalytics:PutItems"
    ],
    "Effect": "Allow",
    "Resource": [
      "*"
    ]
  }
}
EOF
}
`, domain, fromAddress, rName)
}

func testAccEmailChannelConfig_noRole(domain, fromAddress, rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {}

resource "aws_ses_configuration_set" "test" {
  name = %[3]q
}

resource "aws_pinpoint_email_channel" "test" {
  application_id    = aws_pinpoint_app.test.application_id
  enabled           = "false"
  from_address      = %[2]q
  identity          = aws_ses_domain_identity.test.arn
  configuration_set = aws_ses_configuration_set.test.arn
}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}
`, domain, fromAddress, rName)
}

func testAccEmailChannelConfig_orchestrationSendingRoleARN(domain, fromAddress, rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_app" "test" {}

resource "aws_ses_configuration_set" "test" {
  name = %[3]q
}

resource "aws_pinpoint_email_channel" "test" {
  application_id                 = aws_pinpoint_app.test.application_id
  enabled                        = "false"
  from_address                   = %[2]q
  identity                       = aws_ses_domain_identity.test.arn
  configuration_set              = aws_ses_configuration_set.test.arn
  orchestration_sending_role_arn = aws_iam_role.test.arn
}

resource "aws_ses_domain_identity" "test" {
  domain = %[1]q
}

resource "aws_iam_role" "test" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "pinpoint.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = "test"
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Action": [
      "ses:SendEmail",
      "ses:SendRawEmail"
    ],
    "Effect": "Allow",
    "Resource": [
      "${aws_ses_configuration_set.test.arn}",
      "${aws_ses_domain_identity.test.arn}"
    ]
  }
}
EOF
}
`, domain, fromAddress, rName)
}
