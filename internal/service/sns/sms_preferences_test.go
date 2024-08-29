// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsns "github.com/hashicorp/terraform-provider-aws/internal/service/sns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// The preferences are account-wide, so the tests must be serialized
func TestAccSNSSMSPreferences_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		"almostAll":      testAccSMSPreferences_almostAll,
		"defaultSMSType": testAccSMSPreferences_defaultSMSType,
		"deliveryRole":   testAccSMSPreferences_deliveryRole,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccSMSPreferences_defaultSMSType(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sns_sms_preferences.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMSPreferencesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMSPreferencesConfig_defType,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(resourceName, "default_sender_id"),
					resource.TestCheckResourceAttr(resourceName, "default_sms_type", "Transactional"),
					resource.TestCheckNoResourceAttr(resourceName, "delivery_status_iam_role_arn"),
					resource.TestCheckNoResourceAttr(resourceName, "delivery_status_success_sampling_rate"),
					resource.TestCheckNoResourceAttr(resourceName, "monthly_spend_limit"),
					resource.TestCheckNoResourceAttr(resourceName, "usage_report_s3_bucket"),
				),
			},
		},
	})
}

func testAccSMSPreferences_almostAll(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sns_sms_preferences.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMSPreferencesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMSPreferencesConfig_almostAll,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_sms_type", "Transactional"),
					resource.TestCheckResourceAttr(resourceName, "monthly_spend_limit", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "usage_report_s3_bucket", "some-bucket"),
				),
			},
		},
	})
}

func testAccSMSPreferences_deliveryRole(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sns_sms_preferences.test"
	iamRoleName := "aws_iam_role.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSMSPreferencesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSMSPreferencesConfig_deliveryRole(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "delivery_status_iam_role_arn", iamRoleName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "delivery_status_success_sampling_rate", "75"),
				),
			},
		},
	})
}

func testAccCheckSMSPreferencesDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sns_sms_preferences" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).SNSClient(ctx)

			attrs, err := conn.GetSMSAttributes(ctx, &sns.GetSMSAttributesInput{})

			if err != nil {
				return err
			}

			if attrs == nil || len(attrs.Attributes) == 0 {
				return nil
			}

			var errs []error

			// The API is returning undocumented keys, e.g. "UsageReportS3Enabled". Only check the keys we're aware of.
			for _, snsAttrName := range tfsns.SMSPreferencesAttributeMap.APIAttributeNames() {
				v := attrs.Attributes[snsAttrName]
				if snsAttrName != "MonthlySpendLimit" {
					if v != "" {
						errs = append(errs, fmt.Errorf("expected SMS attribute %q to be empty, but received: %q", snsAttrName, v))
					}
				}
			}

			return errors.Join(errs...)
		}

		return nil
	}
}

const testAccSMSPreferencesConfig_defType = `
resource "aws_sns_sms_preferences" "test" {
  default_sms_type = "Transactional"
}
`

const testAccSMSPreferencesConfig_almostAll = `
resource "aws_sns_sms_preferences" "test" {
  monthly_spend_limit    = "1"
  default_sms_type       = "Transactional"
  usage_report_s3_bucket = "some-bucket"
}
`

func testAccSMSPreferencesConfig_deliveryRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_sms_preferences" "test" {
  delivery_status_iam_role_arn          = aws_iam_role.test.arn
  delivery_status_success_sampling_rate = "75"
}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "sns.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:PutMetricFilter",
        "logs:PutRetentionPolicy"
      ],
      "Resource": "*",
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}
`, rName)
}
