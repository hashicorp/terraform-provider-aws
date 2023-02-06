package sns

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTopicPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTopicPolicyUpsert,
		ReadWithoutTimeout:   resourceTopicPolicyRead,
		UpdateWithoutTimeout: resourceTopicPolicyUpsert,
		DeleteWithoutTimeout: resourceTopicPolicyDelete,

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
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
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

func resourceTopicPolicyUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SNSConn()

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return diag.Errorf("policy (%s) is invalid JSON: %s", d.Get("policy").(string), err)
	}

	arn := d.Get("arn").(string)

	err = putTopicPolicy(ctx, conn, arn, policy)

	if err != nil {
		return diag.FromErr(err)
	}

	if d.IsNewResource() {
		d.SetId(arn)
	}

	return resourceTopicPolicyRead(ctx, d, meta)
}

func resourceTopicPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SNSConn()

	attributes, err := FindTopicAttributesByARN(ctx, conn, d.Id())

	var policy string

	if err == nil {
		policy = attributes[TopicAttributeNamePolicy]

		if policy == "" {
			err = tfresource.NewEmptyResultError(d.Id())
		}
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SNS Topic Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading SNS Topic Policy (%s): %s", d.Id(), err)
	}

	d.Set("arn", attributes[TopicAttributeNameTopicARN])
	d.Set("owner", attributes[TopicAttributeNameOwner])

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), policy)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("policy", policyToSet)

	return nil
}

func resourceTopicPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SNSConn()

	// It is impossible to delete a policy or set to empty
	// (confirmed by AWS Support representative)
	// so we instead set it back to the default one.
	return diag.FromErr(putTopicPolicy(ctx, conn, d.Id(), defaultTopicPolicy(d.Id(), d.Get("owner").(string))))
}

func defaultTopicPolicy(topicArn, accountId string) string {
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
`, topicArn, accountId)
}

func putTopicPolicy(ctx context.Context, conn *sns.SNS, arn string, policy string) error {
	return putTopicAttribute(ctx, conn, arn, TopicAttributeNamePolicy, policy)
}
