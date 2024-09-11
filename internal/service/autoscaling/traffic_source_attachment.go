// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_autoscaling_traffic_source_attachment", name="Traffic Source Attachment")
func resourceTrafficSourceAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrafficSourceAttachmentCreate,
		ReadWithoutTimeout:   resourceTrafficSourceAttachmentRead,
		DeleteWithoutTimeout: resourceTrafficSourceAttachmentDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"autoscaling_group_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"traffic_source": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrIdentifier: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
						names.AttrType: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
					},
				},
			},
		},
	}
}

func resourceTrafficSourceAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	asgName := d.Get("autoscaling_group_name").(string)
	trafficSource := expandTrafficSourceIdentifier(d.Get("traffic_source").([]interface{})[0].(map[string]interface{}))
	trafficSourceID := aws.ToString(trafficSource.Identifier)
	trafficSourceType := aws.ToString(trafficSource.Type)
	id := trafficSourceAttachmentCreateResourceID(asgName, trafficSourceType, trafficSourceID)
	input := &autoscaling.AttachTrafficSourcesInput{
		AutoScalingGroupName: aws.String(asgName),
		TrafficSources:       []awstypes.TrafficSourceIdentifier{trafficSource},
	}

	_, err := conn.AttachTrafficSources(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Auto Scaling Traffic Source Attachment (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitTrafficSourceAttachmentCreated(ctx, conn, asgName, trafficSourceType, trafficSourceID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Auto Scaling Traffic Source Attachment (%s) create: %s", id, err)
	}

	return append(diags, resourceTrafficSourceAttachmentRead(ctx, d, meta)...)
}

func resourceTrafficSourceAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	asgName, trafficSourceType, trafficSourceID, err := trafficSourceAttachmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = findTrafficSourceAttachmentByThreePartKey(ctx, conn, asgName, trafficSourceType, trafficSourceID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Auto Scaling Traffic Source Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Traffic Source Attachment (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceTrafficSourceAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	asgName, trafficSourceType, trafficSourceID, err := trafficSourceAttachmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	trafficSource := expandTrafficSourceIdentifier(d.Get("traffic_source").([]interface{})[0].(map[string]interface{}))

	log.Printf("[INFO] Deleting Auto Scaling Traffic Source Attachment: %s", d.Id())
	_, err = conn.DetachTrafficSources(ctx, &autoscaling.DetachTrafficSourcesInput{
		AutoScalingGroupName: aws.String(asgName),
		TrafficSources:       []awstypes.TrafficSourceIdentifier{trafficSource},
	})

	if tfawserr.ErrMessageContains(err, errCodeValidationError, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Auto Scaling Traffic Source Attachment (%s): %s", d.Id(), err)
	}

	if _, err := waitTrafficSourceAttachmentDeleted(ctx, conn, asgName, trafficSourceType, trafficSourceID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Auto Scaling Traffic Source Attachment (%s) delete: %s", d.Id(), err)
	}

	return diags
}

const trafficSourceAttachmentIDSeparator = ","

func trafficSourceAttachmentCreateResourceID(asgName, trafficSourceType, trafficSourceID string) string {
	parts := []string{asgName, trafficSourceType, trafficSourceID}
	id := strings.Join(parts, trafficSourceAttachmentIDSeparator)

	return id
}

func trafficSourceAttachmentParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, trafficSourceAttachmentIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected asg-name%[2]straffic-source-type%[2]straffic-source-id", id, trafficSourceAttachmentIDSeparator)
}

func findTrafficSourceAttachmentByThreePartKey(ctx context.Context, conn *autoscaling.Client, asgName, trafficSourceType, trafficSourceID string) (*awstypes.TrafficSourceState, error) {
	input := &autoscaling.DescribeTrafficSourcesInput{
		AutoScalingGroupName: aws.String(asgName),
		TrafficSourceType:    aws.String(trafficSourceType),
	}

	output, err := findTrafficSourceStates(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	output = slices.Filter(output, func(v awstypes.TrafficSourceState) bool {
		return aws.ToString(v.Identifier) == trafficSourceID
	})

	return tfresource.AssertSingleValueResult(output)
}

func statusTrafficSourceAttachment(ctx context.Context, conn *autoscaling.Client, asgName, trafficSourceType, trafficSourceID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTrafficSourceAttachmentByThreePartKey(ctx, conn, asgName, trafficSourceType, trafficSourceID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.State), nil
	}
}

func waitTrafficSourceAttachmentCreated(ctx context.Context, conn *autoscaling.Client, asgName, trafficSourceType, trafficSourceID string, timeout time.Duration) (*awstypes.TrafficSourceState, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{TrafficSourceStateAdding},
		Target:  []string{TrafficSourceStateAdded, TrafficSourceStateInService},
		Refresh: statusTrafficSourceAttachment(ctx, conn, asgName, trafficSourceType, trafficSourceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TrafficSourceState); ok {
		return output, err
	}

	return nil, err
}

func waitTrafficSourceAttachmentDeleted(ctx context.Context, conn *autoscaling.Client, asgName, trafficSourceType, trafficSourceID string, timeout time.Duration) (*awstypes.TrafficSourceState, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{TrafficSourceStateRemoving, TrafficSourceStateRemoved},
		Target:  []string{},
		Refresh: statusTrafficSourceAttachment(ctx, conn, asgName, trafficSourceType, trafficSourceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TrafficSourceState); ok {
		return output, err
	}

	return nil, err
}
