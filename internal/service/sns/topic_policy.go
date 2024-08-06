// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sns_topic_policy")
func resourceTopicPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTopicPolicyUpsert,
		ReadWithoutTimeout:   resourceTopicPolicyRead,
		UpdateWithoutTimeout: resourceTopicPolicyUpsert,
		DeleteWithoutTimeout: resourceTopicPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrOwner: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPolicy: {
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

func resourceTopicPolicyUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", d.Get(names.AttrPolicy).(string), err)
	}

	arn := d.Get(names.AttrARN).(string)
	err = putTopicPolicy(ctx, conn, arn, policy)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.IsNewResource() {
		d.SetId(arn)
	}

	return append(diags, resourceTopicPolicyRead(ctx, d, meta)...)
}

func resourceTopicPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	attributes, err := findTopicAttributesWithValidAWSPrincipalsByARN(ctx, conn, d.Id())

	var policy string

	if err == nil {
		policy = attributes[topicAttributeNamePolicy]

		if policy == "" {
			err = tfresource.NewEmptyResultError(d.Id())
		}
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SNS Topic Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SNS Topic Policy (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, attributes[topicAttributeNameTopicARN])
	d.Set(names.AttrOwner, attributes[topicAttributeNameOwner])

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), policy)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	return diags
}

func resourceTopicPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SNSClient(ctx)

	// It is impossible to delete a policy or set to empty
	// (confirmed by AWS Support representative)
	// so we instead set it back to the default one.
	err := putTopicPolicy(ctx, conn, d.Id(), defaultTopicPolicy(d.Id(), d.Get(names.AttrOwner).(string)))

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	return sdkdiag.AppendFromErr(diags, err)
}

func defaultTopicPolicy(topicARN, accountID string) string {
	return fmt.Sprintf(`{
  "Version": "2008-10-17",
  "Id": "__default_policy_ID",
  "Statement": [
    {
      "Sid": "__default_statement_ID",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": [
        "SNS:GetTopicAttributes",
        "SNS:SetTopicAttributes",
        "SNS:AddPermission",
        "SNS:RemovePermission",
        "SNS:DeleteTopic",
        "SNS:Subscribe",
        "SNS:ListSubscriptionsByTopic",
        "SNS:Publish",
        "SNS:Receive"
      ],
      "Resource": %[1]q,
      "Condition": {
        "StringEquals": {
          "AWS:SourceOwner": %[2]q
        }
      }
    }
  ]
}
`, topicARN, accountID)
}

func putTopicPolicy(ctx context.Context, conn *sns.Client, arn string, policy string) error {
	return putTopicAttribute(ctx, conn, arn, topicAttributeNamePolicy, policy)
}
