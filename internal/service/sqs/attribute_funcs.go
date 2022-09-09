package sqs

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"log"
)

func generateQueueAttributeUpsertFunc(attributeName string) func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		conn := meta.(*conns.AWSClient).SQSConn

		attrValue, err := structure.NormalizeJsonString(d.Get(getSchemaKey(attributeName)).(string))

		if err != nil {
			return diag.FromErr(fmt.Errorf("%s (%s) is invalid JSON: %w", attributeName, d.Get(getSchemaKey(attributeName)).(string), err))
		}

		var attributes map[string]string

		switch attributeName {
		case sqs.QueueAttributeNamePolicy:
			attributes = map[string]string{
				sqs.QueueAttributeNamePolicy: attrValue,
			}
		case sqs.QueueAttributeNameRedrivePolicy:
			attributes = map[string]string{
				sqs.QueueAttributeNameRedrivePolicy: attrValue,
			}
		default:
			return diag.FromErr(fmt.Errorf("%s is an invalid SQS Queue attribute name", attributeName))
		}

		url := d.Get("queue_url").(string)
		input := &sqs.SetQueueAttributesInput{
			Attributes: aws.StringMap(attributes),
			QueueUrl:   aws.String(url),
		}

		log.Printf("[DEBUG] Setting SQS Queue Attribute '%s': %s", attributeName, input)
		_, err = conn.SetQueueAttributesWithContext(ctx, input)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error setting SQS Queue Attribute '%s' (%s): %w", attributeName, url, err))
		}

		d.SetId(url)

		err = waitQueueAttributesPropagatedWithContext(ctx, conn, d.Id(), attributes)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error waiting for SQS Queue Attribute '%s' (%s) to be set: %w", attributeName, d.Id(), err))
		}

		return generateQueueAttributeReadFunc(attributeName)(ctx, d, meta)
	}
}
func generateQueueAttributeReadFunc(attributeName string) func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		conn := meta.(*conns.AWSClient).SQSConn

		outputRaw, err := tfresource.RetryWhenNotFound(queueAttributeReadTimeout, func() (interface{}, error) {
			return FindQueueAttributeByURL(ctx, conn, d.Id(), attributeName)
		})

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] SQS Queue Policy (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		if err != nil {
			return diag.FromErr(fmt.Errorf("error reading SQS Queue Attribute '%s' (%s): %w", attributeName, d.Id(), err))
		}

		var attributeToSet string
		switch attributeName {
		case sqs.QueueAttributeNamePolicy:
			attributeToSet, err = verify.PolicyToSet(d.Get(getSchemaKey(attributeName)).(string), outputRaw.(string))
			if err != nil {
				return diag.FromErr(err)
			}
		case sqs.QueueAttributeNameRedrivePolicy:
			if BytesEqual([]byte(d.Get(getSchemaKey(attributeName)).(string)), []byte(outputRaw.(string))) {
				attributeToSet = d.Get(getSchemaKey(attributeName)).(string)
			} else {
				attributeToSet = outputRaw.(string)
			}
		default:
			return diag.FromErr(fmt.Errorf("%s is an invalid SQS Queue attribute name", attributeName))
		}

		d.Set(getSchemaKey(attributeName), attributeToSet)

		d.Set("queue_url", d.Id())

		return nil
	}
}

func generateQueueAttributeDeleteFunc(attributeName string) func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		conn := meta.(*conns.AWSClient).SQSConn

		log.Printf("[DEBUG] Deleting SQS Queue Attribute '%s': %s", attributeName, d.Id())

		var emptyAttributes map[string]string
		switch attributeName {
		case sqs.QueueAttributeNamePolicy:
			emptyAttributes = queueEmptyPolicyAttributes
		case sqs.QueueAttributeNameRedrivePolicy:
			emptyAttributes = queueEmptyRedrivePolicyAttributes
		default:
			return diag.FromErr(fmt.Errorf("%s is an invalid SQS Queue attribute name", attributeName))
		}

		_, err := conn.SetQueueAttributes(&sqs.SetQueueAttributesInput{
			Attributes: aws.StringMap(emptyAttributes),
			QueueUrl:   aws.String(d.Id()),
		})

		if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
			return nil
		}

		if err != nil {
			return diag.FromErr(fmt.Errorf("error deleting SQS Queue Attribute '%s' (%s): %w", attributeName, d.Id(), err))
		}

		err = waitQueueAttributesPropagatedWithContext(ctx, conn, d.Id(), emptyAttributes)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error waiting for SQS Queue Attribute '%s' (%s) to delete: %w", attributeName, d.Id(), err))
		}

		return nil
	}
}
