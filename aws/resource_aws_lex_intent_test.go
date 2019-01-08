package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// Intents can accept a custom SlotType but it must be removed from the Intent before the SlotType can be deleted.
// This means we cannot reference the SlotType in the Intent with interpolation because Terraform will try to delete
// the SlotType first which will fail. So we do not have a test for custom slot types.

func TestAccLexIntent(t *testing.T) {
	resourceName := "aws_lex_intent.test"
	testId := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	testIntentId := "test_intent_" + testId

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkLexIntentDestroy(testIntentId),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testLexIntentMinConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					checkResourceStateComputedAttr(resourceName, resourceAwsLexIntent()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(testLexIntentUpdateWithConclusionConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrPrefixSet(resourceName, "conclusion_statement"),
					testCheckResourceAttrPrefixSet(resourceName, "confirmation_prompt"),
					testCheckResourceAttrPrefixSet(resourceName, "fulfillment_activity"),
					testCheckResourceAttrPrefixSet(resourceName, "rejection_statement"),
					testCheckResourceAttrPrefixSet(resourceName, "sample_utterances"),
					testCheckResourceAttrPrefixSet(resourceName, "slot"),

					resource.TestCheckResourceAttr(resourceName, "description", "Intent to order a bouquet of flowers for pick up"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),

					checkResourceStateComputedAttr(resourceName, resourceAwsLexIntent()),
				),
			},
			{
				Config: fmt.Sprintf(testLexIntentUpdateWithConclusionSlotsConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrPrefixSet(resourceName, "slot"),

					checkResourceStateComputedAttr(resourceName, resourceAwsLexIntent()),
				),
			},
			{
				Config: fmt.Sprintf(testLexIntentUpdateWithFollowUpConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrPrefixSet(resourceName, "confirmation_prompt"),
					testCheckResourceAttrPrefixSet(resourceName, "follow_up_prompt"),
					testCheckResourceAttrPrefixSet(resourceName, "fulfillment_activity"),
					testCheckResourceAttrPrefixSet(resourceName, "rejection_statement"),

					resource.TestCheckResourceAttr(resourceName, "description", "Intent to order a bouquet of flowers for pick up"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),

					checkResourceStateComputedAttr(resourceName, resourceAwsLexIntent()),
				),
			},
		},
	})
}

func checkLexIntentDestroy(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		_, err := conn.GetIntent(&lexmodelbuildingservice.GetIntentInput{
			Name:    aws.String(id),
			Version: aws.String("$LATEST"),
		})

		if err != nil {
			aerr, ok := err.(awserr.Error)
			if ok && aerr.Code() == "NotFoundException" {
				return nil
			}

			return fmt.Errorf("could not get intent, %s", id)
		}

		return fmt.Errorf("the intent still exists after delete, %s", id)
	}
}

const testLexIntentMinConfig = `
resource "aws_lex_intent" "test" {
  fulfillment_activity {
    type = "ReturnIntent"
  }

  name = "test_intent_%[1]s"
}
`

// with conclusion statement

const testLexIntentUpdateWithConclusionConfig = `
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

const testLexIntentUpdateWithConclusionSlotsConfig = `
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

const testLexIntentUpdateWithFollowUpConfig = `
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
