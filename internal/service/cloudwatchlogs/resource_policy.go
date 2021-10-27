package cloudwatchlogs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceResourcePolicyPut,
		Read:   resourceResourcePolicyRead,
		Update: resourceResourcePolicyPut,
		Delete: resourceResourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("policy_name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"policy_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"policy_document": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validResourcePolicyDocument,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
			},
		},
	}
}

func resourceResourcePolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudWatchLogsConn

	policyName := d.Get("policy_name").(string)

	input := &cloudwatchlogs.PutResourcePolicyInput{
		PolicyDocument: aws.String(d.Get("policy_document").(string)),
		PolicyName:     aws.String(policyName),
	}

	log.Printf("[DEBUG] Writing CloudWatch log resource policy: %#v", input)
	_, err := conn.PutResourcePolicy(input)

	if err != nil {
		return fmt.Errorf("Writing CloudWatch log resource policy failed: %s", err.Error())
	}

	d.SetId(policyName)
	return resourceResourcePolicyRead(d, meta)
}

func resourceResourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudWatchLogsConn
	policyName := d.Get("policy_name").(string)
	resourcePolicy, exists, err := LookupResourcePolicy(conn, policyName, nil)
	if err != nil {
		return err
	}

	if !exists {
		d.SetId("")
		return nil
	}

	d.Set("policy_document", resourcePolicy.PolicyDocument)

	return nil
}

func resourceResourcePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudWatchLogsConn
	input := cloudwatchlogs.DeleteResourcePolicyInput{
		PolicyName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting CloudWatch log resource policy: %#v", input)
	_, err := conn.DeleteResourcePolicy(&input)
	if err != nil {
		return fmt.Errorf("Deleting CloudWatch log resource policy '%s' failed: %s", *input.PolicyName, err.Error())
	}
	return nil
}

func LookupResourcePolicy(conn *cloudwatchlogs.CloudWatchLogs,
	name string, nextToken *string) (*cloudwatchlogs.ResourcePolicy, bool, error) {
	input := &cloudwatchlogs.DescribeResourcePoliciesInput{
		NextToken: nextToken,
	}
	log.Printf("[DEBUG] Reading CloudWatch log resource policies: %#v", input)
	resp, err := conn.DescribeResourcePolicies(input)
	if err != nil {
		return nil, true, err
	}

	for _, resourcePolicy := range resp.ResourcePolicies {
		if aws.StringValue(resourcePolicy.PolicyName) == name {
			return resourcePolicy, true, nil
		}
	}

	if resp.NextToken != nil {
		return LookupResourcePolicy(conn, name, resp.NextToken)
	}

	return nil, false, nil
}
