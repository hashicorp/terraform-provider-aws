package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsIotPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotPolicyAttachmentCreate,
		Read:   resourceAwsIotPolicyAttachmentRead,
		Delete: resourceAwsIotPolicyAttachmentDelete,
		Schema: map[string]*schema.Schema{
			"policy": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsIotPolicyAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	policyName := d.Get("policy").(string)
	target := d.Get("target").(string)

	_, err := conn.AttachPolicy(&iot.AttachPolicyInput{
		PolicyName: aws.String(policyName),
		Target:     aws.String(target),
	})

	if err != nil {
		log.Printf("[ERROR] Error attaching policy %s to target %s: %s", policyName, target, err)
		return err
	}

	d.SetId(fmt.Sprintf("%s|%s", policyName, target))
	return resourceAwsIotPolicyAttachmentRead(d, meta)
}

func listIotPolicyAttachmentPages(c *iot.IoT, input *iot.ListAttachedPoliciesInput,
	fn func(out *iot.ListAttachedPoliciesOutput, lastPage bool) bool) error {
	for {
		page, err := c.ListAttachedPolicies(input)
		if err != nil {
			return err
		}
		lastPage := page.NextMarker == nil

		shouldContinue := fn(page, lastPage)
		if !shouldContinue || lastPage {
			break
		}
		input.Marker = page.NextMarker
	}
	return nil
}

func resourceAwsIotPolicyAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	policyName := d.Get("policy").(string)
	target := d.Get("target").(string)

	var policy *iot.Policy

	err := listIotPolicyAttachmentPages(conn, &iot.ListAttachedPoliciesInput{
		PageSize:  aws.Int64(250),
		Recursive: aws.Bool(false),
		Target:    aws.String(target),
	}, func(out *iot.ListAttachedPoliciesOutput, lastPage bool) bool {
		for _, att := range out.Policies {
			name := aws.StringValue(att.PolicyName)
			if name == policyName {
				policy = att
				return false
			}
		}
		return true
	})

	if err != nil {
		log.Printf("[ERROR] Error listing policy attachments for target %s: %s", target, err)
		return err
	}

	if policy == nil {
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAwsIotPolicyAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	policyName := d.Get("policy").(string)
	target := d.Get("target").(string)

	_, err := conn.DetachPolicy(&iot.DetachPolicyInput{
		PolicyName: aws.String(policyName),
		Target:     aws.String(target),
	})

	if err != nil {
		log.Printf("[ERROR] Error detaching policy %s from target %s: %s", policyName, target, err)
		return err
	}

	return nil
}
