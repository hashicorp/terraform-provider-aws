package autoscaling

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_autoscaling_traffic_attachment")
func ResourceTrafficAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrafficAttachmentCreate,
		ReadWithoutTimeout:   resourceTrafficAttachmentRead,
		DeleteWithoutTimeout: resourceTrafficAttachmentDelete,

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
						"identifier": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
						"type": {
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

func resourceTrafficAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn()

	asgName := d.Get("autoscaling_group_name").(string)
	trafficSource := expandTrafficSourceIdentifier(d.Get("traffic_source").([]interface{})[0].(map[string]interface{}))
	trafficSourceID := aws.StringValue(trafficSource.Identifier)
	trafficSourceType := aws.StringValue(trafficSource.Type)
	id := trafficAttachmentCreateResourceID(asgName, trafficSourceType, trafficSourceID)
	input := &autoscaling.AttachTrafficSourcesInput{
		AutoScalingGroupName: aws.String(asgName),
		TrafficSources:       []*autoscaling.TrafficSourceIdentifier{trafficSource},
	}

	_, err := conn.AttachTrafficSourcesWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Auto Scaling Traffic Attachment (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitTrafficAttachmentCreated(ctx, conn, asgName, trafficSourceType, trafficSourceID, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Auto Scaling Traffic Attachment (%s) create: %s", id, err)
	}

	return resourceTrafficAttachmentRead(ctx, d, meta)
}

func resourceTrafficAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn()

	asgName, trafficSourceType, trafficSourceID, err := TrafficAttachmentParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = FindTrafficAttachmentByThreePartKey(ctx, conn, asgName, trafficSourceType, trafficSourceID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Auto Scaling Traffic Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Traffic Attachment (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceTrafficAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn()

	asgName, trafficSourceType, trafficSourceID, err := TrafficAttachmentParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting Auto Scaling Traffic Attachment: %s", d.Id())
	_, err = conn.DetachTrafficSourcesWithContext(ctx, &autoscaling.DetachTrafficSourcesInput{
		AutoScalingGroupName: aws.String(asgName),
		TrafficSources:       []*autoscaling.TrafficSourceIdentifier{expandTrafficSourceIdentifier(d.Get("traffic_source").([]interface{})[0].(map[string]interface{}))},
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Auto Scaling Traffic Attachment (%s): %s", d.Id(), err)
	}

	if _, err := waitTrafficAttachmentDeleted(ctx, conn, asgName, trafficSourceType, trafficSourceID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Auto Scaling Traffic Attachment (%s) delete: %s", d.Id(), err)
	}

	return nil
}

const trafficAttachmentIDSeparator = ","

func trafficAttachmentCreateResourceID(asgName, trafficSourceType, trafficSourceID string) string {
	parts := []string{asgName, trafficSourceType, trafficSourceID}
	id := strings.Join(parts, trafficAttachmentIDSeparator)

	return id
}

func TrafficAttachmentParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, trafficAttachmentIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected asg-name%[2]straffic-source-type%[2]straffic-source-id", id, trafficAttachmentIDSeparator)
}

func FindTrafficAttachmentByThreePartKey(ctx context.Context, conn *autoscaling.AutoScaling, asgName, trafficSourceType, trafficSourceID string) (*autoscaling.TrafficSourceState, error) {
	input := &autoscaling.DescribeTrafficSourcesInput{
		AutoScalingGroupName: aws.String(asgName),
		TrafficSourceType:    aws.String(trafficSourceType),
	}

	output, err := findTrafficSourceStates(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	output = slices.Filter(output, func(v *autoscaling.TrafficSourceState) bool {
		return aws.StringValue(v.Identifier) == trafficSourceID
	})

	return tfresource.AssertSinglePtrResult(output)
}

func statusTrafficAttachment(ctx context.Context, conn *autoscaling.AutoScaling, asgName, trafficSourceType, trafficSourceID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTrafficAttachmentByThreePartKey(ctx, conn, asgName, trafficSourceType, trafficSourceID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func waitTrafficAttachmentCreated(ctx context.Context, conn *autoscaling.AutoScaling, asgName, trafficSourceType, trafficSourceID string, timeout time.Duration) (*autoscaling.TrafficSourceState, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{TrafficSourceStateAdding},
		Target:  []string{TrafficSourceStateAdded, TrafficSourceStateInService},
		Refresh: statusTrafficAttachment(ctx, conn, asgName, trafficSourceType, trafficSourceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*autoscaling.TrafficSourceState); ok {
		return output, err
	}

	return nil, err
}

func waitTrafficAttachmentDeleted(ctx context.Context, conn *autoscaling.AutoScaling, asgName, trafficSourceType, trafficSourceID string, timeout time.Duration) (*autoscaling.TrafficSourceState, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{TrafficSourceStateRemoving, TrafficSourceStateRemoved},
		Target:  []string{},
		Refresh: statusTrafficAttachment(ctx, conn, asgName, trafficSourceType, trafficSourceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*autoscaling.TrafficSourceState); ok {
		return output, err
	}

	return nil, err
}
