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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_gamelift_game_session_queue", name="Game Session Queue")
// @Tags(identifierAttribute="arn")
func resourceGameSessionQueue() *schema.Resource {
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
	}
}

func resourceGameSessionQueueCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &gamelift.CreateGameSessionQueueInput{
		Name:                  aws.String(name),
		Destinations:          expandGameSessionQueueDestinations(d.Get("destinations").([]any)),
		PlayerLatencyPolicies: expandGameSessionPlayerLatencyPolicies(d.Get("player_latency_policy").([]any)),
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

func resourceGameSessionQueueRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	sessionQueue, err := findGameSessionQueueByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] GameLift Game Session Queue %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading GameLift Game Session Queue (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, sessionQueue.GameSessionQueueArn)
	d.Set("custom_event_data", sessionQueue.CustomEventData)
	if err := d.Set("destinations", flattenGameSessionQueueDestinations(sessionQueue.Destinations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting destinations: %s", err)
	}
	d.Set(names.AttrName, sessionQueue.Name)
	d.Set("notification_target", sessionQueue.NotificationTarget)
	if err := d.Set("player_latency_policy", flattenPlayerLatencyPolicies(sessionQueue.PlayerLatencyPolicies)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting player_latency_policy: %s", err)
	}
	d.Set("timeout_in_seconds", sessionQueue.TimeoutInSeconds)

	return diags
}

func resourceGameSessionQueueUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &gamelift.UpdateGameSessionQueueInput{
			Destinations:          expandGameSessionQueueDestinations(d.Get("destinations").([]any)),
			Name:                  aws.String(d.Id()),
			PlayerLatencyPolicies: expandGameSessionPlayerLatencyPolicies(d.Get("player_latency_policy").([]any)),
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

func resourceGameSessionQueueDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GameLiftClient(ctx)

	log.Printf("[INFO] Deleting GameLift Session Queue: %s", d.Id())
	_, err := conn.DeleteGameSessionQueue(ctx, &gamelift.DeleteGameSessionQueueInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "does not exist") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GameLift Game Session Queue (%s): %s", d.Id(), err)
	}

	// Deletions can take a few seconds.
	const (
		timeout = 30 * time.Second
	)
	_, err = tfresource.RetryUntilNotFound(ctx, timeout, func() (any, error) {
		return findGameSessionQueueByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for GameLift Game Session Queue (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findGameSessionQueueByName(ctx context.Context, conn *gamelift.Client, name string) (*awstypes.GameSessionQueue, error) {
	input := &gamelift.DescribeGameSessionQueuesInput{
		Names: []string{name},
	}

	return findGameSessionQueue(ctx, conn, input)
}

func findGameSessionQueue(ctx context.Context, conn *gamelift.Client, input *gamelift.DescribeGameSessionQueuesInput) (*awstypes.GameSessionQueue, error) {
	output, err := findGameSessionQueues(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findGameSessionQueues(ctx context.Context, conn *gamelift.Client, input *gamelift.DescribeGameSessionQueuesInput) ([]awstypes.GameSessionQueue, error) {
	var output []awstypes.GameSessionQueue

	pages := gamelift.NewDescribeGameSessionQueuesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.GameSessionQueues...)
	}

	return output, nil
}

func flattenGameSessionQueueDestinations(apiObjects []awstypes.GameSessionQueueDestination) []any {
	return tfslices.ApplyToAll(apiObjects, func(v awstypes.GameSessionQueueDestination) any {
		return aws.ToString(v.DestinationArn)
	})
}

func flattenPlayerLatencyPolicies(apiObjects []awstypes.PlayerLatencyPolicy) []any {
	tfList := make([]any, 0)

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"maximum_individual_player_latency_milliseconds": aws.ToInt32(apiObject.MaximumIndividualPlayerLatencyMilliseconds),
			"policy_duration_seconds":                        aws.ToInt32(apiObject.PolicyDurationSeconds),
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandGameSessionQueueDestinations(tfList []any) []awstypes.GameSessionQueueDestination {
	if len(tfList) < 1 {
		return nil
	}

	return tfslices.ApplyToAll(tfList, func(v any) awstypes.GameSessionQueueDestination {
		return awstypes.GameSessionQueueDestination{
			DestinationArn: aws.String(v.(string)),
		}
	})
}

func expandGameSessionPlayerLatencyPolicies(tfList []any) []awstypes.PlayerLatencyPolicy {
	if len(tfList) < 1 {
		return nil
	}

	var apiObjects []awstypes.PlayerLatencyPolicy

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)
		apiObjects = append(apiObjects, awstypes.PlayerLatencyPolicy{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int32(int32(tfMap["maximum_individual_player_latency_milliseconds"].(int))),
			PolicyDurationSeconds:                      aws.Int32(int32(tfMap["policy_duration_seconds"].(int))),
		})
	}

	return apiObjects
}
