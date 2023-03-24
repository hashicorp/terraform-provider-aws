package sns

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_sns_topic_data_protection_policy")
func ResourceTopicDataProtectionPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: ResourceTopicDataProtectionPolicyUpsert,
		ReadWithoutTimeout:   ResourceTopicDataProtectionPolicyRead,
		UpdateWithoutTimeout: ResourceTopicDataProtectionPolicyUpsert,
		DeleteWithoutTimeout: ResourceTopicDataProtectionPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"policy": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func ResourceTopicDataProtectionPolicyUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSConn()

	topicArn := d.Get("arn").(string)
	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", d.Get("policy").(string), err)
	}

	input := &sns.PutDataProtectionPolicyInput{
		DataProtectionPolicy: aws.String(policy),
		ResourceArn:          aws.String(topicArn),
	}

	output, err := conn.PutDataProtectionPolicyWithContext(ctx, input)

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "Something went wrong creating the SNS Data Protection Policy")
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SNS Data Protection Policy (%s): %s", d.Id(), err)
	}

	d.SetId(topicArn)

	return ResourceTopicDataProtectionPolicyRead(ctx, d, meta)
}

func ResourceTopicDataProtectionPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSConn()

	output, err := conn.GetDataProtectionPolicyWithContext(ctx, &sns.GetDataProtectionPolicyInput{
		ResourceArn: aws.String(d.Id()),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, sns.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] SNS Data Protection Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SNS Data Protection Policy: %s", err)
	}

	if output == nil || output.DataProtectionPolicy == nil {
		return sdkdiag.AppendErrorf(diags, "reading SNS Data Protection Policy (%s): empty output", d.Id())
	}

	dataProtectionPolicy := output.DataProtectionPolicy

	d.Set("arn", d.Id())
	d.Set("policy", dataProtectionPolicy)

	return diags
}

func ResourceTopicDataProtectionPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSConn()

	topicArn := d.Get("arn").(string)

	input := &sns.PutDataProtectionPolicyInput{
		DataProtectionPolicy: aws.String(""),
		ResourceArn:          aws.String(topicArn),
	}

	output, err := conn.PutDataProtectionPolicyWithContext(ctx, input)

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "Something went wrong creating the SNS Data Protection Policy")
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SNS Data Protection Policy (%s): %s", d.Id(), err)
	}

	return diags
}
