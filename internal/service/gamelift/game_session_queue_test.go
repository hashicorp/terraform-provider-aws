// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package gamelift_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfgamelift "github.com/hashicorp/terraform-provider-aws/internal/service/gamelift"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const testAccGameSessionQueuePrefix = "tfAccQueue-"

func TestAccGameLiftGameSessionQueue_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.GameSessionQueue

	resourceName := "aws_gamelift_game_session_queue.test"
	queueName := testAccGameSessionQueuePrefix + sdkacctest.RandString(8)
	playerLatencyPolicies := []awstypes.PlayerLatencyPolicy{
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int32(100),
			PolicyDurationSeconds:                      aws.Int32(5),
		},
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int32(200),
			PolicyDurationSeconds:                      nil,
		},
	}
	timeoutInSeconds := int64(124)

	uQueueName := queueName + "-updated"
	uPlayerLatencyPolicies := []awstypes.PlayerLatencyPolicy{
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int32(150),
			PolicyDurationSeconds:                      aws.Int32(10),
		},
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int32(250),
			PolicyDurationSeconds:                      nil,
		},
	}
	uTimeoutInSeconds := int64(600)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGameSessionQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGameSessionQueueConfig_basic(queueName,
					playerLatencyPolicies, timeoutInSeconds, "Custom Event Data"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(ctx, t, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`gamesessionqueue/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, queueName),
					resource.TestCheckResourceAttr(resourceName, "destinations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "notification_target", ""),
					resource.TestCheckResourceAttr(resourceName, "custom_event_data", "Custom Event Data"),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.0.maximum_individual_player_latency_milliseconds",
						strconv.Itoa(int(aws.ToInt32(playerLatencyPolicies[0].MaximumIndividualPlayerLatencyMilliseconds)))),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.0.policy_duration_seconds",
						strconv.Itoa(int(aws.ToInt32(playerLatencyPolicies[0].PolicyDurationSeconds)))),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.1.maximum_individual_player_latency_milliseconds",
						strconv.Itoa(int(aws.ToInt32(playerLatencyPolicies[1].MaximumIndividualPlayerLatencyMilliseconds)))),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.1.policy_duration_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout_in_seconds", strconv.Itoa(int(timeoutInSeconds))),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				Config: testAccGameSessionQueueConfig_basic(uQueueName, uPlayerLatencyPolicies, uTimeoutInSeconds, "Custom Event Data"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(ctx, t, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`gamesessionqueue/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, uQueueName),
					resource.TestCheckResourceAttr(resourceName, "destinations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "notification_target", ""),
					resource.TestCheckResourceAttr(resourceName, "custom_event_data", "Custom Event Data"),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.0.maximum_individual_player_latency_milliseconds",
						strconv.Itoa(int(aws.ToInt32(uPlayerLatencyPolicies[0].MaximumIndividualPlayerLatencyMilliseconds)))),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.0.policy_duration_seconds",
						strconv.Itoa(int(aws.ToInt32(uPlayerLatencyPolicies[0].PolicyDurationSeconds)))),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.1.maximum_individual_player_latency_milliseconds",
						strconv.Itoa(int(aws.ToInt32(uPlayerLatencyPolicies[1].MaximumIndividualPlayerLatencyMilliseconds)))),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.1.policy_duration_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout_in_seconds", strconv.Itoa(int(uTimeoutInSeconds))),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccGameLiftGameSessionQueue_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.GameSessionQueue

	resourceName := "aws_gamelift_game_session_queue.test"
	queueName := testAccGameSessionQueuePrefix + sdkacctest.RandString(8)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGameSessionQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGameSessionQueueConfig_basicTags1(queueName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGameSessionQueueConfig_basicTags2(queueName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccGameSessionQueueConfig_basicTags1(queueName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(ctx, t, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGameLiftGameSessionQueue_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.GameSessionQueue

	resourceName := "aws_gamelift_game_session_queue.test"
	queueName := testAccGameSessionQueuePrefix + sdkacctest.RandString(8)
	playerLatencyPolicies := []awstypes.PlayerLatencyPolicy{
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int32(100),
			PolicyDurationSeconds:                      aws.Int32(5),
		},
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int32(200),
			PolicyDurationSeconds:                      nil,
		},
	}
	timeoutInSeconds := int64(124)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.GameLiftEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGameSessionQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGameSessionQueueConfig_basic(queueName,
					playerLatencyPolicies, timeoutInSeconds, "Custom Event Data"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(ctx, t, resourceName, &conf),
					acctest.CheckSDKResourceDisappears(ctx, t, tfgamelift.ResourceGameSessionQueue(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGameSessionQueueExists(ctx context.Context, t *testing.T, n string, v *awstypes.GameSessionQueue) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GameLiftClient(ctx)

		output, err := tfgamelift.FindGameSessionQueueByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGameSessionQueueDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).GameLiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_gamelift_game_session_queue" {
				continue
			}

			_, err := tfgamelift.FindGameSessionQueueByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("GameLift Game Session Queue %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccGameSessionQueueConfig_basic(queueName string,
	playerLatencyPolicies []awstypes.PlayerLatencyPolicy, timeoutInSeconds int64, customEventData string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_game_session_queue" "test" {
  name         = "%s"
  destinations = []

  player_latency_policy {
    maximum_individual_player_latency_milliseconds = %d
    policy_duration_seconds                        = %d
  }

  player_latency_policy {
    maximum_individual_player_latency_milliseconds = %d
  }

  timeout_in_seconds = %d

  custom_event_data = "%s"
}
`,
		queueName,
		*playerLatencyPolicies[0].MaximumIndividualPlayerLatencyMilliseconds,
		*playerLatencyPolicies[0].PolicyDurationSeconds,
		*playerLatencyPolicies[1].MaximumIndividualPlayerLatencyMilliseconds,
		timeoutInSeconds,
		customEventData)
}

func testAccGameSessionQueueConfig_basicTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_game_session_queue" "test" {
  name         = %[1]q
  destinations = []

  player_latency_policy {
    maximum_individual_player_latency_milliseconds = 100000
    policy_duration_seconds                        = 10
  }

  player_latency_policy {
    maximum_individual_player_latency_milliseconds = 100000
  }

  timeout_in_seconds = 10

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccGameSessionQueueConfig_basicTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_game_session_queue" "test" {
  name         = %[1]q
  destinations = []

  player_latency_policy {
    maximum_individual_player_latency_milliseconds = 100000
    policy_duration_seconds                        = 10
  }

  player_latency_policy {
    maximum_individual_player_latency_milliseconds = 100000
  }

  timeout_in_seconds = 10

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
