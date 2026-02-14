// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lexmodels_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflexmodels "github.com/hashicorp/terraform-provider-aws/internal/service/lexmodels"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLexModelsBotAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotAliasOutput
	resourceName := "aws_lex_bot_alias.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	testBotAliasName := "test_bot_alias" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotAliasDestroy(ctx, t, testBotAliasName, testBotAliasName),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_basic(testBotName),
					testAccBotAliasConfig_basic(testBotAliasName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotAliasExists(ctx, t, resourceName, &v),

					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "lex", "bot:{bot_name}:{name}"),
					resource.TestCheckResourceAttrSet(resourceName, "checksum"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Testing lex bot alias create."),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttr(resourceName, "bot_name", testBotName),
					resource.TestCheckResourceAttr(resourceName, "bot_version", tflexmodels.BotVersionLatest),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, testBotAliasName),
					resource.TestCheckResourceAttr(resourceName, "conversation_logs.#", "0"),
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

func testAccBotAlias_botVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotAliasOutput
	resourceName := "aws_lex_bot_alias.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	testBotAliasName := "test_bot_alias" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	// If this test runs in parallel with other Lex Bot tests, it loses its description
	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotAliasDestroy(ctx, t, testBotAliasName, testBotAliasName),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_basic(testBotName),
					testAccBotAliasConfig_basic(testBotAliasName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotAliasExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "bot_version", tflexmodels.BotVersionLatest),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrCreatedDate},
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_createVersion(testBotAliasName),
					testAccBotAliasConfig_botVersionUpdate(testBotAliasName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotAliasExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "bot_version", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrCreatedDate},
			},
		},
	})
}

func TestAccLexModelsBotAlias_conversationLogsText(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotAliasOutput
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	testBotAliasName := "test_bot_alias" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resourceName := "aws_lex_bot_alias.test"
	iamRoleResourceName := "aws_iam_role.test"
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotAliasDestroy(ctx, t, testBotID, testBotAliasName),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_basic(testBotID),
					testAccBotAliasConfig_conversationLogsText(testBotAliasName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotAliasExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "bot_version", tflexmodels.BotVersionLatest),
					resource.TestCheckResourceAttrPair(resourceName, "conversation_logs.0.iam_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "conversation_logs.0.log_settings.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "conversation_logs.0.log_settings.*", map[string]string{
						names.AttrDestination: "CLOUDWATCH_LOGS",
						"log_type":            "TEXT",
						names.AttrKMSKeyARN:   "",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "conversation_logs.0.log_settings.*.resource_arn", cloudwatchLogGroupResourceName, names.AttrARN),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "conversation_logs.0.log_settings.*", map[string]*regexp.Regexp{
						"resource_prefix": regexache.MustCompile(regexp.QuoteMeta(fmt.Sprintf(`aws/lex/%s/%s/%s/`, testBotID, testBotAliasName, tflexmodels.BotVersionLatest))),
					}),
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

func TestAccLexModelsBotAlias_conversationLogsAudio(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotAliasOutput
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	testBotAliasName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resourceName := "aws_lex_bot_alias.test"
	iamRoleResourceName := "aws_iam_role.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	kmsKeyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotAliasDestroy(ctx, t, testBotID, testBotAliasName),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_basic(testBotID),
					testAccBotAliasConfig_conversationLogsAudio(testBotAliasName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotAliasExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "bot_version", tflexmodels.BotVersionLatest),
					resource.TestCheckResourceAttrPair(resourceName, "conversation_logs.0.iam_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "conversation_logs.0.log_settings.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "conversation_logs.0.log_settings.*", map[string]string{
						names.AttrDestination: "S3",
						"log_type":            "AUDIO",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "conversation_logs.0.log_settings.*.resource_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "conversation_logs.0.log_settings.*.kms_key_arn", kmsKeyResourceName, names.AttrARN),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "conversation_logs.0.log_settings.*", map[string]*regexp.Regexp{
						"resource_prefix": regexache.MustCompile(regexp.QuoteMeta(fmt.Sprintf(`aws/lex/%s/%s/%s/`, testBotID, testBotAliasName, tflexmodels.BotVersionLatest))),
					}),
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

func TestAccLexModelsBotAlias_conversationLogsBoth(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotAliasOutput
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	testBotAliasName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resourceName := "aws_lex_bot_alias.test"
	iamRoleResourceName := "aws_iam_role.test"
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	kmsKeyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotAliasDestroy(ctx, t, testBotID, testBotAliasName),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_basic(testBotID),
					testAccBotAliasConfig_conversationLogsBoth(testBotAliasName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotAliasExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "bot_version", tflexmodels.BotVersionLatest),
					resource.TestCheckResourceAttrPair(resourceName, "conversation_logs.0.iam_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "conversation_logs.0.log_settings.#", "2"),

					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "conversation_logs.0.log_settings.*", map[string]string{
						names.AttrDestination: "CLOUDWATCH_LOGS",
						"log_type":            "TEXT",
						names.AttrKMSKeyARN:   "",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "conversation_logs.0.log_settings.*.resource_arn", cloudwatchLogGroupResourceName, names.AttrARN),

					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "conversation_logs.0.log_settings.*", map[string]string{
						names.AttrDestination: "S3",
						"log_type":            "AUDIO",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "conversation_logs.0.log_settings.*.resource_arn", s3BucketResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "conversation_logs.0.log_settings.*.kms_key_arn", kmsKeyResourceName, names.AttrARN),
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

func TestAccLexModelsBotAlias_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotAliasOutput
	resourceName := "aws_lex_bot_alias.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	testBotAliasName := "test_bot_alias" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotAliasDestroy(ctx, t, testBotAliasName, testBotAliasName),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_basic(testBotName),
					testAccBotAliasConfig_basic(testBotAliasName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotAliasExists(ctx, t, resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_basic(testBotName),
					testAccBotAliasConfig_descriptionUpdate(testBotAliasName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotAliasExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Testing lex bot alias update."),
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

func TestAccLexModelsBotAlias_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotAliasOutput
	resourceName := "aws_lex_bot_alias.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	testBotAliasName := "test_bot_alias" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_basic(testBotName),
					testAccBotAliasConfig_basic(testBotAliasName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotAliasExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tflexmodels.ResourceBotAlias(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBotAliasExists(ctx context.Context, t *testing.T, rName string, output *lexmodelbuildingservice.GetBotAliasOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Lex bot alias ID is set")
		}

		botName := rs.Primary.Attributes["bot_name"]
		botAliasName := rs.Primary.Attributes[names.AttrName]

		var err error
		conn := acctest.ProviderMeta(ctx, t).LexModelsClient(ctx)

		output, err = conn.GetBotAlias(ctx, &lexmodelbuildingservice.GetBotAliasInput{
			BotName: aws.String(botName),
			Name:    aws.String(botAliasName),
		})
		if errs.IsA[*awstypes.NotFoundException](err) {
			return fmt.Errorf("error bot alias '%q' not found", rs.Primary.ID)
		}
		if err != nil {
			return fmt.Errorf("error getting bot alias '%q': %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckBotAliasDestroy(ctx context.Context, t *testing.T, botName, botAliasName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LexModelsClient(ctx)

		_, err := conn.GetBotAlias(ctx, &lexmodelbuildingservice.GetBotAliasInput{
			BotName: aws.String(botName),
			Name:    aws.String(botAliasName),
		})

		if err != nil {
			if errs.IsA[*awstypes.NotFoundException](err) {
				return nil
			}

			return fmt.Errorf("error getting bot alias '%s': %w", botAliasName, err)
		}

		return fmt.Errorf("error bot alias still exists after delete, %s", botAliasName)
	}
}

func testAccBotAliasConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot_alias" "test" {
  bot_name    = aws_lex_bot.test.name
  bot_version = aws_lex_bot.test.version
  description = "Testing lex bot alias create."
  name        = "%s"
}
`, rName)
}

func testAccBotAliasConfig_botVersionUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot_alias" "test" {
  bot_name    = aws_lex_bot.test.name
  bot_version = "1"
  description = "Testing lex bot alias create."
  name        = "%s"
}
`, rName)
}

func testAccBotAliasConfig_conversationLogsText(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot_alias" "test" {
  bot_name    = aws_lex_bot.test.name
  bot_version = aws_lex_bot.test.version
  description = "Testing lex bot alias create."
  name        = "%[1]s"
  conversation_logs {
    iam_role_arn = aws_iam_role.test.arn
    log_settings {
      destination  = "CLOUDWATCH_LOGS"
      log_type     = "TEXT"
      resource_arn = aws_cloudwatch_log_group.test.arn
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = "%[1]s"
}

resource "aws_iam_role" "test" {
  name               = "%[1]s"
  assume_role_policy = data.aws_iam_policy_document.lex_assume_role_policy.json
}

data "aws_iam_policy_document" "lex_assume_role_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lex.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "lex_cloud_watch_logs_policy" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = [
      aws_cloudwatch_log_group.test.arn,
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  name   = "%[1]s"
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.lex_cloud_watch_logs_policy.json
}
`, rName)
}

func testAccBotAliasConfig_conversationLogsAudio(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot_alias" "test" {
  bot_name    = aws_lex_bot.test.name
  bot_version = aws_lex_bot.test.version
  description = "Testing lex bot alias create."
  name        = "%[1]s"
  conversation_logs {
    iam_role_arn = aws_iam_role.test.arn
    log_settings {
      destination  = "S3"
      log_type     = "AUDIO"
      resource_arn = aws_s3_bucket.test.arn
      kms_key_arn  = aws_kms_key.test.arn
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = "%[1]s"
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_iam_role" "test" {
  name               = "%[1]s"
  assume_role_policy = data.aws_iam_policy_document.lex_assume_role_policy.json
}

data "aws_iam_policy_document" "lex_assume_role_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lex.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "lex_s3_policy" {
  statement {
    effect = "Allow"
    actions = [
      "s3:PutObject",
    ]
    resources = [
      aws_s3_bucket.test.arn,
    ]
  }
}

resource "aws_iam_role_policy" "test" {
  name   = "%[1]s"
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.lex_s3_policy.json
}
`, rName)
}

func testAccBotAliasConfig_conversationLogsBoth(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot_alias" "test" {
  bot_name    = aws_lex_bot.test.name
  bot_version = aws_lex_bot.test.version
  description = "Testing lex bot alias create."
  name        = "%[1]s"
  conversation_logs {
    iam_role_arn = aws_iam_role.test.arn
    log_settings {
      destination  = "CLOUDWATCH_LOGS"
      log_type     = "TEXT"
      resource_arn = aws_cloudwatch_log_group.test.arn
    }
    log_settings {
      destination  = "S3"
      log_type     = "AUDIO"
      resource_arn = aws_s3_bucket.test.arn
      kms_key_arn  = aws_kms_key.test.arn
    }
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name = "%[1]s"
}

resource "aws_s3_bucket" "test" {
  bucket = "%[1]s"
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_iam_role" "test" {
  name               = "%[1]s"
  assume_role_policy = data.aws_iam_policy_document.lex_assume_role_policy.json
}

data "aws_iam_policy_document" "lex_assume_role_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lex.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "lex_cloud_watch_logs_policy" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = [
      aws_cloudwatch_log_group.test.arn,
    ]
  }
}

resource "aws_iam_role_policy" "lex_cloud_watch_logs_policy" {
  name   = "%[1]s-text"
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.lex_cloud_watch_logs_policy.json
}

data "aws_iam_policy_document" "lex_s3_policy" {
  statement {
    effect = "Allow"
    actions = [
      "s3:PutObject",
    ]
    resources = [
      aws_s3_bucket.test.arn,
    ]
  }
}

resource "aws_iam_role_policy" "lex_s3_policy" {
  name   = "%[1]s-audio"
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.lex_s3_policy.json
}
`, rName)
}

func testAccBotAliasConfig_descriptionUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot_alias" "test" {
  bot_name    = aws_lex_bot.test.name
  bot_version = aws_lex_bot.test.version
  description = "Testing lex bot alias update."
  name        = "%s"
}
`, rName)
}
