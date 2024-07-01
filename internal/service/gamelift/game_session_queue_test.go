// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfgamelift "github.com/hashicorp/terraform-provider-aws/internal/service/gamelift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const testAccGameSessionQueuePrefix = "tfAccQueue-"

func TestAccGameLiftGameSessionQueue_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf gamelift.GameSessionQueue

	resourceName := "aws_gamelift_game_session_queue.test"
	queueName := testAccGameSessionQueuePrefix + sdkacctest.RandString(8)
	playerLatencyPolicies := []gamelift.PlayerLatencyPolicy{
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int64(100),
			PolicyDurationSeconds:                      aws.Int64(5),
		},
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int64(200),
			PolicyDurationSeconds:                      nil,
		},
	}
	timeoutInSeconds := int64(124)

	uQueueName := queueName + "-updated"
	uPlayerLatencyPolicies := []gamelift.PlayerLatencyPolicy{
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int64(150),
			PolicyDurationSeconds:                      aws.Int64(10),
		},
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int64(250),
			PolicyDurationSeconds:                      nil,
		},
	}
	uTimeoutInSeconds := int64(600)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, gamelift.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGameSessionQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGameSessionQueueConfig_basic(queueName,
					playerLatencyPolicies, timeoutInSeconds, "Custom Event Data"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`gamesessionqueue/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, queueName),
					resource.TestCheckResourceAttr(resourceName, "destinations.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "notification_target", ""),
					resource.TestCheckResourceAttr(resourceName, "custom_event_data", "Custom Event Data"),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.0.maximum_individual_player_latency_milliseconds",
						fmt.Sprintf("%d", *playerLatencyPolicies[0].MaximumIndividualPlayerLatencyMilliseconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.0.policy_duration_seconds",
						fmt.Sprintf("%d", *playerLatencyPolicies[0].PolicyDurationSeconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.1.maximum_individual_player_latency_milliseconds",
						fmt.Sprintf("%d", *playerLatencyPolicies[1].MaximumIndividualPlayerLatencyMilliseconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.1.policy_duration_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timeout_in_seconds", fmt.Sprintf("%d", timeoutInSeconds)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccGameSessionQueueConfig_basic(uQueueName, uPlayerLatencyPolicies, uTimeoutInSeconds, "Custom Event Data"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(ctx, resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "gamelift", regexache.MustCompile(`gamesessionqueue/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, uQueueName),
					resource.TestCheckResourceAttr(resourceName, "destinations.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "notification_target", ""),
					resource.TestCheckResourceAttr(resourceName, "custom_event_data", "Custom Event Data"),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.0.maximum_individual_player_latency_milliseconds",
						fmt.Sprintf("%d", *uPlayerLatencyPolicies[0].MaximumIndividualPlayerLatencyMilliseconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.0.policy_duration_seconds",
						fmt.Sprintf("%d", *uPlayerLatencyPolicies[0].PolicyDurationSeconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.1.maximum_individual_player_latency_milliseconds",
						fmt.Sprintf("%d", *uPlayerLatencyPolicies[1].MaximumIndividualPlayerLatencyMilliseconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.1.policy_duration_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "timeout_in_seconds", fmt.Sprintf("%d", uTimeoutInSeconds)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	var conf gamelift.GameSessionQueue

	resourceName := "aws_gamelift_game_session_queue.test"
	queueName := testAccGameSessionQueuePrefix + sdkacctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, gamelift.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGameSessionQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGameSessionQueueConfig_basicTags1(queueName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(ctx, resourceName, &conf),
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
				Config: testAccGameSessionQueueConfig_basicTags2(queueName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccGameSessionQueueConfig_basicTags1(queueName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGameLiftGameSessionQueue_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf gamelift.GameSessionQueue

	resourceName := "aws_gamelift_game_session_queue.test"
	queueName := testAccGameSessionQueuePrefix + sdkacctest.RandString(8)
	playerLatencyPolicies := []gamelift.PlayerLatencyPolicy{
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int64(100),
			PolicyDurationSeconds:                      aws.Int64(5),
		},
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int64(200),
			PolicyDurationSeconds:                      nil,
		},
	}
	timeoutInSeconds := int64(124)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, gamelift.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GameLiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGameSessionQueueDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGameSessionQueueConfig_basic(queueName,
					playerLatencyPolicies, timeoutInSeconds, "Custom Event Data"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfgamelift.ResourceGameSessionQueue(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGameSessionQueueExists(ctx context.Context, n string, v *gamelift.GameSessionQueue) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No GameLift Game Session Queue ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn(ctx)

		output, err := tfgamelift.FindGameSessionQueueByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckGameSessionQueueDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_gamelift_game_session_queue" {
				continue
			}

			_, err := tfgamelift.FindGameSessionQueueByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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
	playerLatencyPolicies []gamelift.PlayerLatencyPolicy, timeoutInSeconds int64, customEventData string) string {
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
