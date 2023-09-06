// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chimesdkvoice_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/aws/aws-sdk-go/service/chimesdkvoice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfchimesdkvoice "github.com/hashicorp/terraform-provider-aws/internal/service/chimesdkvoice"
)

func TestAccChimeSDKVoiceSipRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var chimeSipRule *chimesdkvoice.SipRule

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkvoice_sip_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, chimesdkvoice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSipRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSipRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSipRuleExists(ctx, resourceName, chimeSipRule),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "trigger_type", "RequestUriHostname"),
					resource.TestCheckResourceAttrSet(resourceName, "trigger_value"),
					resource.TestCheckResourceAttr(resourceName, "target_applications.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_applications.0.priority", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "target_applications.0.sip_media_application_id"),
					resource.TestCheckResourceAttrSet(resourceName, "target_applications.0.aws_region"),
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

func TestAccChimeSDKVoiceSipRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var chimeSipRule *chimesdkvoice.SipRule

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkvoice_sip_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, chime.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSipRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSipRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSipRuleExists(ctx, resourceName, chimeSipRule),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfchimesdkvoice.ResourceSipRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccChimeSDKVoiceSipRule_update(t *testing.T) {
	ctx := acctest.Context(t)
	var chimeSipRule *chimesdkvoice.SipRule

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkvoice_sip_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, chimesdkvoice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSipRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSipRuleConfig_update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSipRuleExists(ctx, resourceName, chimeSipRule),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "trigger_type", "RequestUriHostname"),
					resource.TestCheckResourceAttrSet(resourceName, "trigger_value"),
					resource.TestCheckResourceAttr(resourceName, "target_applications.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_applications.0.priority", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "target_applications.0.sip_media_application_id"),
					resource.TestCheckResourceAttrSet(resourceName, "target_applications.0.aws_region"),
				),
			},
			{
				Config: testAccSipRuleConfig_update(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSipRuleExists(ctx, resourceName, chimeSipRule),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "disabled", "true"),
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

func testAccCheckSipRuleExists(ctx context.Context, name string, sr *chimesdkvoice.SipRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ChimeSdkVoice Sip Rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceConn(ctx)
		input := &chimesdkvoice.GetSipRuleInput{
			SipRuleId: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetSipRuleWithContext(ctx, input)
		if err != nil {
			return err
		}

		sr = resp.SipRule

		return nil
	}
}

func testAccCheckSipRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chimesdkvoice_sip_rule" {
				continue
			}
			conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceConn(ctx)
			input := &chimesdkvoice.GetSipRuleInput{
				SipRuleId: aws.String(rs.Primary.ID),
			}
			resp, err := conn.GetSipRuleWithContext(ctx, input)
			if err == nil {
				if resp.SipRule != nil && aws.StringValue(resp.SipRule.Name) != "" {
					return fmt.Errorf("error ChimeSdkVoice Sip Rule still exists")
				}
			}
			return nil
		}
		return nil
	}
}

func testAccSipRuleConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_chime_voice_connector" "test" {
  name               = %[1]q
  require_encryption = true
}

resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  function_name    = %[1]q
  role             = aws_iam_role.test.arn
  runtime          = "nodejs16.x"
  handler          = "index.handler"
}

resource "aws_chimesdkvoice_sip_media_application" "test" {
  name       = %[1]q
  aws_region = data.aws_region.current.name
  endpoints {
    lambda_arn = aws_lambda_function.test.arn
  }
}

`, rName)
}

func testAccSipRuleConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccSipRuleConfigBase(rName),
		fmt.Sprintf(`
resource "aws_chimesdkvoice_sip_rule" "test" {
  name          = %[1]q
  disabled      = "false"
  trigger_type  = "RequestUriHostname"
  trigger_value = aws_chime_voice_connector.test.outbound_host_name
  target_applications {
    priority                 = 1
    sip_media_application_id = aws_chimesdkvoice_sip_media_application.test.id
    aws_region               = data.aws_region.current.name
  }
}
`, rName))
}

func testAccSipRuleConfig_update(rName string) string {
	return acctest.ConfigCompose(
		testAccSipRuleConfigBase(rName),
		fmt.Sprintf(`
resource "aws_chimesdkvoice_sip_rule" "test" {
  name          = %[1]q
  disabled      = "true"
  trigger_type  = "RequestUriHostname"
  trigger_value = aws_chime_voice_connector.test.outbound_host_name
  target_applications {
    priority                 = 1
    sip_media_application_id = aws_chimesdkvoice_sip_media_application.test.id
    aws_region               = data.aws_region.current.name
  }
}
`, rName))
}
