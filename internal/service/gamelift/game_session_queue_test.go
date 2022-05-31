package gamelift_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

const testAccGameSessionQueuePrefix = "tfAccQueue-"

func TestAccGameLiftGameSessionQueue_basic(t *testing.T) {
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
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameSessionQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameSessionQueueConfig_basic(queueName,
					playerLatencyPolicies, timeoutInSeconds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`gamesessionqueue/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", queueName),
					resource.TestCheckResourceAttr(resourceName, "destinations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.0.maximum_individual_player_latency_milliseconds",
						fmt.Sprintf("%d", *playerLatencyPolicies[0].MaximumIndividualPlayerLatencyMilliseconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.0.policy_duration_seconds",
						fmt.Sprintf("%d", *playerLatencyPolicies[0].PolicyDurationSeconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.1.maximum_individual_player_latency_milliseconds",
						fmt.Sprintf("%d", *playerLatencyPolicies[1].MaximumIndividualPlayerLatencyMilliseconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.1.policy_duration_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout_in_seconds", fmt.Sprintf("%d", timeoutInSeconds)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccGameSessionQueueConfig_basic(uQueueName, uPlayerLatencyPolicies, uTimeoutInSeconds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(resourceName, &conf),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`gamesessionqueue/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", uQueueName),
					resource.TestCheckResourceAttr(resourceName, "destinations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.0.maximum_individual_player_latency_milliseconds",
						fmt.Sprintf("%d", *uPlayerLatencyPolicies[0].MaximumIndividualPlayerLatencyMilliseconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.0.policy_duration_seconds",
						fmt.Sprintf("%d", *uPlayerLatencyPolicies[0].PolicyDurationSeconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.1.maximum_individual_player_latency_milliseconds",
						fmt.Sprintf("%d", *uPlayerLatencyPolicies[1].MaximumIndividualPlayerLatencyMilliseconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.1.policy_duration_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout_in_seconds", fmt.Sprintf("%d", uTimeoutInSeconds)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var conf gamelift.GameSessionQueue

	resourceName := "aws_gamelift_game_session_queue.test"
	queueName := testAccGameSessionQueuePrefix + sdkacctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameSessionQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameSessionQueueConfig_basicTags1(queueName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGameSessionQueueConfig_basicTags2(queueName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccGameSessionQueueConfig_basicTags1(queueName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccGameLiftGameSessionQueue_disappears(t *testing.T) {
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
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(gamelift.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, gamelift.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckGameSessionQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGameSessionQueueConfig_basic(queueName,
					playerLatencyPolicies, timeoutInSeconds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGameSessionQueueExists(resourceName, &conf),
					testAccCheckGameSessionQueueDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGameSessionQueueExists(n string, res *gamelift.GameSessionQueue) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no GameLift Session Queue Name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn

		name := rs.Primary.Attributes["name"]
		limit := int64(1)
		out, err := conn.DescribeGameSessionQueues(&gamelift.DescribeGameSessionQueuesInput{
			Names: aws.StringSlice([]string{name}),
			Limit: &limit,
		})
		if err != nil {
			return err
		}
		attributes := out.GameSessionQueues
		if len(attributes) < 1 {
			return fmt.Errorf("gmelift Session Queue %q not found", name)
		}
		if len(attributes) != 1 {
			return fmt.Errorf("expected exactly 1 GameLift Session Queue, found %d under %q",
				len(attributes), name)
		}
		queue := attributes[0]

		if *queue.Name != name {
			return fmt.Errorf("gamelift Session Queue not found")
		}

		*res = *queue

		return nil
	}
}

func testAccCheckGameSessionQueueDisappears(res *gamelift.GameSessionQueue) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn

		input := &gamelift.DeleteGameSessionQueueInput{Name: res.Name}

		_, err := conn.DeleteGameSessionQueue(input)

		return err
	}
}

func testAccCheckGameSessionQueueDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GameLiftConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_gamelift_game_session_queue" {
			continue
		}

		input := &gamelift.DescribeGameSessionQueuesInput{
			Names: aws.StringSlice([]string{rs.Primary.ID}),
			Limit: aws.Int64(1),
		}

		// Deletions can take a few seconds
		err := resource.Retry(30*time.Second, func() *resource.RetryError {
			out, err := conn.DescribeGameSessionQueues(input)

			if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
				return nil
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			attributes := out.GameSessionQueues

			if len(attributes) > 0 {
				return resource.RetryableError(fmt.Errorf("gamelift Session Queue still exists"))
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccGameSessionQueueConfig_basic(queueName string,
	playerLatencyPolicies []gamelift.PlayerLatencyPolicy, timeoutInSeconds int64) string {
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
}
`,
		queueName,
		*playerLatencyPolicies[0].MaximumIndividualPlayerLatencyMilliseconds,
		*playerLatencyPolicies[0].PolicyDurationSeconds,
		*playerLatencyPolicies[1].MaximumIndividualPlayerLatencyMilliseconds,
		timeoutInSeconds)
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
