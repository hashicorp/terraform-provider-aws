package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"strings"
)

const testAccGameliftQueuePrefix = "tfAccQueue-"

func init() {
	resource.AddTestSweepers("aws_gamelift_queue", &resource.Sweeper{
		Name: "aws_gamelift_queue",
		F:    testSweepGameliftQueue,
	})
}

func testSweepGameliftQueue(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).gameliftconn

	out, err := conn.DescribeGameSessionQueues(&gamelift.DescribeGameSessionQueuesInput{})

	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Gamelife Queue sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error listing Gamelift Session Queue: %s", err)

	}

	if len(out.GameSessionQueues) == 0 {
		log.Print("[DEBUG] No Gamelift Session Queue to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d Gamelift Session Queue", len(out.GameSessionQueues))

	for _, queue := range out.GameSessionQueues {
		if !strings.HasPrefix(*queue.Name, testAccGameliftQueuePrefix) {
			continue
		}

		log.Printf("[INFO] Deleting Gamelift Session Queue %q", *queue.Name)
		_, err := conn.DeleteGameSessionQueue(&gamelift.DeleteGameSessionQueueInput{
			Name: aws.String(*queue.Name),
		})
		if err != nil {
			return fmt.Errorf("error deleting Gamelift Session Queue (%s): %s",
				*queue.Name, err)
		}
	}

	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Gamelift Session Queue sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error listing Gamelift Session Queue: %s", err)
	}

	return nil

}

func TestAccAWSGameliftQueue_basic(t *testing.T) {
	var conf gamelift.GameSessionQueue

	rString := acctest.RandString(8)
	queueName := getComposedQueueName(rString)
	destinations := gamelift.GameSessionQueueDestination{
		DestinationArn: aws.String(acctest.RandString(8)),
	}
	playerLatencyPolicies := gamelift.PlayerLatencyPolicy{
		MaximumIndividualPlayerLatencyMilliseconds: aws.Int64(20),
		PolicyDurationSeconds:                      aws.Int64(30),
	}
	timeoutInSeconds := int64(124)

	//uQueueName := getComposedQueueName(fmt.Sprintf("else-%s", rString))
	uDestinations := gamelift.GameSessionQueueDestination{
		DestinationArn: aws.String(acctest.RandString(8)),
	}
	uPlayerLatencyPolicies := gamelift.PlayerLatencyPolicy{
		MaximumIndividualPlayerLatencyMilliseconds: aws.Int64(30),
		PolicyDurationSeconds:                      aws.Int64(40),
	}
	uTimeoutInSeconds := int64(600)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftQueueBasicConfig(queueName, destinations,
					playerLatencyPolicies, timeoutInSeconds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftQueueExists("aws_gamelift_queue.test", &conf),
					resource.TestCheckResourceAttr("aws_gamelift_queue.test", "name", queueName),

					resource.TestCheckResourceAttr("aws_gamelift_queue.test",
						"destinations.#",
						"1"),

					resource.TestCheckResourceAttr("aws_gamelift_queue.test",
						"destinations.0",
						fmt.Sprintf("%s", *destinations.DestinationArn)),

					resource.TestCheckResourceAttr("aws_gamelift_queue.test",
						"player_latency_policies.#", "1"),

					resource.TestCheckResourceAttr("aws_gamelift_queue.test",
						"player_latency_policies.0.maximum_individual_player_latency_milliseconds",
						fmt.Sprintf("%d", *playerLatencyPolicies.MaximumIndividualPlayerLatencyMilliseconds)),

					resource.TestCheckResourceAttr("aws_gamelift_queue.test",
						"player_latency_policies.0.policy_duration_seconds",
						fmt.Sprintf("%d", *playerLatencyPolicies.PolicyDurationSeconds)),

					resource.TestCheckResourceAttr("aws_gamelift_queue.test",
						"timeout_in_seconds", fmt.Sprintf("%d", timeoutInSeconds)),
				),
			},
			{
				Config: testAccAWSGameliftQueueBasicConfig(queueName, uDestinations,
					uPlayerLatencyPolicies, uTimeoutInSeconds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftQueueExists("aws_gamelift_queue.test", &conf),
					resource.TestCheckResourceAttr("aws_gamelift_queue.test", "name", queueName),

					resource.TestCheckResourceAttr("aws_gamelift_queue.test",
						"destinations.#", "1"),

					resource.TestCheckResourceAttr("aws_gamelift_queue.test",
						"destinations.0",
						fmt.Sprintf("%s", *uDestinations.DestinationArn)),

					resource.TestCheckResourceAttr("aws_gamelift_queue.test",
						"player_latency_policies.#", "1"),

					resource.TestCheckResourceAttr("aws_gamelift_queue.test",
						"player_latency_policies.0.maximum_individual_player_latency_milliseconds",
						fmt.Sprintf("%d", *uPlayerLatencyPolicies.MaximumIndividualPlayerLatencyMilliseconds)),

					resource.TestCheckResourceAttr("aws_gamelift_queue.test",
						"player_latency_policies.0.policy_duration_seconds",
						fmt.Sprintf("%d", *uPlayerLatencyPolicies.PolicyDurationSeconds)),

					resource.TestCheckResourceAttr("aws_gamelift_queue.test",
						"timeout_in_seconds", fmt.Sprintf("%d", uTimeoutInSeconds)),
				),
			},
		},
	})
}

func testAccCheckAWSGameliftQueueExists(n string, res *gamelift.GameSessionQueue) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Gamelift Session Queue Name is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

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
			return fmt.Errorf("expected exactly 1 Gamelift Session Queue, found %d under %q",
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

func testAccCheckAWSGameliftQueueDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).gameliftconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_gamelift_queue" {
			continue
		}

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

		if len(attributes) > 0 {
			return fmt.Errorf("gamelift Session Queue still exists")
		}

		return nil
	}

	return nil

}

func testAccAWSGameliftQueueBasicConfig(queueName string, destinations gamelift.GameSessionQueueDestination,
	playerLatencyPolicies gamelift.PlayerLatencyPolicy, timeoutInSeconds int64) string {
	return fmt.Sprintf(`
resource "aws_gamelift_queue" "test" {
  name = "%s"
  destinations = ["%s"]
  player_latency_policies {
	maximum_individual_player_latency_milliseconds = %d
	policy_duration_seconds = %d
  }
  timeout_in_seconds = %d
}
`,
		queueName,
		*destinations.DestinationArn,
		*playerLatencyPolicies.MaximumIndividualPlayerLatencyMilliseconds,
		*playerLatencyPolicies.PolicyDurationSeconds,
		timeoutInSeconds)
}

func getComposedQueueName(name string) string {
	return fmt.Sprintf("%s%s", testAccGameliftQueuePrefix, name)
}
