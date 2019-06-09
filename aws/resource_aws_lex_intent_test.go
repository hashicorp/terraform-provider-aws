package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// Intents can accept a custom SlotType but it must be removed from the Intent before the SlotType can be deleted.
// This means we cannot reference the SlotType in the Intent with interpolation because Terraform will try to delete
// the SlotType first which will fail. So we do not have a test for custom slot types.

func TestAccAwsLexIntent(t *testing.T) {
	resourceName := "aws_lex_intent.test"
	testID := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	testIntentID := "test_intent_" + testID

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexIntentDestroy(testIntentID, "$LATEST"),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccAwsLexIntentMinConfig, testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(testIntentID, "$LATEST"),

					// user provided attributes
					resource.TestCheckResourceAttr(resourceName, "fulfillment_activity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_activity.0.type", "ReturnIntent"),
					resource.TestCheckResourceAttr(resourceName, "name", testIntentID),

					// computed attributes
					resource.TestCheckResourceAttrSet(resourceName, "checksum"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(testAccAwsLexIntentUpdateWithConclusionConfig, testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// user updated attributes
					resource.TestCheckResourceAttr(resourceName, "name", testIntentID),
					resource.TestCheckResourceAttr(resourceName, "conclusion_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "conclusion_statement.0.message.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "confirmation_prompt.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "confirmation_prompt.0.max_attempts", "2"),
					resource.TestCheckResourceAttr(resourceName, "confirmation_prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", "Intent to order a bouquet of flowers for pick up"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_activity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_activity.0.type", "ReturnIntent"),
					resource.TestCheckResourceAttr(resourceName, "rejection_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rejection_statement.0.message.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sample_utterances.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "slot.#", "1"),

					// computed attributes
					resource.TestCheckResourceAttrSet(resourceName, "checksum"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			{
				Config: fmt.Sprintf(testAccAwsLexIntentUpdateWithConclusionSlotsConfig, testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// user updated attributes
					resource.TestCheckResourceAttr(resourceName, "name", testIntentID),
					resource.TestCheckResourceAttr(resourceName, "conclusion_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "conclusion_statement.0.message.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "confirmation_prompt.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "confirmation_prompt.0.max_attempts", "2"),
					resource.TestCheckResourceAttr(resourceName, "confirmation_prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", "Intent to order a bouquet of flowers for pick up"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_activity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_activity.0.type", "ReturnIntent"),
					resource.TestCheckResourceAttr(resourceName, "rejection_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rejection_statement.0.message.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sample_utterances.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "slot.#", "2"),

					// computed attributes
					resource.TestCheckResourceAttrSet(resourceName, "checksum"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			{
				Config: fmt.Sprintf(testAccAwsLexIntentUpdateWithFollowUpConfig, testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// user provided attributes
					resource.TestCheckResourceAttr(resourceName, "name", testIntentID),
					resource.TestCheckResourceAttr(resourceName, "confirmation_prompt.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "confirmation_prompt.0.max_attempts", "2"),
					resource.TestCheckResourceAttr(resourceName, "confirmation_prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", "Intent to order a bouquet of flowers for pick up"),
					resource.TestCheckResourceAttr(resourceName, "follow_up_prompt.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "follow_up_prompt.0.prompt.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "follow_up_prompt.0.prompt.0.max_attempts", "2"),
					resource.TestCheckResourceAttr(resourceName, "follow_up_prompt.0.prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "follow_up_prompt.0.rejection_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "follow_up_prompt.0.rejection_statement.0.message.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_activity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_activity.0.type", "ReturnIntent"),
					resource.TestCheckResourceAttr(resourceName, "rejection_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rejection_statement.0.message.#", "1"),

					// computed attributes
					resource.TestCheckResourceAttrSet(resourceName, "checksum"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
		},
	})
}

func testAccCheckAwsLexIntentExists(intentName, intentVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		_, err := conn.GetIntent(&lexmodelbuildingservice.GetIntentInput{
			Name:    aws.String(intentName),
			Version: aws.String(intentVersion),
		})
		if err != nil {
			if isAWSErr(err, "NotFoundException", "") {
				return fmt.Errorf("error intent %s not found, %s", intentName, err)
			}

			return fmt.Errorf("error getting intent %s: %s", intentName, err)
		}

		return nil
	}
}

func testAccCheckAwsLexIntentDestroy(intentName, intentVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		_, err := conn.GetIntent(&lexmodelbuildingservice.GetIntentInput{
			Name:    aws.String(intentName),
			Version: aws.String("$LATEST"),
		})

		if err != nil {
			if isAWSErr(err, "NotFoundException", "") {
				return nil
			}

			return fmt.Errorf("error getting intent %s: %s", intentName, err)
		}

		return fmt.Errorf("error intent still exists after delete, %s", intentName)
	}
}

const testAccAwsLexIntentMinConfig = `
resource "aws_lex_intent" "test" {
  fulfillment_activity {
    type = "ReturnIntent"
  }

  name = "test_intent_%[1]s"
}
`

// with conclusion statement

const testAccAwsLexIntentUpdateWithConclusionConfig = `
resource "aws_lex_intent" "test" {
  conclusion_statement {
    message {
      content      = "Your order for {FlowerType} has been placed and will be ready by {PickupTime} on {PickupDate}"
      content_type = "PlainText"
    }
  }

  confirmation_prompt {
    max_attempts = 2

    message {
      content      = "Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}. Does this sound okay?"
      content_type = "PlainText"
    }
  }

  description = "Intent to order a bouquet of flowers for pick up"

  // requires a lambda function to test
  // dialog_code_hook {
  //   message_version = "1"
  //   uri             = "arn:aws:lambda:us-east-1:123456789012:function:RetrieveAvailableFlowers"
  // }

  fulfillment_activity {
    // requires a lambda function to test
    // code_hook {
    //   message_version = "1"
    //   uri             = "arn:aws:lambda:us-east-1:123456789012:function:ProcessFlowerOrder"
    // }

    type = "ReturnIntent"
  }
  name = "test_intent_%[1]s"
  rejection_statement {
    message {
      content      = "Okay, I will not place your order."
      content_type = "PlainText"
    }
  }
  sample_utterances = [
    "I would like to order some flowers",
    "I would like to pick up flowers",
  ]
  slot {
    description = "The date to pick up the flowers"
    name        = "PickupDate"
    priority    = 2

    sample_utterances = [
      "I would like to order {FlowerType}",
    ]

    slot_constraint = "Required"
    slot_type       = "AMAZON.DATE"

    value_elicitation_prompt {
      max_attempts = 2

      message {
        content      = "What day do you want the {FlowerType} to be picked up?"
        content_type = "PlainText"
      }
    }
  }
}
`

// with conclusion update slots

const testAccAwsLexIntentUpdateWithConclusionSlotsConfig = `
resource "aws_lex_intent" "test" {
  conclusion_statement {
    message {
      content      = "Your order for {FlowerType} has been placed and will be ready by {PickupTime} on {PickupDate}"
      content_type = "PlainText"
    }
  }

  confirmation_prompt {
    max_attempts = 2

    message {
      content      = "Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}. Does this sound okay?"
      content_type = "PlainText"
    }
  }

  description = "Intent to order a bouquet of flowers for pick up"

  fulfillment_activity {
    type = "ReturnIntent"
  }

  name = "test_intent_%[1]s"

  rejection_statement {
    message {
      content      = "Okay, I will not place your order."
      content_type = "PlainText"
    }
  }

  sample_utterances = [
    "I would like to order some flowers",
    "I would like to pick up flowers",
  ]

  slot {
    description = "The date to pick up the flowers"
    name        = "PickupDate"
    priority    = 2

    sample_utterances = [
      "I would like to order {FlowerType}",
    ]

    slot_constraint = "Required"
    slot_type       = "AMAZON.DATE"

    value_elicitation_prompt {
      max_attempts = 2

      message {
        content      = "What day do you want the {FlowerType} to be picked up?"
        content_type = "PlainText"
      }
    }
  }

  slot {
    description = "The time to pick up the flowers"
    name        = "PickupTime"
    priority    = 3

    sample_utterances = [
      "I would like to order {FlowerType}",
    ]

    slot_constraint = "Required"
    slot_type       = "AMAZON.TIME"

    value_elicitation_prompt {
      max_attempts = 2

      message {
        content      = "Pick up the {FlowerType} at what time on {PickupDate}?"
        content_type = "PlainText"
      }
    }
  }
}
`

// with follow up prompt

const testAccAwsLexIntentUpdateWithFollowUpConfig = `
resource "aws_lex_intent" "test" {
  confirmation_prompt {
    max_attempts = 2

    message {
      content      = "Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}. Does this sound okay?"
      content_type = "PlainText"
    }
  }

  description = "Intent to order a bouquet of flowers for pick up"

  follow_up_prompt {
    prompt {
      max_attempts = 2

      message {
        content      = "Would you like to place another order?"
        content_type = "PlainText"
      }
    }

    rejection_statement {
      message {
        content      = "Your order for {FlowerType} has been placed and will be ready by {PickupTime} on {PickupDate}"
        content_type = "PlainText"
      }
    }
  }

  fulfillment_activity {
    type = "ReturnIntent"
  }

  name = "test_intent_%[1]s"

  rejection_statement {
    message {
      content      = "Okay, I will not place your order."
      content_type = "PlainText"
    }
  }
}
`
