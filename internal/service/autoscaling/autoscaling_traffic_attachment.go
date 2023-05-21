package autoscaling

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_autoscaling_traffic_attachment")
func ResourceAutoscalingTrafficAttachment() *schema.Resource {
	return &schema.Resource{

		CreateWithoutTimeout: resourceTrafficAttachmentCreate,
		ReadWithoutTimeout:   resourceTrafficAttachmentRead,
		DeleteWithoutTimeout: resourceTrafficAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
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
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn()
	sourceIdentifier := expandTrafficSourceIdentifier(d.Get("traffic_sources").([]interface{})[0].(map[string]interface{}))
	asgName := d.Get("autoscaling_group_name").(string)
	in := &autoscaling.AttachTrafficSourcesInput{
		AutoScalingGroupName: aws.String(asgName),
		TrafficSources:       []*autoscaling.TrafficSourceIdentifier{sourceIdentifier},
	}

	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, d.Timeout(schema.TimeoutCreate),
		func() (interface{}, error) {
			return conn.AttachTrafficSourcesWithContext(ctx, in)
		},
		// ValidationError: Trying to update too many Load Balancers/Target Groups at once. The limit is 10
		ErrCodeValidationError, "update too many")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "attaching Auto Scaling Group (%s) with (%s): %s", asgName, *sourceIdentifier.Identifier, err)
	}

	d.SetId(id.PrefixedUniqueId(fmt.Sprintf("%s-", asgName)))

	return resourceTrafficAttachmentRead(ctx, d, meta)
}

func resourceTrafficAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AutoScalingConn()
	asgName := d.Get("autoscaling_group_name").(string)

	var err error
	trafficSources := expandTrafficSourceIdentifier(d.Get("traffic_sources").([]interface{})[0].(map[string]interface{}))
	sourceIdentifier := aws.StringValue(trafficSources.Identifier)
	sourceType := aws.StringValue(trafficSources.Type)
	out, err := FindTrafficAttachment(ctx, conn, asgName, sourceType, sourceIdentifier)
	fmt.Println(out)
	// if sourceIdentifier.Type == aws.String("elb") {
	// 	err = FindAttachmentByLoadBalancerName(ctx, conn, asgName, *sourceIdentifier.Identifier)
	// } else {
	// 	err = FindTrafficAttachmentByTargetGroupARN(ctx, conn, asgName, *sourceIdentifier.Identifier)
	// }

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Lattice Target Group Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading VPC Lattice Target Group Attachment (%s): %s", d.Id(), err)
	}

	// if err := d.Set("traffic_sources", []interface{}{flattenTargetSummary(output)}); err != nil {
	// 	return diag.Errorf("setting target: %s", err)
	// }

	return nil
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

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.AutoScaling, create.ErrActionDeleting, ResNameAutoscalingTrafficAttachment, d.Id(), err)
	}

	return nil
}

// func FindTrafficAttachmentByLoadBalancerName(ctx context.Context, conn *autoscaling.AutoScaling, asgName, loadBalancerName string) error {
// 	asg, err := FindGroupByName(ctx, conn, asgName)
// 	fmt.Println(asg)
// 	if err != nil {
// 		return err
// 	}

// 	for _, v := range asg.LoadBalancerNames {
// 		fmt.Println(aws.StringValue(v))
// 		if aws.StringValue(v) == loadBalancerName {
// 			fmt.Println("Found")
// 			return nil
// 		}
// 	}

// 	return &retry.NotFoundError{
// 		LastError: fmt.Errorf("Auto Scaling Group (%s) load balancer (%s) attachment not found", asgName, loadBalancerName),
// 	}
// }

// func FindTrafficAttachmentByTargetGroupARN(ctx context.Context, conn *autoscaling.AutoScaling, asgName, targetGroupARN string) error {
// 	asg, err := FindGroupByName(ctx, conn, asgName)
// 	fmt.Println(asg)
// 	if err != nil {
// 		return err
// 	}

// 	for _, v := range asg.TargetGroupARNs {
// 		fmt.Println(aws.StringValue(v))
// 		if aws.StringValue(v) == targetGroupARN {
// 			return nil
// 		}
// 	}

// 	return &retry.NotFoundError{
// 		LastError: fmt.Errorf("Auto Scaling Group (%s) target group (%s) attachment not found", asgName, targetGroupARN),
// 	}
// }

// func flattenTrafficSourceIdentifier(apiObject *autoscaling.) *autoscaling.TrafficSourceIdentifier {
// 	apiObject := &autoscaling.TrafficSourceIdentifier{}

// 	if v, ok := tfMap["identifier"].(string); ok && v != "" {
// 		apiObject.Identifier = aws.String(v)
// 	}

// 	if v, ok := tfMap["type"].(string); ok && v != "" {
// 		apiObject.Type = aws.String(v)
// 	}

// 	return apiObject
// }

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

// func findTrafficAttachment(ctx context.Context, conn *autoscaling.AutoScaling, asgName string, sourceType string, sourceIdentifier string) (*autoscaling.DescribeTrafficSourcesOutput, error) {
// 	input := &autoscaling.DescribeTrafficSourcesInput{
// 		AutoScalingGroupName: aws.String(asgName),
// 		TrafficSourceType:    aws.String(sourceType),
// 	}

// 	paginator, err := conn.DescribeTrafficSources(input)
// 	for paginator {
// 		output, err := paginator.NextPage(ctx)

// 		if errs.IsA[*types.ResourceNotFoundException](err) {
// 			return nil, &retry.NotFoundError{
// 				LastError:   err,
// 				LastRequest: input,
// 			}
// 		}

// 		if err != nil {
// 			return nil, err
// 		}

// 		if output != nil && len(output.Items) == 1 {
// 			return &(output.Items[0]), nil
// 		}
// 	}

// 	return nil, &retry.NotFoundError{}
// }

func FindTrafficAttachment(ctx context.Context, conn *autoscaling.AutoScaling, asgName string, sourceType string, sourceIdentifier string) (*autoscaling.DescribeTrafficSourcesOutput, error) {
	input := &autoscaling.DescribeTrafficSourcesInput{
		AutoScalingGroupName: aws.String(asgName),
		TrafficSourceType:    aws.String(sourceType),
	}

	out, err := conn.DescribeTrafficSourcesWithContext(ctx, input)
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

	for _, v := range out.TrafficSources {
		if aws.StringValue(v.Identifier) == sourceIdentifier {
			fmt.Println(out)
			return out, nil
		}
	}

	return nil, &retry.NotFoundError{}
}
