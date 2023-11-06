// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_sns_topic_data_protection_policy")
func resourceTopicDataProtectionPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTopicDataProtectionPolicyUpsert,
		ReadWithoutTimeout:   resourceTopicDataProtectionPolicyRead,
		UpdateWithoutTimeout: resourceTopicDataProtectionPolicyUpsert,
		DeleteWithoutTimeout: resourceTopicDataProtectionPolicyDelete,

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

func resourceTopicDataProtectionPolicyUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	topicARN := d.Get("arn").(string)
	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", d.Get("policy").(string), err)
	}

	input := &sns.PutDataProtectionPolicyInput{
		DataProtectionPolicy: aws.String(policy),
		ResourceArn:          aws.String(topicARN),
	}

	_, err = conn.PutDataProtectionPolicy(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SNS Data Protection Policy (%s): %s", topicARN, err)
	}

	if d.IsNewResource() {
		d.SetId(topicARN)
	}

	return append(diags, resourceTopicDataProtectionPolicyRead(ctx, d, meta)...)
}

func resourceTopicDataProtectionPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	output, err := findDataProtectionPolicyByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SNS Data Protection Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SNS Data Protection Policy (%s): %s", d.Id(), err)
	}

	d.Set("arn", d.Id())
	d.Set("policy", output)

	return diags
}

func resourceTopicDataProtectionPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	_, err := conn.PutDataProtectionPolicy(ctx, &sns.PutDataProtectionPolicyInput{
		DataProtectionPolicy: aws.String(""),
		ResourceArn:          aws.String(d.Get("arn").(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SNS Data Protection Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findDataProtectionPolicyByARN(ctx context.Context, conn *sns.Client, arn string) (*string, error) {
	input := &sns.GetDataProtectionPolicyInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetDataProtectionPolicy(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil || aws.ToString(output.DataProtectionPolicy) == "" {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DataProtectionPolicy, nil
}
