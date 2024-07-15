// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_gamelift_game_session_queue", name="Game Session Queue")
// @Tags(identifierAttribute="arn")
func ResourceGameSessionQueue() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGameSessionQueueCreate,
		ReadWithoutTimeout:   resourceGameSessionQueueRead,
		UpdateWithoutTimeout: resourceGameSessionQueueUpdate,
		DeleteWithoutTimeout: resourceGameSessionQueueDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_event_data": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"destinations": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"notification_target": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"timeout_in_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(10, 600),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceGameSessionQueueCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &gamelift.CreateGameSessionQueueInput{
		Name:                  aws.String(name),
		Destinations:          expandGameSessionQueueDestinations(d.Get("destinations").([]interface{})),
		PlayerLatencyPolicies: expandGameSessionPlayerLatencyPolicies(d.Get("player_latency_policy").([]interface{})),
		TimeoutInSeconds:      aws.Int32(int32(d.Get("timeout_in_seconds").(int))),
		Tags:                  getTagsIn(ctx),
	}

	if v, ok := d.GetOk("custom_event_data"); ok {
		input.CustomEventData = aws.String(v.(string))
	}

	if v, ok := d.GetOk("notification_target"); ok {
		input.NotificationTarget = aws.String(v.(string))
	}

	output, err := conn.CreateGameSessionQueue(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift Game Session Queue (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.GameSessionQueue.Name))

	return append(diags, resourceGameSessionQueueRead(ctx, d, meta)...)
}

func resourceGameSessionQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	sessionQueue, err := FindGameSessionQueueByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GameLift Game Session Queue %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Game Session Queue (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(sessionQueue.GameSessionQueueArn)
	d.Set(names.AttrARN, arn)
	d.Set("custom_event_data", sessionQueue.CustomEventData)
	d.Set(names.AttrName, sessionQueue.Name)
	d.Set("notification_target", sessionQueue.NotificationTarget)
	d.Set("timeout_in_seconds", sessionQueue.TimeoutInSeconds)
	if err := d.Set("destinations", flattenGameSessionQueueDestinations(sessionQueue.Destinations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting destinations: %s", err)
	}
	if err := d.Set("player_latency_policy", flattenPlayerLatencyPolicies(sessionQueue.PlayerLatencyPolicies)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting player_latency_policy: %s", err)
	}

	return diags
}

func resourceGameSessionQueueUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &gamelift.UpdateGameSessionQueueInput{
			Name:                  aws.String(d.Id()),
			Destinations:          expandGameSessionQueueDestinations(d.Get("destinations").([]interface{})),
			PlayerLatencyPolicies: expandGameSessionPlayerLatencyPolicies(d.Get("player_latency_policy").([]interface{})),
			TimeoutInSeconds:      aws.Int32(int32(d.Get("timeout_in_seconds").(int))),
		}

		if v, ok := d.GetOk("custom_event_data"); ok {
			input.CustomEventData = aws.String(v.(string))
		}

		if v, ok := d.GetOk("notification_target"); ok {
			input.NotificationTarget = aws.String(v.(string))
		}

		_, err := conn.UpdateGameSessionQueue(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GameLift Game Session Queue (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceGameSessionQueueRead(ctx, d, meta)...)
}

func resourceGameSessionQueueDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	log.Printf("[INFO] Deleting GameLift Session Queue: %s", d.Id())
	_, err := conn.DeleteGameSessionQueue(ctx, &gamelift.DeleteGameSessionQueueInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GameLift Game Session Queue (%s): %s", d.Id(), err)
	}

	// Deletions can take a few seconds.
	_, err = tfresource.RetryUntilNotFound(ctx, 30*time.Second, func() (interface{}, error) {
		return FindGameSessionQueueByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GameLift Game Session Queue (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindGameSessionQueueByName(ctx context.Context, conn *gamelift.Client, name string) (*awstypes.GameSessionQueue, error) {
	input := &gamelift.DescribeGameSessionQueuesInput{
		Names: []string{name},
	}

	output, err := conn.DescribeGameSessionQueues(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output.GameSessionQueues)
}

func flattenGameSessionQueueDestinations(destinations []awstypes.GameSessionQueueDestination) []interface{} {
	l := make([]interface{}, 0)

	for _, destination := range destinations {
		l = append(l, aws.ToString(destination.DestinationArn))
	}

	return l
}

func flattenPlayerLatencyPolicies(playerLatencyPolicies []awstypes.PlayerLatencyPolicy) []interface{} {
	l := make([]interface{}, 0)
	for _, policy := range playerLatencyPolicies {
		m := map[string]interface{}{
			"maximum_individual_player_latency_milliseconds": aws.ToInt32(policy.MaximumIndividualPlayerLatencyMilliseconds),
			"policy_duration_seconds":                        aws.ToInt32(policy.PolicyDurationSeconds),
		}
		l = append(l, m)
	}
	return l
}

func expandGameSessionQueueDestinations(destinationsMap []interface{}) []awstypes.GameSessionQueueDestination {
	if len(destinationsMap) < 1 {
		return nil
	}
	var destinations []awstypes.GameSessionQueueDestination
	for _, destination := range destinationsMap {
		destinations = append(
			destinations,
			awstypes.GameSessionQueueDestination{
				DestinationArn: aws.String(destination.(string)),
			})
	}
	return destinations
}

func expandGameSessionPlayerLatencyPolicies(destinationsPlayerLatencyPolicyMap []interface{}) []awstypes.PlayerLatencyPolicy {
	if len(destinationsPlayerLatencyPolicyMap) < 1 {
		return nil
	}
	var playerLatencyPolicies []awstypes.PlayerLatencyPolicy
	for _, playerLatencyPolicy := range destinationsPlayerLatencyPolicyMap {
		item := playerLatencyPolicy.(map[string]interface{})
		playerLatencyPolicies = append(
			playerLatencyPolicies,
			awstypes.PlayerLatencyPolicy{
				MaximumIndividualPlayerLatencyMilliseconds: aws.Int32(int32(item["maximum_individual_player_latency_milliseconds"].(int))),
				PolicyDurationSeconds:                      aws.Int32(int32(item["policy_duration_seconds"].(int))),
			})
	}
	return playerLatencyPolicies
}
