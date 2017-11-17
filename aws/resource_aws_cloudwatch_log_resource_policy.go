package aws

import (
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

func resourceAwsCloudWatchLogResourcePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudWatchLogResourcePolicyPut,
		Update: resourceAwsCloudWatchLogResourcePolicyPut,
		Read:   resourceAwsCloudWatchLogResourcePolicyRead,
		Delete: resourceAwsCloudWatchLogResourcePolicyDelete,

		Schema: map[string]*schema.Schema{
			"policy_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"policy_document": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validateJsonString,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
		},
	}
}

func resourceAwsCloudWatchLogResourcePolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchlogsconn

	input := &cloudwatchlogs.PutResourcePolicyInput{}

	if v, ok := d.GetOk("policy_name"); ok {
		input.PolicyName = aws.String(v.(string))
	}
	if v, ok := d.GetOk("policy_document"); ok {
		input.PolicyDocument = aws.String(v.(string))
	}

	resp, err := conn.PutResourcePolicy(input)
	if err != nil {
		return err
	}

	d.SetId(*resp.ResourcePolicy.PolicyName)
	return resourceAwsCloudWatchLogResourcePolicyRead(d, meta)
}

func resourceAwsCloudWatchLogResourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchlogsconn

	resourcePolicy, exists, err := lookupCloudWatchLogResourcePolicy(conn, d.Id(), nil)
	if err != nil {
		return err
	}

	if !exists {
		d.SetId("")
		return nil
	}

	if resourcePolicy.PolicyDocument != nil {
		d.Set("policy_document", resourcePolicy.PolicyDocument)
	}

	return nil
}

func resourceAwsCloudWatchLogResourcePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchlogsconn

	input := &cloudwatchlogs.DeleteResourcePolicyInput{
		PolicyName: aws.String(d.Id()),
	}

	_, err := conn.DeleteResourcePolicy(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case cloudwatchlogs.ErrCodeResourceNotFoundException:
				d.SetId("")
				return nil
			default:
				return err
			}
		}
		return err
	}

	d.SetId("")
	return nil
}

func lookupCloudWatchLogResourcePolicy(conn *cloudwatchlogs.CloudWatchLogs, name string, nextToken *string) (*cloudwatchlogs.ResourcePolicy, bool, error) {
	input := &cloudwatchlogs.DescribeResourcePoliciesInput{
		NextToken: nextToken,
	}

	resp, err := conn.DescribeResourcePolicies(input)
	if err != nil {
		return nil, true, err
	}

	for _, resourcePolicy := range resp.ResourcePolicies {
		if *resourcePolicy.PolicyName == name {
			return resourcePolicy, true, nil
		}
	}

	if resp.NextToken != nil {
		return lookupCloudWatchLogResourcePolicy(conn, name, resp.NextToken)
	}

	return nil, false, nil
}
