package elbv2

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTargetGroupAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLbAttachmentCreate,
		Read:   resourceAwsLbAttachmentRead,
		Delete: resourceAwsLbAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"target_group_arn": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"target_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"port": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
			},

			"availability_zone": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
		},
	}
}

func resourceAwsLbAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBV2Conn

	target := &elbv2.TargetDescription{
		Id: aws.String(d.Get("target_id").(string)),
	}

	if v, ok := d.GetOk("port"); ok {
		target.Port = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		target.AvailabilityZone = aws.String(v.(string))
	}

	params := &elbv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(d.Get("target_group_arn").(string)),
		Targets:        []*elbv2.TargetDescription{target},
	}

	log.Printf("[INFO] Registering Target %s with Target Group %s", d.Get("target_id").(string),
		d.Get("target_group_arn").(string))

	err := resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, err := conn.RegisterTargets(params)

		if tfawserr.ErrMessageContains(err, "InvalidTarget", "") {
			return resource.RetryableError(fmt.Errorf("Error attaching instance to LB, retrying: %s", err))
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.RegisterTargets(params)
	}
	if err != nil {
		return fmt.Errorf("Error registering targets with target group: %s", err)
	}

	//lintignore:R016 // Allow legacy unstable ID usage in managed resource
	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-", d.Get("target_group_arn"))))

	return nil
}

func resourceAwsLbAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBV2Conn

	target := &elbv2.TargetDescription{
		Id: aws.String(d.Get("target_id").(string)),
	}

	if v, ok := d.GetOk("port"); ok {
		target.Port = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		target.AvailabilityZone = aws.String(v.(string))
	}

	params := &elbv2.DeregisterTargetsInput{
		TargetGroupArn: aws.String(d.Get("target_group_arn").(string)),
		Targets:        []*elbv2.TargetDescription{target},
	}

	_, err := conn.DeregisterTargets(params)
	if err != nil && !tfawserr.ErrMessageContains(err, elbv2.ErrCodeTargetGroupNotFoundException, "") {
		return fmt.Errorf("Error deregistering Targets: %s", err)
	}

	return nil
}

// resourceAwsLbAttachmentRead requires all of the fields in order to describe the correct
// target, so there is no work to do beyond ensuring that the target and group still exist.
func resourceAwsLbAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ELBV2Conn

	target := &elbv2.TargetDescription{
		Id: aws.String(d.Get("target_id").(string)),
	}

	if v, ok := d.GetOk("port"); ok {
		target.Port = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		target.AvailabilityZone = aws.String(v.(string))
	}

	resp, err := conn.DescribeTargetHealth(&elbv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(d.Get("target_group_arn").(string)),
		Targets:        []*elbv2.TargetDescription{target},
	})

	if err != nil {
		if tfawserr.ErrMessageContains(err, elbv2.ErrCodeTargetGroupNotFoundException, "") {
			log.Printf("[WARN] Target group does not exist, removing target attachment %s", d.Id())
			d.SetId("")
			return nil
		}
		if tfawserr.ErrMessageContains(err, elbv2.ErrCodeInvalidTargetException, "") {
			log.Printf("[WARN] Target does not exist, removing target attachment %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading Target Health: %s", err)
	}

	for _, targetDesc := range resp.TargetHealthDescriptions {
		if targetDesc == nil || targetDesc.Target == nil {
			continue
		}

		if aws.StringValue(targetDesc.Target.Id) == d.Get("target_id").(string) {
			// These will catch targets being removed by hand (draining as we plan) or that have been removed for a while
			// without trying to re-create ones that are just not in use. For example, a target can be `unused` if the
			// target group isnt assigned to anything, a scenario where we don't want to continuously recreate the resource.
			if targetDesc.TargetHealth == nil {
				continue
			}

			reason := aws.StringValue(targetDesc.TargetHealth.Reason)

			if reason == elbv2.TargetHealthReasonEnumTargetNotRegistered || reason == elbv2.TargetHealthReasonEnumTargetDeregistrationInProgress {
				log.Printf("[WARN] Target Attachment does not exist, recreating attachment")
				d.SetId("")
				return nil
			}
		}
	}

	if len(resp.TargetHealthDescriptions) != 1 {
		log.Printf("[WARN] Target does not exist, removing target attachment %s", d.Id())
		d.SetId("")
		return nil
	}

	return nil
}
