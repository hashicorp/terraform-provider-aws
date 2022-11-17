package logs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			"policy_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

	name := d.Get("policy_name").(string)
	input := &cloudwatchlogs.PutResourcePolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(name),
	}

	output, err := conn.PutResourcePolicy(input)

	if err != nil {
		return fmt.Errorf("creating CloudWatch Logs Resource Policy (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.ResourcePolicy.PolicyName))

	return resourceResourcePolicyRead(d, meta)
}

func resourceResourcePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LogsConn

	resourcePolicy, err := FindResourcePolicyByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Resource Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading CloudWatch Logs Resource Policy (%s): %w", d.Id(), err)
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

	log.Printf("[DEBUG] Deleting CloudWatch Logs Resource Policy: %s", d.Id())
	_, err := conn.DeleteResourcePolicy(&cloudwatchlogs.DeleteResourcePolicyInput{
		PolicyName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting CloudWatch Logs Resource Policy (%s): %w", d.Id(), err)
	}

	return nil
}

func FindResourcePolicyByName(conn *cloudwatchlogs.CloudWatchLogs, name string) (*cloudwatchlogs.ResourcePolicy, error) {
	input := &cloudwatchlogs.DescribeResourcePoliciesInput{}
	var output *cloudwatchlogs.ResourcePolicy

	err := describeResourcePoliciesPages(conn, input, func(page *cloudwatchlogs.DescribeResourcePoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResourcePolicies {
			if aws.StringValue(v.PolicyName) == name {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
