package sqs

import (
	"context"
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
	ToSet         func(string, string) (string, error)
}

func (h *queueAttributeHandler) Upsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SQSConn()

	attrValue, err := structure.NormalizeJsonString(d.Get(h.SchemaKey).(string))
	if err != nil {
		return diag.Errorf("%s (%s) is invalid JSON: %s", h.SchemaKey, d.Get(h.SchemaKey).(string), err)
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

	if err := waitQueueAttributesPropagated(ctx, conn, d.Id(), attributes); err != nil {
		return diag.Errorf("waiting for SQS Queue (%s) attribute (%s) create: %s", d.Id(), h.AttributeName, err)
	}

	return h.Read(ctx, d, meta)
}

func (h *queueAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SQSConn()

	outputRaw, err := tfresource.RetryWhenNotFound(ctx, queueAttributeReadTimeout, func() (interface{}, error) {
		return FindQueueAttributeByURL(ctx, conn, d.Id(), h.AttributeName)
	})

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SQS Queue (%s) attribute (%s) not found, removing from state", d.Id(), h.AttributeName)
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading SQS Queue (%s) attribute (%s): %s", d.Id(), h.AttributeName, err)
	}

	newValue, err := h.ToSet(d.Get(h.SchemaKey).(string), outputRaw.(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if h.SchemaKey == "policy" {
		newValue, err = verify.PolicyToSet(d.Get("policy").(string), newValue)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.Set(h.SchemaKey, newValue)
	d.Set("queue_url", d.Id())

	return nil
}

func (h *queueAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SQSConn()

	log.Printf("[DEBUG] Deleting SQS Queue (%s) attribute: %s", d.Id(), h.AttributeName)
	attributes := map[string]string{
		h.AttributeName: "",
	}
	_, err := conn.SetQueueAttributesWithContext(ctx, &sqs.SetQueueAttributesInput{
		Attributes: aws.StringMap(attributes),
		QueueUrl:   aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting SQS Queue (%s) attribute (%s): %s", d.Id(), h.AttributeName, err)
	}

	if err := waitQueueAttributesPropagated(ctx, conn, d.Id(), attributes); err != nil {
		return diag.Errorf("waiting for SQS Queue (%s) attribute (%s) delete: %s", d.Id(), h.AttributeName, err)
	}

	return nil
}
