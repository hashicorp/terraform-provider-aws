package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsLbTargetGroupAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLbAttachmentCreate,
		Read:   resourceAwsLbAttachmentRead,
		Delete: resourceAwsLbAttachmentDelete,
		Update: resourceAwsLbAttachmentUpdate,

		Schema: map[string]*schema.Schema{
			"target_group_arn": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"target_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},

			"port": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
			},

			"instances": &schema.Schema{
				Type:     schema.TypeSet,
				ForceNew: false,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
				Set:      schema.HashString,
			},

			"availability_zone": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
		},
	}
}

func resourceAwsLbAttachmentUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Update Target  %s (%d) with Target Group %s", d.Get("target_id").(string),
		d.Get("port").(int), d.Get("target_group_arn").(string))

	elbconn := meta.(*AWSClient).elbv2conn

	if _, ok := d.GetOk("instances"); !ok {
		log.Printf("[ERROR] Call update and instances is empty for Target Group %s", d.Get("target_group_arn").(string))
		return nil
	}

	resp, err := elbconn.DescribeTargetHealth(&elbv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(d.Get("target_group_arn").(string)),
	})

	if err != nil {
		return errwrap.Wrapf("Error reading Target Health: {{err}}", err)
	}

	targets := d.Get("instances").(*schema.Set)

	paramsDereg := &elbv2.DeregisterTargetsInput{
		TargetGroupArn: aws.String(d.Get("target_group_arn").(string)),
	}

	for _, targetDeployed := range resp.TargetHealthDescriptions {
		log.Printf("[DEBUG] Look for targets to remove from target group : current loop %s", *(targetDeployed.Target.Id))

		//Target deployed not in  given in the resource  -> remove it
		if !targets.Contains(*(targetDeployed.Target.Id)) {
			targetPortToAdd := &elbv2.TargetDescription{
				Id: aws.String(*(targetDeployed.Target.Id)),
			}
			if v, ok := d.GetOk("port"); ok {
				targetPortToAdd.Port = aws.Int64(int64(v.(int)))
			}
			paramsDereg.Targets = append(paramsDereg.Targets, targetPortToAdd)

		}
	}

	if len(paramsDereg.Targets) > 0 {
		log.Printf("[DEBUG] %d target to remove from %s", len(paramsDereg.Targets), d.Get("target_group_arn").(string))
		_, err = elbconn.DeregisterTargets(paramsDereg)
		if err != nil {
			return errwrap.Wrapf("Error deregistering Targets: {{err}}", err)
		}
	}

	//Get the ID from the TargetHealthDescription
	thd := make(map[string]bool)
	for _, v := range resp.TargetHealthDescriptions {
		thd[*(v.Target.Id)] = true
	}
	paramsRegister := &elbv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(d.Get("target_group_arn").(string)),
	}
	for _, target := range targets.List() {
		if _, ok := thd[target.(string)]; !ok {
			targetPortToAdd := &elbv2.TargetDescription{
				Id: aws.String(target.(string)),
			}
			if v, ok := d.GetOk("port"); ok {
				targetPortToAdd.Port = aws.Int64(int64(v.(int)))
			}
			paramsRegister.Targets = append(paramsRegister.Targets, targetPortToAdd)
		}
	}
	if len(paramsRegister.Targets) > 0 {
		log.Printf("[DEBUG] %d target to add to %s", len(paramsRegister.Targets), d.Get("target_group_arn").(string))
		_, err = elbconn.RegisterTargets(paramsRegister)
		if err != nil {
			return errwrap.Wrapf("Error registering targets with target group: {{err}}", err)
		}
	}
	return nil
}

func resourceAwsLbAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	params := &elbv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(d.Get("target_group_arn").(string)),
	}

	if _, ok := d.GetOk("target_id"); ok {
		targetPortToAdd := &elbv2.TargetDescription{
			Id: aws.String(d.Get("target_id").(string)),
		}
		if v, ok := d.GetOk("port"); ok {
			targetPortToAdd.Port = aws.Int64(int64(v.(int)))
		}
		if v, ok := d.GetOk("availability_zone"); ok {
			targetPortToAdd.AvailabilityZone = aws.String(v.(string))
		}

		params.Targets = append(params.Targets, targetPortToAdd)
		log.Printf("[INFO] Registering Target %s (%d) with Target Group %s", d.Get("target_id").(string),
			d.Get("port").(int), d.Get("target_group_arn").(string))
	} else {
		targets := d.Get("instances").(*schema.Set)

		for _, target := range targets.List() {
			targetPortToAdd := &elbv2.TargetDescription{
				Id: aws.String(target.(string)),
			}
			if v, ok := d.GetOk("port"); ok {
				targetPortToAdd.Port = aws.Int64(int64(v.(int)))
			}
			if v, ok := d.GetOk("availability_zone"); ok {
				targetPortToAdd.AvailabilityZone = aws.String(v.(string))
			}

			params.Targets = append(params.Targets, targetPortToAdd)
		}
	}

	_, err := elbconn.RegisterTargets(params)
	if err != nil {
		return fmt.Errorf("Error registering targets with target group: %s", err)
	}

	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-", d.Get("target_group_arn"))))

	return nil
}

func resourceAwsLbAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	params := &elbv2.DeregisterTargetsInput{
		TargetGroupArn: aws.String(d.Get("target_group_arn").(string)),
	}

	if _, ok := d.GetOk("target_id"); ok {
		targetPortToAdd := &elbv2.TargetDescription{
			Id: aws.String(d.Get("target_id").(string)),
		}
		if v, ok := d.GetOk("port"); ok {
			targetPortToAdd.Port = aws.Int64(int64(v.(int)))
		}
		params.Targets = append(params.Targets, targetPortToAdd)
		log.Printf("[INFO] Registering Target %s (%d) with Target Group %s", d.Get("target_id").(string),
			d.Get("port").(int), d.Get("target_group_arn").(string))
	} else {
		targets := d.Get("instances").(*schema.Set)

		for _, target := range targets.List() {
			targetPortToAdd := &elbv2.TargetDescription{
				Id: aws.String(target.(string)),
			}
			if v, ok := d.GetOk("port"); ok {
				targetPortToAdd.Port = aws.Int64(int64(v.(int)))
			}
			if v, ok := d.GetOk("availability_zone"); ok {
				targetPortToAdd.AvailabilityZone = aws.String(v.(string))
			}

			params.Targets = append(params.Targets, targetPortToAdd)
		}
	}
	_, err := elbconn.DeregisterTargets(params)
	if err != nil && !isAWSErr(err, elbv2.ErrCodeTargetGroupNotFoundException, "") {
		return fmt.Errorf("Error deregistering Targets: %s", err)
	}

	return nil
}

// resourceAwsLbAttachmentRead requires all of the fields in order to describe the correct
// target, so there is no work to do beyond ensuring that the target and group still exist.
func resourceAwsLbAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	elbconn := meta.(*AWSClient).elbv2conn

	params := &elbv2.DescribeTargetHealthInput{
		TargetGroupArn: aws.String(d.Get("target_group_arn").(string)),
	}

	if _, ok := d.GetOk("target_id"); ok {
		targetPortToAdd := &elbv2.TargetDescription{
			Id: aws.String(d.Get("target_id").(string)),
		}
		if v, ok := d.GetOk("port"); ok {
			targetPortToAdd.Port = aws.Int64(int64(v.(int)))
		}
		params.Targets = append(params.Targets, targetPortToAdd)
		log.Printf("[INFO] Registering Target %s (%d) with Target Group %s", d.Get("target_id").(string),
			d.Get("port").(int), d.Get("target_group_arn").(string))
	} else {
		targets := d.Get("instances").(*schema.Set)

		for _, target := range targets.List() {
			targetPortToAdd := &elbv2.TargetDescription{
				Id: aws.String(target.(string)),
			}
			if v, ok := d.GetOk("port"); ok {
				targetPortToAdd.Port = aws.Int64(int64(v.(int)))
			}
			if v, ok := d.GetOk("availability_zone"); ok {
				targetPortToAdd.AvailabilityZone = aws.String(v.(string))
			}

			params.Targets = append(params.Targets, targetPortToAdd)
		}
	}

	resp, err := elbconn.DescribeTargetHealth(params)

	if err != nil {
		if isAWSErr(err, elbv2.ErrCodeTargetGroupNotFoundException, "") {
			log.Printf("[WARN] Target group does not exist, removing target attachment %s", d.Id())
			d.SetId("")
			return nil
		}
		if isAWSErr(err, elbv2.ErrCodeInvalidTargetException, "") {
			log.Printf("[WARN] Target does not exist, removing target attachment %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading Target Health: %s", err)
	}

	if len(resp.TargetHealthDescriptions) < 1 {
		log.Printf("[WARN] Target does not exist, removing target attachment %s", d.Id())
		d.SetId("")
		return nil
	}

	return nil
}
