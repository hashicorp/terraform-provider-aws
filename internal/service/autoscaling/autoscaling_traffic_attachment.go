package autoscaling

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_autoscaling_traffic_attachment" name="Autoscaling Traffic Attachment")
func ResourceAutoscalingTrafficAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrafficAttachmentCreate,
		ReadWithoutTimeout:   resourceTrafficAttachmentRead,
		DeleteWithoutTimeout: resourceTrafficAttachmentDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"autoscaling_group_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"traffic_sources": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
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

const (
	ResNameAutoscalingTrafficAttachment = "Autoscaling Traffic Attachment"
)

func resourceTrafficAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AutoScalingConn()
	trafficSources := expandTrafficSourceIdentifier(d.Get("traffic_sources").([]interface{})[0].(map[string]interface{}))
	trafficSourcesIdentifier := aws.StringValue(trafficSources.Identifier)
	trafficSourcesType := aws.StringValue(trafficSources.Type)
	asgName := d.Get("autoscaling_group_name").(string)
	id := strings.Join([]string{asgName, aws.StringValue(trafficSources.Type), aws.StringValue(trafficSources.Identifier)}, "/")
	in := &autoscaling.AttachTrafficSourcesInput{
		AutoScalingGroupName: aws.String(asgName),
		TrafficSources:       []*autoscaling.TrafficSourceIdentifier{trafficSources},
	}

	_, err := conn.AttachTrafficSourcesWithContext(ctx, in)

	if err != nil {
		return diag.Errorf("creating Autoscaling Attachment (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitTrafficAttachmentCreated(ctx, conn, asgName, trafficSourcesType, trafficSourcesIdentifier, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Autescaling Attachment (%s) create: %s", id, err)
	}

	return resourceTrafficAttachmentRead(ctx, d, meta)
}

func resourceTrafficAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn()

	asgName := d.Get("autoscaling_group_name").(string)
	trafficSources := expandTrafficSourceIdentifier(d.Get("traffic_sources").([]interface{})[0].(map[string]interface{}))
	sourceIdentifier := aws.StringValue(trafficSources.Identifier)
	sourceType := aws.StringValue(trafficSources.Type)
	_, err := FindTrafficAttachment(ctx, conn, asgName, sourceType, sourceIdentifier)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Source Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return diag.Errorf("reading Source Attachment (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceTrafficAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	conn := meta.(*conns.AWSClient).AutoScalingConn()
	asgName := d.Get("autoscaling_group_name").(string)
	sourceIdentifier := expandTrafficSourceIdentifier(d.Get("traffic_sources").([]interface{})[0].(map[string]interface{}))

	log.Printf("[INFO] Deleting AutoScaling AutoscalingTrafficAttachment %s", d.Id())

	_, err := conn.DetachTrafficSourcesWithContext(ctx, &autoscaling.DetachTrafficSourcesInput{
		AutoScalingGroupName: aws.String(asgName),
		TrafficSources:       []*autoscaling.TrafficSourceIdentifier{sourceIdentifier},
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Source Attachment (%s): %s", d.Id(), err)
	}

	if _, err := waitTrafficAttachmentDeleted(ctx, conn, asgName, *sourceIdentifier.Type, *sourceIdentifier.Identifier, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting Source Attachment (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func waitTrafficAttachmentCreated(ctx context.Context, conn *autoscaling.AutoScaling, asgName string, sourceType string, sourceIdentifier string, timeout time.Duration) (*autoscaling.TrafficSourceState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{"Adding"},
		Target:                    []string{"InService", "Added"},
		Refresh:                   statusTrafficAttachment(ctx, conn, asgName, sourceType, sourceIdentifier),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*autoscaling.TrafficSourceState); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.State)))
		return output, err
	}

	return nil, err
}

func waitTrafficAttachmentDeleted(ctx context.Context, conn *autoscaling.AutoScaling, asgName string, sourceType string, sourceIdentifier string, timeout time.Duration) (*autoscaling.TrafficSourceState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{"Removing", "Removed"},
		Target:                    []string{},
		Refresh:                   statusTrafficAttachment(ctx, conn, asgName, sourceType, sourceIdentifier),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*autoscaling.TrafficSourceState); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.State)))
		return output, err
	}

	return nil, err
}

func statusTrafficAttachment(ctx context.Context, conn *autoscaling.AutoScaling, asgName string, sourceType string, sourceIdentifier string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTrafficAttachment(ctx, conn, asgName, sourceType, sourceIdentifier)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(*output.TrafficSources[0].State), nil
	}
}

func FindTrafficAttachment(ctx context.Context, conn *autoscaling.AutoScaling, asgName string, sourceType string, sourceIdentifier string) (*autoscaling.DescribeTrafficSourcesOutput, error) {

	// Fetch the ASG first
	_, err := FindGroupByName(ctx, conn, asgName)
	if err != nil {
		return nil, err
	}

	input := &autoscaling.DescribeTrafficSourcesInput{
		AutoScalingGroupName: aws.String(asgName),
		TrafficSourceType:    aws.String(sourceType),
	}

	var result *autoscaling.DescribeTrafficSourcesOutput

	err = conn.DescribeTrafficSourcesPagesWithContext(ctx, input,
		func(page *autoscaling.DescribeTrafficSourcesOutput, lastPage bool) bool {
			for _, v := range page.TrafficSources {
				if aws.StringValue(v.Identifier) == sourceIdentifier {
					result = page
					return false
				}
			}

			return true
		},
	)

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		return nil, err
	}

	if result == nil {
		return nil, &retry.NotFoundError{
			LastError: fmt.Errorf("Auto Scaling Group (%s) attachment not found", asgName),
		}
	}

	return result, nil
}

func expandTrafficSourceIdentifier(tfMap map[string]interface{}) *autoscaling.TrafficSourceIdentifier {
	apiObject := &autoscaling.TrafficSourceIdentifier{}

	if v, ok := tfMap["identifier"].(string); ok && v != "" {
		apiObject.Identifier = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}
