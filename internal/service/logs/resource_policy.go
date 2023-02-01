package logs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func init() {
	_sp.registerSDKResourceFactory("aws_cloudwatch_log_resource_policy", resourceResourcePolicy)
}

func resourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourcePolicyPut,
		ReadWithoutTimeout:   resourceResourcePolicyRead,
		UpdateWithoutTimeout: resourceResourcePolicyPut,
		DeleteWithoutTimeout: resourceResourcePolicyDelete,

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

func resourceResourcePolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn()

	policy, err := structure.NormalizeJsonString(d.Get("policy_document").(string))

	if err != nil {
		return diag.Errorf("policy (%s) is invalid JSON: %s", policy, err)
	}

	name := d.Get("policy_name").(string)
	input := &cloudwatchlogs.PutResourcePolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(name),
	}

	output, err := conn.PutResourcePolicyWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating CloudWatch Logs Resource Policy (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.ResourcePolicy.PolicyName))

	return resourceResourcePolicyRead(ctx, d, meta)
}

func resourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn()

	resourcePolicy, err := FindResourcePolicyByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Resource Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CloudWatch Logs Resource Policy (%s): %s", d.Id(), err)
	}

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy_document").(string), aws.StringValue(resourcePolicy.PolicyDocument))

	if err != nil {
		return diag.Errorf("while setting policy (%s), encountered: %s", policyToSet, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return diag.Errorf("policy (%s) is invalid JSON: %s", policyToSet, err)
	}

	d.Set("policy_document", policyToSet)

	return nil
}

func resourceResourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn()

	log.Printf("[DEBUG] Deleting CloudWatch Logs Resource Policy: %s", d.Id())
	_, err := conn.DeleteResourcePolicyWithContext(ctx, &cloudwatchlogs.DeleteResourcePolicyInput{
		PolicyName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting CloudWatch Logs Resource Policy (%s): %s", d.Id(), err)
	}

	return nil
}

func FindResourcePolicyByName(ctx context.Context, conn *cloudwatchlogs.CloudWatchLogs, name string) (*cloudwatchlogs.ResourcePolicy, error) {
	input := &cloudwatchlogs.DescribeResourcePoliciesInput{}
	var output *cloudwatchlogs.ResourcePolicy

	err := describeResourcePoliciesPages(ctx, conn, input, func(page *cloudwatchlogs.DescribeResourcePoliciesOutput, lastPage bool) bool {
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
