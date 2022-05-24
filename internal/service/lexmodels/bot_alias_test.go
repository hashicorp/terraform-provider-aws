package lexmodels_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflexmodels "github.com/hashicorp/terraform-provider-aws/internal/service/lexmodels"
)

func TestAccLexModelsBotAlias_basic(t *testing.T) {
	var v lexmodelbuildingservice.GetBotAliasOutput
	resourceName := "aws_lex_bot_alias.test"
	testBotAliasID := "test_bot_alias" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotAliasDestroy(testBotAliasID, testBotAliasID),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotAliasID),
					testAccBotConfig_basic(testBotAliasID),
					testAccBotAliasConfig_basic(testBotAliasID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotAliasExists(resourceName, &v),

					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "checksum"),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_date"),
					resource.TestCheckResourceAttr(resourceName, "description", "Testing lex bot alias create."),
					acctest.CheckResourceAttrRFC3339(resourceName, "last_updated_date"),
					resource.TestCheckResourceAttr(resourceName, "bot_name", testBotAliasID),
					resource.TestCheckResourceAttr(resourceName, "bot_version", tflexmodels.BotVersionLatest),
					resource.TestCheckResourceAttr(resourceName, "name", testBotAliasID),
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
	var v lexmodelbuildingservice.GetBotAliasOutput
	resourceName := "aws_lex_bot_alias.test"
	testBotAliasID := "test_bot_alias" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	// If this test runs in parallel with other Lex Bot tests, it loses its description
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotAliasDestroy(testBotAliasID, testBotAliasID),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotAliasID),
					testAccBotConfig_basic(testBotAliasID),
					testAccBotAliasConfig_basic(testBotAliasID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotAliasExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "bot_version", tflexmodels.BotVersionLatest),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_date"},
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotAliasID),
					testAccBotConfig_createVersion(testBotAliasID),
					testAccBotAliasConfig_botVersionUpdate(testBotAliasID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotAliasExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "bot_version", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created_date"},
			},
		},
	})
}

func TestAccLexModelsBotAlias_conversationLogsText(t *testing.T) {
	var v lexmodelbuildingservice.GetBotAliasOutput
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	testBotAliasID := "test_bot_alias" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resourceName := "aws_lex_bot_alias.test"
	iamRoleResourceName := "aws_iam_role.test"
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotAliasDestroy(testBotID, testBotAliasID),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_basic(testBotID),
					testAccBotAliasConfig_conversationLogsText(testBotAliasID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotAliasExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "bot_version", tflexmodels.BotVersionLatest),
					resource.TestCheckResourceAttrPair(resourceName, "conversation_logs.0.iam_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "conversation_logs.0.log_settings.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "conversation_logs.0.log_settings.*", map[string]string{
						"destination": "CLOUDWATCH_LOGS",
						"log_type":    "TEXT",
						"kms_key_arn": "",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "conversation_logs.0.log_settings.*.resource_arn", cloudwatchLogGroupResourceName, "arn"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "conversation_logs.0.log_settings.*", map[string]*regexp.Regexp{
						"resource_prefix": regexp.MustCompile(regexp.QuoteMeta(fmt.Sprintf(`aws/lex/%s/%s/%s/`, testBotID, testBotAliasID, tflexmodels.BotVersionLatest))),
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
	var v lexmodelbuildingservice.GetBotAliasOutput
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	testBotAliasID := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resourceName := "aws_lex_bot_alias.test"
	iamRoleResourceName := "aws_iam_role.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotAliasDestroy(testBotID, testBotAliasID),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_basic(testBotID),
					testAccBotAliasConfig_conversationLogsAudio(testBotAliasID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotAliasExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "bot_version", tflexmodels.BotVersionLatest),
					resource.TestCheckResourceAttrPair(resourceName, "conversation_logs.0.iam_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "conversation_logs.0.log_settings.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "conversation_logs.0.log_settings.*", map[string]string{
						"destination": "S3",
						"log_type":    "AUDIO",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "conversation_logs.0.log_settings.*.resource_arn", s3BucketResourceName, "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "conversation_logs.0.log_settings.*.kms_key_arn", kmsKeyResourceName, "arn"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "conversation_logs.0.log_settings.*", map[string]*regexp.Regexp{
						"resource_prefix": regexp.MustCompile(regexp.QuoteMeta(fmt.Sprintf(`aws/lex/%s/%s/%s/`, testBotID, testBotAliasID, tflexmodels.BotVersionLatest))),
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
	var v lexmodelbuildingservice.GetBotAliasOutput
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	testBotAliasID := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resourceName := "aws_lex_bot_alias.test"
	iamRoleResourceName := "aws_iam_role.test"
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	kmsKeyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotAliasDestroy(testBotID, testBotAliasID),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_basic(testBotID),
					testAccBotAliasConfig_conversationLogsBoth(testBotAliasID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotAliasExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "bot_version", tflexmodels.BotVersionLatest),
					resource.TestCheckResourceAttrPair(resourceName, "conversation_logs.0.iam_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "conversation_logs.0.log_settings.#", "2"),

					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "conversation_logs.0.log_settings.*", map[string]string{
						"destination": "CLOUDWATCH_LOGS",
						"log_type":    "TEXT",
						"kms_key_arn": "",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "conversation_logs.0.log_settings.*.resource_arn", cloudwatchLogGroupResourceName, "arn"),

					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "conversation_logs.0.log_settings.*", map[string]string{
						"destination": "S3",
						"log_type":    "AUDIO",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "conversation_logs.0.log_settings.*.resource_arn", s3BucketResourceName, "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "conversation_logs.0.log_settings.*.kms_key_arn", kmsKeyResourceName, "arn"),
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
	var v lexmodelbuildingservice.GetBotAliasOutput
	resourceName := "aws_lex_bot_alias.test"
	testBotAliasID := "test_bot_alias" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotAliasDestroy(testBotAliasID, testBotAliasID),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotAliasID),
					testAccBotConfig_basic(testBotAliasID),
					testAccBotAliasConfig_basic(testBotAliasID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotAliasExists(resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotAliasID),
					testAccBotConfig_basic(testBotAliasID),
					testAccBotAliasConfig_descriptionUpdate(testBotAliasID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotAliasExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "Testing lex bot alias update."),
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
	var v lexmodelbuildingservice.GetBotAliasOutput
	resourceName := "aws_lex_bot_alias.test"
	testBotAliasID := "test_bot_alias" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotAliasID),
					testAccBotConfig_basic(testBotAliasID),
					testAccBotAliasConfig_basic(testBotAliasID),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotAliasExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tflexmodels.ResourceBotAlias(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBotAliasExists(rName string, output *lexmodelbuildingservice.GetBotAliasOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Lex bot alias ID is set")
		}

		botName := rs.Primary.Attributes["bot_name"]
		botAliasName := rs.Primary.Attributes["name"]

		var err error
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelsConn

		output, err = conn.GetBotAlias(&lexmodelbuildingservice.GetBotAliasInput{
			BotName: aws.String(botName),
			Name:    aws.String(botAliasName),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return fmt.Errorf("error bot alias '%q' not found", rs.Primary.ID)
		}
		if err != nil {
			return fmt.Errorf("error getting bot alias '%q': %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckBotAliasDestroy(botName, botAliasName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelsConn

		_, err := conn.GetBotAlias(&lexmodelbuildingservice.GetBotAliasInput{
			BotName: aws.String(botName),
			Name:    aws.String(botAliasName),
		})

		if err != nil {
			if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
				return nil
			}

			return fmt.Errorf("error getting bot alias '%s': %s", botAliasName, err)
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

resource "aws_kms_key" "test" {}

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

resource "aws_kms_key" "test" {}

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
