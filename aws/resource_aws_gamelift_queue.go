package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"log"
)

func resourceAwsGameliftQueue() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGameliftQueuesCreate,
		Read:   resourceAwsGameliftQueuesRead,
		Update: resourceAwsGameliftQueuesUpdate,
		Delete: resourceAwsGameliftQueuesDelete,

		Schema: map[string]*schema.Schema{
			"destinations": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"player_latency_policies": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maximum_individual_player_latency_milliseconds": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"policy_duration_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
			},
			"timeout_in_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsGameliftQueuesCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	input := getFullInputCreate(d)
	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}
	log.Printf("[INFO] Creating Gamelift Session Queue: %s", input)
	out, err := conn.CreateGameSessionQueue(&input)
	if err != nil {
		return err
	}

	d.SetId(*out.GameSessionQueue.GameSessionQueueArn)
	d.Set("name", out.GameSessionQueue.Name)

	return resourceAwsGameliftQueuesRead(d, meta)
}

func resourceAwsGameliftQueuesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn
	log.Printf("[INFO] Describing Gamelift Session Queues: %s", d.Get("name"))
	limit := int64(1)
	out, err := conn.DescribeGameSessionQueues(&gamelift.DescribeGameSessionQueuesInput{
		Names: aws.StringSlice([]string{d.Get("name").(string)}),
		Limit: &limit,
	})
	if err != nil {
		if isAWSErr(err, gamelift.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] Gamelift Session Queues (%s) not found, removing from state", d.Get("name"))
			return nil
		}
		return err
	}
	sessionQueues := out.GameSessionQueues

	if len(sessionQueues) < 1 {
		log.Printf("[WARN] Gamelift Session Queue (%s) not found, removing from state", d.Get("name"))
		return nil
	}
	if len(sessionQueues) != 1 {
		return fmt.Errorf("expected exactly 1 Gamelift Session Queues, found %d under %q",
			len(sessionQueues), d.Get("name"))
	}
	sessionQueue := sessionQueues[0]

	d.Set("destinations", sessionQueue.Destinations)
	d.Set("name", sessionQueue.Name)
	d.Set("player_latency_policies", sessionQueue.PlayerLatencyPolicies)
	d.Set("timeout_in_seconds", sessionQueue.TimeoutInSeconds)

	return nil
}

func resourceAwsGameliftQueuesUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	name := d.Get("name").(string)

	log.Printf("[INFO] Updating Gamelift Session Queue: %s", name)

	if d.HasChange("name") || d.HasChange("destinations") ||
		d.HasChange("player_latency_policies") || d.HasChange("timeout_in_seconds") {

		input := getFullInputUpdate(d)

		_, err := conn.UpdateGameSessionQueue(&input)
		if err != nil {
			return err
		}
	}

	return resourceAwsGameliftQueuesRead(d, meta)
}

func resourceAwsGameliftQueuesDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn
	name := d.Get("name").(string)
	log.Printf("[INFO] Deleting Gamelift Session Queue: %s", name)
	_, err := conn.DeleteGameSessionQueue(&gamelift.DeleteGameSessionQueueInput{
		Name: aws.String(name),
	})
	if err != nil {
		return err
	}

	d.SetId("")
	return nil

}

func getFullInputCreate(d *schema.ResourceData) gamelift.CreateGameSessionQueueInput {
	return gamelift.CreateGameSessionQueueInput{
		Name:                  aws.String(d.Get("name").(string)),
		Destinations:          getDestinations(d.Get("destinations").([]interface{})),
		PlayerLatencyPolicies: getPlayerLatencyPolicies(d.Get("player_latency_policies").([]interface{})),
		TimeoutInSeconds:      aws.Int64(int64(d.Get("timeout_in_seconds").(int))),
	}
}

func getFullInputUpdate(d *schema.ResourceData) gamelift.UpdateGameSessionQueueInput {
	return gamelift.UpdateGameSessionQueueInput{
		Name:                  aws.String(d.Get("name").(string)),
		Destinations:          getDestinations(d.Get("destinations").([]interface{})),
		PlayerLatencyPolicies: getPlayerLatencyPolicies(d.Get("player_latency_policies").([]interface{})),
		TimeoutInSeconds:      aws.Int64(int64(d.Get("timeout_in_seconds").(int))),
	}
}

func getDestinations(destinationsMap []interface{}) []*gamelift.GameSessionQueueDestination {
	if len(destinationsMap) < 1 {
		return nil
	}
	var destinations []*gamelift.GameSessionQueueDestination
	for _, destination := range destinationsMap {
		destinations = append(
			destinations,
			&gamelift.GameSessionQueueDestination{
				DestinationArn: aws.String(destination.(string)),
			})
	}
	return destinations
}

func getPlayerLatencyPolicies(destinationsPlayerLatencyPolicyMap []interface{}) []*gamelift.PlayerLatencyPolicy {
	if len(destinationsPlayerLatencyPolicyMap) < 1 {
		return nil
	}
	var playerLatencyPolicies []*gamelift.PlayerLatencyPolicy
	for _, playerLatencyPolicy := range destinationsPlayerLatencyPolicyMap {
		item := playerLatencyPolicy.(map[string]interface{})
		playerLatencyPolicies = append(
			playerLatencyPolicies,
			&gamelift.PlayerLatencyPolicy{
				MaximumIndividualPlayerLatencyMilliseconds: aws.Int64(int64(item["maximum_individual_player_latency_milliseconds"].(int))),
				PolicyDurationSeconds:                      aws.Int64(int64(item["policy_duration_seconds"].(int))),
			})
	}
	return playerLatencyPolicies
}
