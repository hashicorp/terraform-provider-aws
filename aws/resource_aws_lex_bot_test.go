package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsLexBot_basic(t *testing.T) {
	var v lexmodelbuildingservice.GetBotOutput
	rName := "aws_lex_bot.test"
	testBotID := "test_bot_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAwsLexBotConfig_intent(testBotID),
					testAccAwsLexBotConfig_basic(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexBotExists(rName, &v),
					testAccCheckAwsLexBotNotExists(testBotID, "1"),

					resource.TestCheckNoResourceAttr(rName, "abort_statement"),
					resource.TestCheckResourceAttrSet(rName, "arn"),
					resource.TestCheckResourceAttrSet(rName, "checksum"),
					resource.TestCheckResourceAttr(rName, "child_directed", "false"),
					resource.TestCheckNoResourceAttr(rName, "clarification_prompt"),
					resource.TestCheckResourceAttr(rName, "create_version", "false"),
					testAccCheckResourceAttrRfc3339(rName, "created_date"),
					resource.TestCheckResourceAttr(rName, "description", "Bot to order flowers on the behalf of a user"),
					resource.TestCheckResourceAttr(rName, "detect_sentiment", "false"),
					resource.TestCheckResourceAttr(rName, "enable_model_improvements", "false"),
					resource.TestCheckResourceAttr(rName, "failure_reason", ""),
					resource.TestCheckResourceAttr(rName, "idle_session_ttl_in_seconds", "300"),
					resource.TestCheckNoResourceAttr(rName, "intent"),
					testAccCheckResourceAttrRfc3339(rName, "last_updated_date"),
					resource.TestCheckResourceAttr(rName, "locale", "en-US"),
					resource.TestCheckResourceAttr(rName, "name", testBotID),
					resource.TestCheckResourceAttr(rName, "nlu_intent_confidence_threshold", "0"),
					resource.TestCheckResourceAttr(rName, "process_behavior", "SAVE"),
					resource.TestCheckResourceAttr(rName, "status", "NOT_BUILT"),
					resource.TestCheckResourceAttr(rName, "version", LexBotVersionLatest),
					resource.TestCheckNoResourceAttr(rName, "voice_id"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexBot_disappears(t *testing.T) {
	var v lexmodelbuildingservice.GetBotOutput
	rName := "aws_lex_bot.test"
	testBotID := "test_bot_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAwsLexBotConfig_intent(testBotID),
					testAccAwsLexBotConfig_basic(testBotID),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLexBotExists(rName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsLexBot(), rName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsLexBotExistsWithVersion(rName, botVersion string, output *lexmodelbuildingservice.GetBotOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Lex bot ID is set")
		}

		var err error
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		output, err = conn.GetBot(&lexmodelbuildingservice.GetBotInput{
			Name:           aws.String(rs.Primary.ID),
			VersionOrAlias: aws.String(botVersion),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return fmt.Errorf("error bot %q version %s not found", rs.Primary.ID, botVersion)
		}
		if err != nil {
			return fmt.Errorf("error getting bot %q version %s: %w", rs.Primary.ID, botVersion, err)
		}

		return nil
	}
}

func testAccCheckAwsLexBotExists(rName string, output *lexmodelbuildingservice.GetBotOutput) resource.TestCheckFunc {
	return testAccCheckAwsLexBotExistsWithVersion(rName, LexBotVersionLatest, output)
}

func testAccCheckAwsLexBotNotExists(botName, botVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		_, err := conn.GetBot(&lexmodelbuildingservice.GetBotInput{
			Name:           aws.String(botName),
			VersionOrAlias: aws.String(botVersion),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error getting bot %s version %s: %s", botName, botVersion, err)
		}

		return fmt.Errorf("error bot %s version %s exists", botName, botVersion)
	}
}

func testAccCheckAwsLexBotDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lex_bot" {
			continue
		}

		output, err := conn.GetBotVersions(&lexmodelbuildingservice.GetBotVersionsInput{
			Name: aws.String(rs.Primary.ID),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}

		if output == nil || len(output.Bots) == 0 {
			return nil
		}

		return fmt.Errorf("Lex bot %q still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAwsLexBotConfig_intent(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name           = "%s"
  create_version = true
  fulfillment_activity {
    type = "ReturnIntent"
  }
  sample_utterances = [
    "I would like to pick up flowers",
  ]
}
`, rName)
}

func testAccAwsLexBotConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  name           = "%s"
  description    = "Bot to order flowers on the behalf of a user"
  child_directed = false
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
    }
  }
  clarification_prompt {
    max_attempts = 2
    message {
      content      = "I didn't understand you, what would you like to do?"
      content_type = "PlainText"
    }
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
}
`, rName)
}

func testAccAwsLexBotConfig_createVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  name             = "%s"
  description      = "Bot to order flowers on the behalf of a user"
  child_directed   = false
  create_version   = true
  process_behavior = "BUILD"
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
    }
  }
  clarification_prompt {
    max_attempts = 2
    message {
      content      = "I didn't understand you, what would you like to do?"
      content_type = "PlainText"
    }
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
}
`, rName)
}
