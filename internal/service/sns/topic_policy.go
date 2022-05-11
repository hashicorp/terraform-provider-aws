package sns

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTopicPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceTopicPolicyUpsert,
		Read:   resourceTopicPolicyRead,
		Update: resourceTopicPolicyUpsert,
		Delete: resourceTopicPolicyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceTopicPolicyUpsert(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SNSConn

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", d.Get("policy").(string), err)
	}

	arn := d.Get("arn").(string)

	err = putTopicPolicy(conn, arn, policy)

	if err != nil {
		return err
	}

	if d.IsNewResource() {
		d.SetId(arn)
	}

	return resourceTopicPolicyRead(d, meta)
}

func resourceTopicPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SNSConn

	attributes, err := FindTopicAttributesByARN(conn, d.Id())

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
		return fmt.Errorf("error reading SNS Topic Policy (%s): %w", d.Id(), err)
	}

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), policy)

	if err != nil {
		return err
	}

	d.Set("arn", attributes[TopicAttributeNameTopicArn])
	d.Set("owner", attributes[TopicAttributeNameOwner])
	d.Set("policy", policyToSet)

	return nil
}

func resourceTopicPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SNSConn

	// It is impossible to delete a policy or set to empty
	// (confirmed by AWS Support representative)
	// so we instead set it back to the default one.
	err := putTopicPolicy(conn, d.Id(), defaultTopicPolicy(d.Id(), d.Get("owner").(string)))

	if err != nil {
		return err
	}

	return nil
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

func putTopicPolicy(conn *sns.SNS, arn string, policy string) error {
	return putTopicAttribute(conn, arn, TopicAttributeNamePolicy, policy)
}
