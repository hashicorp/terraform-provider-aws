package sqs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var (
	queueEmptyPolicyAttributes = map[string]string{
		sqs.QueueAttributeNamePolicy: "",
	}
)

func ResourceQueuePolicy() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceQueuePolicyUpsert,
		Read:   resourceQueuePolicyRead,
		Update: resourceQueuePolicyUpsert,
		Delete: resourceQueuePolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		MigrateState:  QueuePolicyMigrateState,
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
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

			"queue_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceQueuePolicyUpsert(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SQSConn

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return fmt.Errorf("policy (%s) is invalid JSON: %w", d.Get("policy").(string), err)
	}

	policyAttributes := map[string]string{
		sqs.QueueAttributeNamePolicy: policy,
	}

	url := d.Get("queue_url").(string)
	input := &sqs.SetQueueAttributesInput{
		Attributes: aws.StringMap(policyAttributes),
		QueueUrl:   aws.String(url),
	}

	log.Printf("[DEBUG] Setting SQS Queue Policy: %s", input)
	_, err = conn.SetQueueAttributes(input)

	if err != nil {
		return fmt.Errorf("error setting SQS Queue Policy (%s): %w", url, err)
	}

	d.SetId(url)

	err = waitQueueAttributesPropagated(conn, d.Id(), policyAttributes)

	if err != nil {
		return fmt.Errorf("error waiting for SQS Queue Policy (%s) to be set: %w", d.Id(), err)
	}

	return resourceQueuePolicyRead(d, meta)
}

func resourceQueuePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SQSConn

	outputRaw, err := tfresource.RetryWhenNotFound(queuePolicyReadTimeout, func() (interface{}, error) {
		return FindQueuePolicyByURL(conn, d.Id())
	})

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SQS Queue Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading SQS Queue Policy (%s): %w", d.Id(), err)
	}

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), outputRaw.(string))

	if err != nil {
		return err
	}

	d.Set("policy", policyToSet)

	d.Set("queue_url", d.Id())

	return nil
}

func resourceQueuePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SQSConn

	log.Printf("[DEBUG] Deleting SQS Queue Policy: %s", d.Id())
	_, err := conn.SetQueueAttributes(&sqs.SetQueueAttributesInput{
		Attributes: aws.StringMap(queueEmptyPolicyAttributes),
		QueueUrl:   aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting SQS Queue Policy (%s): %w", d.Id(), err)
	}

	err = waitQueueAttributesPropagated(conn, d.Id(), queueEmptyPolicyAttributes)

	if err != nil {
		return fmt.Errorf("error waiting for SQS Queue Policy (%s) to delete: %w", d.Id(), err)
	}

	return nil
}
