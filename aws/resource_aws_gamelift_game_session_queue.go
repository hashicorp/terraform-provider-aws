package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"log"
)

func resourceAwsGameliftGameSessionQueue() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGameliftGameSessionQueueCreate,
		Read:   resourceAwsGameliftGameSessionQueueRead,
		Update: resourceAwsGameliftGameSessionQueueUpdate,
		Delete: resourceAwsGameliftGameSessionQueueDelete,

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
			"player_latency_policy": {
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
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(10, 600),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsGameliftGameSessionQueueCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	input := gamelift.CreateGameSessionQueueInput{
		Name:                  aws.String(d.Get("name").(string)),
		Destinations:          expandGameliftGameSessionQueueDestinations(d.Get("destinations").([]interface{})),
		PlayerLatencyPolicies: expandGameliftGameSessionPlayerLatencyPolicies(d.Get("player_latency_policy").([]interface{})),
		TimeoutInSeconds:      aws.Int64(int64(d.Get("timeout_in_seconds").(int))),
	}
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

	return resourceAwsGameliftGameSessionQueueRead(d, meta)
}

func resourceAwsGameliftGameSessionQueueRead(d *schema.ResourceData, meta interface{}) error {
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
			d.SetId("")
			return nil
		}
		return err
	}
	sessionQueues := out.GameSessionQueues

	if len(sessionQueues) < 1 {
		log.Printf("[WARN] Gamelift Session Queue (%s) not found, removing from state", d.Get("name"))
		d.SetId("")
		return nil
	}
	if len(sessionQueues) != 1 {
		return fmt.Errorf("expected exactly 1 Gamelift Session Queues, found %d under %q",
			len(sessionQueues), d.Get("name"))
	}
	sessionQueue := sessionQueues[0]

	d.Set("arn", sessionQueue.GameSessionQueueArn)
	d.Set("name", sessionQueue.Name)
	d.Set("timeout_in_seconds", sessionQueue.TimeoutInSeconds)
	d.Set("destinations", sessionQueue.Destinations)
	if err := d.Set("player_latency_policy", flattenGameliftPlayerLatencyPolicies(sessionQueue.PlayerLatencyPolicies)); err != nil {
		return fmt.Errorf("error setting player_latency_policy: %s", err)
	}

	return nil
}

func flattenGameliftPlayerLatencyPolicies(playerLatencyPolicies []*gamelift.PlayerLatencyPolicy) []interface{} {
	lst := []interface{}{}
	for _, policy := range playerLatencyPolicies {
		m := map[string]interface{}{
			"maximum_individual_player_latency_milliseconds": aws.Int64Value(policy.MaximumIndividualPlayerLatencyMilliseconds),
			"policy_duration_seconds":                        aws.Int64Value(policy.PolicyDurationSeconds),
		}
		lst = append(lst, m)
	}
	return lst
}

func resourceAwsGameliftGameSessionQueueUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	name := d.Get("name").(string)

	log.Printf("[INFO] Updating Gamelift Session Queue: %s", name)

	input := gamelift.UpdateGameSessionQueueInput{
		Name:                  aws.String(d.Get("name").(string)),
		Destinations:          expandGameliftGameSessionQueueDestinations(d.Get("destinations").([]interface{})),
		PlayerLatencyPolicies: expandGameliftGameSessionPlayerLatencyPolicies(d.Get("player_latency_policy").([]interface{})),
		TimeoutInSeconds:      aws.Int64(int64(d.Get("timeout_in_seconds").(int))),
	}

	_, err := conn.UpdateGameSessionQueue(&input)
	if err != nil {
		return err
	}

	return resourceAwsGameliftGameSessionQueueRead(d, meta)
}

func resourceAwsGameliftGameSessionQueueDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn
	name := d.Get("name").(string)
	log.Printf("[INFO] Deleting Gamelift Session Queue: %s", name)
	_, err := conn.DeleteGameSessionQueue(&gamelift.DeleteGameSessionQueueInput{
		Name: aws.String(name),
	})
	if isAWSErr(err, gamelift.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting Gamelift Game Session Queue (%s): %s", d.Id(), err)
	}

	return nil
}

func expandGameliftGameSessionQueueDestinations(destinationsMap []interface{}) []*gamelift.GameSessionQueueDestination {
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

func expandGameliftGameSessionPlayerLatencyPolicies(destinationsPlayerLatencyPolicyMap []interface{}) []*gamelift.PlayerLatencyPolicy {
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
