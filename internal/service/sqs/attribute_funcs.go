package sqs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

type queueAttributeHandler struct {
	AttributeName string
	SchemaKey     string
}

func (h *queueAttributeHandler) Create() schema.CreateContextFunc {
	return h.upsert
}

func (h *queueAttributeHandler) Read() schema.ReadContextFunc {
	return h.read
}

func (h *queueAttributeHandler) Update() schema.UpdateContextFunc {
	return h.upsert
}

func (h *queueAttributeHandler) Delete() schema.DeleteContextFunc {
	return h.delete
}

func (h *queueAttributeHandler) upsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SQSConn

	attrValue, err := structure.NormalizeJsonString(d.Get(h.SchemaKey).(string))

	if err != nil {
		return diag.FromErr(fmt.Errorf("%s (%s) is invalid JSON: %w", h.SchemaKey, d.Get(h.SchemaKey).(string), err))
	}

	attributes := map[string]string{
		h.AttributeName: attrValue,
	}
	url := d.Get("queue_url").(string)
	input := &sqs.SetQueueAttributesInput{
		Attributes: aws.StringMap(attributes),
		QueueUrl:   aws.String(url),
	}

	log.Printf("[DEBUG] Setting SQS Queue attributes: %s", input)
	_, err = conn.SetQueueAttributesWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("setting SQS Queue (%s) attribute (%s): %s", url, h.AttributeName, err)
	}

	d.SetId(url)

	if err := waitQueueAttributesPropagatedWithContext(ctx, conn, d.Id(), attributes); err != nil {
		return diag.Errorf("waiting for SQS Queue (%s) attribute (%s) create: %s", d.Id(), h.AttributeName, err)
	}

	return h.read(ctx, d, meta)
}

func (h *queueAttributeHandler) read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func (h *queueAttributeHandler) delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SQSConn

	log.Printf("[DEBUG] Deleting SQS Queue (%s) attribute: %s", d.Id(), h.AttributeName)
	attributes := map[string]string{
		h.AttributeName: "",
	}
	_, err := conn.SetQueueAttributes(&sqs.SetQueueAttributesInput{
		Attributes: aws.StringMap(attributes),
		QueueUrl:   aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting SQS Queue (%s) attribute (%s): %s", d.Id(), h.AttributeName, err)
	}

	if err := waitQueueAttributesPropagatedWithContext(ctx, conn, d.Id(), attributes); err != nil {
		return diag.Errorf("waiting for SQS Queue (%s) attribute (%s) delete: %s", d.Id(), h.AttributeName, err)
	}

	return nil
}

func generateQueueAttributeUpsertFunc(attributeName string) func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		conn := meta.(*conns.AWSClient).SQSConn

		attrValue, err := structure.NormalizeJsonString(d.Get(getSchemaKey(attributeName)).(string))

		if err != nil {
			return diag.FromErr(fmt.Errorf("%s (%s) is invalid JSON: %w", attributeName, d.Get(getSchemaKey(attributeName)).(string), err))
		}

		attributes := map[string]string{
			attributeName: attrValue,
		}
		url := d.Get("queue_url").(string)
		input := &sqs.SetQueueAttributesInput{
			Attributes: aws.StringMap(attributes),
			QueueUrl:   aws.String(url),
		}

		log.Printf("[DEBUG] Setting SQS Queue attributes: %s", input)
		_, err = conn.SetQueueAttributesWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("setting SQS Queue (%s) attribute (%s): %s", url, attributeName, err)
		}

		d.SetId(url)

		if err := waitQueueAttributesPropagatedWithContext(ctx, conn, d.Id(), attributes); err != nil {
			return diag.Errorf("waiting for SQS Queue (%s) attribute (%s) create: %s", d.Id(), attributeName, err)
		}

		return generateQueueAttributeReadFunc(attributeName)(ctx, d, meta)
	}
}

func generateQueueAttributeReadFunc(attributeName string) func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		conn := meta.(*conns.AWSClient).SQSConn

		outputRaw, err := tfresource.RetryWhenNotFoundContext(ctx, queueAttributeReadTimeout, func() (interface{}, error) {
			return FindQueueAttributeByURL(ctx, conn, d.Id(), attributeName)
		})

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] SQS Queue (%s) attribute (%s) not found, removing from state", d.Id(), attributeName)
			d.SetId("")
			return nil
		}

		if err != nil {
			return diag.Errorf("reading SQS Queue (%s) attribute (%s): %s", d.Id(), attributeName, err)
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

		log.Printf("[DEBUG] Deleting SQS Queue (%s) attribute: %s", d.Id(), attributeName)
		attributes := map[string]string{
			attributeName: "",
		}
		_, err := conn.SetQueueAttributes(&sqs.SetQueueAttributesInput{
			Attributes: aws.StringMap(attributes),
			QueueUrl:   aws.String(d.Id()),
		})

		if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
			return nil
		}

		if err != nil {
			return diag.Errorf("deleting SQS Queue (%s) attribute (%s): %s", d.Id(), attributeName, err)
		}

		if err := waitQueueAttributesPropagatedWithContext(ctx, conn, d.Id(), attributes); err != nil {
			return diag.Errorf("waiting for SQS Queue (%s) attribute (%s) delete: %s", d.Id(), attributeName, err)
		}

		return nil
	}
}
