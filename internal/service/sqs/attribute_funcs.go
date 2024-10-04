// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type queueAttributeHandler struct {
	AttributeName types.QueueAttributeName
	SchemaKey     string
	ToSet         func(string, string) (string, error)
}

func (h *queueAttributeHandler) Upsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SQSClient(ctx)

	attrValue, err := structure.NormalizeJsonString(d.Get(h.SchemaKey).(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "%s (%s) is invalid JSON: %s", h.SchemaKey, d.Get(h.SchemaKey).(string), err)
	}

	attributes := map[types.QueueAttributeName]string{
		h.AttributeName: attrValue,
	}
	url := d.Get("queue_url").(string)
	input := &sqs.SetQueueAttributesInput{
		Attributes: flex.ExpandStringyValueMap(attributes),
		QueueUrl:   aws.String(url),
	}

	_, err = tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.SetQueueAttributes(ctx, input)
	}, errCodeInvalidAttributeValue, "Invalid value for the parameter Policy")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting SQS Queue (%s) attribute (%s): %s", url, h.AttributeName, err)
	}

	d.SetId(url)

	if err := waitQueueAttributesPropagated(ctx, conn, d.Id(), attributes); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SQS Queue (%s) attribute (%s) create: %s", d.Id(), h.AttributeName, err)
	}

	return append(diags, h.Read(ctx, d, meta)...)
}

func (h *queueAttributeHandler) Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SQSClient(ctx)

	outputRaw, err := tfresource.RetryWhenNotFound(ctx, queueAttributeReadTimeout, func() (interface{}, error) {
		return findQueueAttributeByTwoPartKey(ctx, conn, d.Id(), h.AttributeName)
	})

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SQS Queue (%s) attribute (%s) not found, removing from state", d.Id(), h.AttributeName)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SQS Queue (%s) attribute (%s): %s", d.Id(), h.AttributeName, err)
	}

	newValue, err := h.ToSet(d.Get(h.SchemaKey).(string), aws.ToString(outputRaw.(*string)))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if h.SchemaKey == names.AttrPolicy {
		newValue, err = verify.PolicyToSet(d.Get(names.AttrPolicy).(string), newValue)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	d.Set(h.SchemaKey, newValue)
	d.Set("queue_url", d.Id())

	return diags
}

func (h *queueAttributeHandler) Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SQSClient(ctx)

	log.Printf("[DEBUG] Deleting SQS Queue (%s) attribute: %s", d.Id(), h.AttributeName)
	attributes := map[types.QueueAttributeName]string{
		h.AttributeName: "",
	}
	_, err := conn.SetQueueAttributes(ctx, &sqs.SetQueueAttributesInput{
		Attributes: flex.ExpandStringyValueMap(attributes),
		QueueUrl:   aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeQueueDoesNotExist) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SQS Queue (%s) attribute (%s): %s", d.Id(), h.AttributeName, err)
	}

	if err := waitQueueAttributesPropagated(ctx, conn, d.Id(), attributes); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SQS Queue (%s) attribute (%s) delete: %s", d.Id(), h.AttributeName, err)
	}

	return diags
}
