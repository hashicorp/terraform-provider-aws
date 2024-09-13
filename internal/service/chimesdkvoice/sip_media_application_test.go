// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chimesdkvoice_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/chimesdkvoice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfchimesdkvoice "github.com/hashicorp/terraform-provider-aws/internal/service/chimesdkvoice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccChimeSDKVoiceSipMediaApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var chimeSipMediaApplication *awstypes.SipMediaApplication

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkvoice_sip_media_application.test"
	lambdaFunctionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSipMediaApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSipMediaApplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSipMediaApplicationExists(ctx, resourceName, chimeSipMediaApplication),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "aws_region"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "endpoints.0.lambda_arn", lambdaFunctionResourceName, names.AttrARN),
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

func TestAccChimeSDKVoiceSipMediaApplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var chimeSipMediaApplication *awstypes.SipMediaApplication

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkvoice_sip_media_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSipMediaApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSipMediaApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSipMediaApplicationExists(ctx, resourceName, chimeSipMediaApplication),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfchimesdkvoice.ResourceSipMediaApplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccChimeSDKVoiceSipMediaApplication_update(t *testing.T) {
	ctx := acctest.Context(t)
	var chimeSipMediaApplication *awstypes.SipMediaApplication

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkvoice_sip_media_application.test"
	lambdaFunctionResourceName := "aws_lambda_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSipMediaApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSipMediaApplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSipMediaApplicationExists(ctx, resourceName, chimeSipMediaApplication),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "aws_region"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "endpoints.0.lambda_arn", lambdaFunctionResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccSipMediaApplicationConfig_basic(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSipMediaApplicationExists(ctx, resourceName, chimeSipMediaApplication),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "aws_region"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttrPair(resourceName, "endpoints.0.lambda_arn", lambdaFunctionResourceName, names.AttrARN),
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

func TestAccChimeSDKVoiceSipMediaApplication_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var sipMediaApplication *awstypes.SipMediaApplication

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_chimesdkvoice_sip_media_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ChimeSDKVoiceServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSipMediaApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSipMediaApplicationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSipMediaApplicationExists(ctx, resourceName, sipMediaApplication),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSipMediaApplicationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSipMediaApplicationExists(ctx, resourceName, sipMediaApplication),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccSipMediaApplicationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSipMediaApplicationExists(ctx, resourceName, sipMediaApplication),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckSipMediaApplicationExists(ctx context.Context, name string, vc *awstypes.SipMediaApplication) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ChimeSdkVoice Sip Media Application ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

		resp, err := tfchimesdkvoice.FindSIPResourceWithRetry(ctx, false, func() (*awstypes.SipMediaApplication, error) {
			return tfchimesdkvoice.FindSIPMediaApplicationByID(ctx, conn, rs.Primary.ID)
		})

		if err != nil {
			return err
		}

		vc = resp

		return nil
	}
}

func testAccCheckSipMediaApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_chimesdkvoice_sip_media_application" {
				continue
			}
			conn := acctest.Provider.Meta().(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

			_, err := tfchimesdkvoice.FindSIPResourceWithRetry(ctx, false, func() (*awstypes.SipMediaApplication, error) {
				return tfchimesdkvoice.FindSIPMediaApplicationByID(ctx, conn, rs.Primary.ID)
			})

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("sip media application still exists: (%s)", rs.Primary.ID)
		}

		return nil
	}
}

func testAccSipMediaApplicationConfigBase(rName string) string {
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

resource "aws_lambda_function" "test" {
  filename         = "test-fixtures/lambdatest.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambdatest.zip")
  function_name    = %[1]q
  role             = aws_iam_role.test.arn
  runtime          = "nodejs16.x"
  handler          = "index.handler"
}
`, rName)
}

func testAccSipMediaApplicationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccSipMediaApplicationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_chimesdkvoice_sip_media_application" "test" {
  name       = %[1]q
  aws_region = data.aws_region.current.name
  endpoints {
    lambda_arn = aws_lambda_function.test.arn
  }
}
`, rName))
}

func testAccSipMediaApplicationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccSipMediaApplicationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_chimesdkvoice_sip_media_application" "test" {
  name       = %[1]q
  aws_region = data.aws_region.current.name
  endpoints {
    lambda_arn = aws_lambda_function.test.arn
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccSipMediaApplicationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccSipMediaApplicationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_chimesdkvoice_sip_media_application" "test" {
  name       = %[1]q
  aws_region = data.aws_region.current.name
  endpoints {
    lambda_arn = aws_lambda_function.test.arn
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
