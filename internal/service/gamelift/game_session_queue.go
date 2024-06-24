// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
	conn := meta.(*conns.AWSClient).GameLiftConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &gamelift.CreateGameSessionQueueInput{
		Name:                  aws.String(name),
		Destinations:          expandGameSessionQueueDestinations(d.Get("destinations").([]interface{})),
		PlayerLatencyPolicies: expandGameSessionPlayerLatencyPolicies(d.Get("player_latency_policy").([]interface{})),
		TimeoutInSeconds:      aws.Int64(int64(d.Get("timeout_in_seconds").(int))),
		Tags:                  getTagsIn(ctx),
	}

	if v, ok := d.GetOk("custom_event_data"); ok {
		input.CustomEventData = aws.String(v.(string))
	}

	if v, ok := d.GetOk("notification_target"); ok {
		input.NotificationTarget = aws.String(v.(string))
	}

	output, err := conn.CreateGameSessionQueueWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GameLift Game Session Queue (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.GameSessionQueue.Name))

	return append(diags, resourceGameSessionQueueRead(ctx, d, meta)...)
}

func resourceGameSessionQueueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn(ctx)

	sessionQueue, err := FindGameSessionQueueByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GameLift Game Session Queue %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Game Session Queue (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(sessionQueue.GameSessionQueueArn)
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
	conn := meta.(*conns.AWSClient).GameLiftConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &gamelift.UpdateGameSessionQueueInput{
			Name:                  aws.String(d.Id()),
			Destinations:          expandGameSessionQueueDestinations(d.Get("destinations").([]interface{})),
			PlayerLatencyPolicies: expandGameSessionPlayerLatencyPolicies(d.Get("player_latency_policy").([]interface{})),
			TimeoutInSeconds:      aws.Int64(int64(d.Get("timeout_in_seconds").(int))),
		}

		if v, ok := d.GetOk("custom_event_data"); ok {
			input.CustomEventData = aws.String(v.(string))
		}

		if v, ok := d.GetOk("notification_target"); ok {
			input.NotificationTarget = aws.String(v.(string))
		}

		_, err := conn.UpdateGameSessionQueueWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GameLift Game Session Queue (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceGameSessionQueueRead(ctx, d, meta)...)
}

func resourceGameSessionQueueDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftConn(ctx)

	log.Printf("[INFO] Deleting GameLift Session Queue: %s", d.Id())
	_, err := conn.DeleteGameSessionQueueWithContext(ctx, &gamelift.DeleteGameSessionQueueInput{
		Name: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
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

func FindGameSessionQueueByName(ctx context.Context, conn *gamelift.GameLift, name string) (*gamelift.GameSessionQueue, error) {
	input := &gamelift.DescribeGameSessionQueuesInput{
		Names: aws.StringSlice([]string{name}),
	}

	output, err := conn.DescribeGameSessionQueuesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.GameSessionQueues) == 0 || output.GameSessionQueues[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.GameSessionQueues); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.GameSessionQueues[0], nil
}

func flattenGameSessionQueueDestinations(destinations []*gamelift.GameSessionQueueDestination) []interface{} {
	l := make([]interface{}, 0)

	for _, destination := range destinations {
		if destination == nil {
			continue
		}
		l = append(l, aws.StringValue(destination.DestinationArn))
	}

	return l
}

func flattenPlayerLatencyPolicies(playerLatencyPolicies []*gamelift.PlayerLatencyPolicy) []interface{} {
	l := make([]interface{}, 0)
	for _, policy := range playerLatencyPolicies {
		m := map[string]interface{}{
			"maximum_individual_player_latency_milliseconds": aws.Int64Value(policy.MaximumIndividualPlayerLatencyMilliseconds),
			"policy_duration_seconds":                        aws.Int64Value(policy.PolicyDurationSeconds),
		}
		l = append(l, m)
	}
	return l
}

func expandGameSessionQueueDestinations(destinationsMap []interface{}) []*gamelift.GameSessionQueueDestination {
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

func expandGameSessionPlayerLatencyPolicies(destinationsPlayerLatencyPolicyMap []interface{}) []*gamelift.PlayerLatencyPolicy {
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
