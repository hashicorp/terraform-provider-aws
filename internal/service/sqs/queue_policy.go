package sqs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var (
	sqsQueueEmptyPolicyAttributes = map[string]string{
		sqs.QueueAttributeNamePolicy: "",
	}
)

func ResourceQueuePolicy() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceAwsSqsQueuePolicyUpsert,
		Read:   resourceQueuePolicyRead,
		Update: resourceAwsSqsQueuePolicyUpsert,
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
			},

			"queue_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsSqsQueuePolicyUpsert(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SQSConn

	policyAttributes := map[string]string{
		sqs.QueueAttributeNamePolicy: d.Get("policy").(string),
	}
	url := d.Get("queue_url").(string)
	input := &sqs.SetQueueAttributesInput{
		Attributes: aws.StringMap(policyAttributes),
		QueueUrl:   aws.String(url),
	}

	log.Printf("[DEBUG] Setting SQS Queue Policy: %s", input)
	_, err := conn.SetQueueAttributes(input)

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

	policy, err := FindQueuePolicyByURL(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SQS Queue Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading SQS Queue Policy (%s): %w", d.Id(), err)
	}

	d.Set("policy", policy)
	d.Set("queue_url", d.Id())

	return nil
}

func resourceQueuePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SQSConn

	log.Printf("[DEBUG] Deleting SQS Queue Policy: %s", d.Id())
	_, err := conn.SetQueueAttributes(&sqs.SetQueueAttributesInput{
		Attributes: aws.StringMap(sqsQueueEmptyPolicyAttributes),
		QueueUrl:   aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting SQS Queue Policy (%s): %w", d.Id(), err)
	}

	err = waitQueueAttributesPropagated(conn, d.Id(), sqsQueueEmptyPolicyAttributes)

	if err != nil {
		return fmt.Errorf("error waiting for SQS Queue Policy (%s) to delete: %w", d.Id(), err)
	}

	return nil
}
