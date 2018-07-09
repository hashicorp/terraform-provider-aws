package aws

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
)

func resourceAwsSesDomainIdentityPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesDomainIdentityPolicyCreate,
		Read:   resourceAwsSesDomainIdentityPolicyRead,
		Update: resourceAwsSesDomainIdentityPolicyUpdate,
		Delete: resourceAwsSesDomainIdentityPolicyDelete,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validateJsonString,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
		},
	}
}

func resourceAwsSesDomainIdentityPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn

	arn := d.Get("arn").(string)
	policyName := d.Get("name").(string)
	policy := d.Get("policy").(string)

	req := ses.PutIdentityPolicyInput{
		Identity: aws.String(arn),
		PolicyName: aws.String(policyName),
		Policy: aws.String(policy),
	}

	_, err := conn.PutIdentityPolicy(&req)
	if err != nil {
		return err
	}

	d.SetId(resource.PrefixedUniqueId(fmt.Sprintf("%s-", policyName)))
	return resourceAwsSesDomainIdentityPolicyRead(d, meta)
}

func resourceAwsSesDomainIdentityPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn

	arn := d.Get("arn").(string)
	policyName := d.Get("name").(string)
	policy := d.Get("policy").(string)

	req := ses.PutIdentityPolicyInput{
		Identity:   aws.String(arn),
		PolicyName: aws.String(policyName),
		Policy:     aws.String(policy),
	}

	_, err := conn.PutIdentityPolicy(&req)
	if err != nil {
		return err
	}

	return resourceAwsSesDomainIdentityPolicyRead(d, meta)
}

func resourceAwsSesDomainIdentityPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn

	arn := d.Get("arn").(string)
	policyName := d.Get("name").(string)
	policyNames := make([]*string, 1)
	policyNames[0] = aws.String(policyName)

	policiesOutput, err := conn.GetIdentityPolicies(&ses.GetIdentityPoliciesInput{
		Identity: aws.String(arn),
		PolicyNames: policyNames,
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFound" {
			log.Printf("[WARN] SES Domain Identity Policy (%s) not found, error code (404)", policyName)
			d.SetId("")
			return nil
		}

		return err
	}

	if policiesOutput.Policies == nil {
		log.Printf("[WARN] SES Domain Identity Policy (%s) not found (nil)", policyName)
		d.SetId("")
		return nil
	}
	policies := policiesOutput.Policies

	policy, ok := policies[*aws.String(policyName)]
	if !ok {
		log.Printf("[WARN] SES Domain Identity Policy (%s) not found in attributes", policyName)
		d.SetId("")
		return nil
	}

	d.Set("policy", policy)
	return nil
}

func resourceAwsSesDomainIdentityPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sesConn

	arn := d.Get("arn").(string)
	policyName := d.Get("name").(string)

	req := ses.DeleteIdentityPolicyInput{
		Identity: aws.String(arn),
		PolicyName: aws.String(policyName),
	}

	log.Printf("[DEBUG] Deleting SES Domain Identity Policy: %s", req)
	_, err := conn.DeleteIdentityPolicy(&req)
	return err
}
