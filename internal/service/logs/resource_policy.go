package logs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
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
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceResourcePolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	policy, err := structure.NormalizeJsonString(d.Get("policy_document").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policy, err)
	}

	policyName := d.Get("policy_name").(string)

	input := &cloudwatchlogs.PutResourcePolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(policyName),
	}

	log.Printf("[DEBUG] Writing CloudWatch log resource policy: %#v", input)
	_, err = conn.PutResourcePolicy(input)

	if err != nil {
		return fmt.Errorf("Writing CloudWatch log resource policy failed: %s", err.Error())
	}

	d.SetId(policyName)
	return resourceResourcePolicyRead(d, meta)
}

func resourceResourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn
	policyName := d.Get("policy_name").(string)
	resourcePolicy, exists, err := LookupResourcePolicy(conn, policyName, nil)
	if err != nil {
		return err
	}

	if !exists {
		d.SetId("")
		return nil
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy_document").(string), aws.StringValue(resourcePolicy.PolicyDocument))

	if err != nil {
		return fmt.Errorf("while setting policy (%s), encountered: %w", policyToSet, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", policyToSet, err)
	}

	d.Set("policy_document", policyToSet)

	return nil
}

func resourceResourcePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn
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
