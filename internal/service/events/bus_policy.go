// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_cloudwatch_event_bus_policy")
func ResourceBusPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBusPolicyPut,
		ReadWithoutTimeout:   resourceBusPolicyRead,
		UpdateWithoutTimeout: resourceBusPolicyPut,
		DeleteWithoutTimeout: resourceBusPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("event_bus_name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"event_bus_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validBusName,
				Default:      DefaultEventBusName,
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

func resourceBusPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	var eventBusName string
	if d.IsNewResource() {
		eventBusName = d.Get("event_bus_name").(string)
	} else {
		eventBusName = d.Id()
	}
	input := &eventbridge.PutPermissionInput{
		EventBusName: aws.String(eventBusName),
		Policy:       aws.String(policy),
	}

	_, err = conn.PutPermissionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Event Bus (%s) Policy: %s", eventBusName, err)
	}

	if d.IsNewResource() {
		d.SetId(eventBusName)
	}

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindEventBusPolicyByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "wait for EventBridge Event Bus (%s) Policy create: %s", d.Id(), err)
	}

	return append(diags, resourceBusPolicyRead(ctx, d, meta)...)
}

func resourceBusPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	policy, err := FindEventBusPolicyByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge Event Bus (%s) Policy not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Event Bus (%s) Policy: %s", d.Id(), err)
	}

	eventBusName := d.Id()
	if eventBusName == "" {
		eventBusName = DefaultEventBusName
	}
	d.Set("event_bus_name", eventBusName)

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.StringValue(policy))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("policy", policyToSet)

	return diags
}

func resourceBusPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	log.Printf("[DEBUG] Deleting EventBridge Event Bus Policy: %s", d.Id())
	_, err := conn.RemovePermissionWithContext(ctx, &eventbridge.RemovePermissionInput{
		EventBusName:         aws.String(d.Id()),
		RemoveAllPermissions: aws.Bool(true),
	})

	if tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Event Bus (%s) Policy: %s", d.Id(), err)
	}

	return diags
}

func FindEventBusPolicyByName(ctx context.Context, conn *eventbridge.EventBridge, name string) (*string, error) {
	output, err := FindEventBusByName(ctx, conn, name)

	if err != nil {
		return nil, err
	}

	if aws.StringValue(output.Policy) == "" {
		return nil, &retry.NotFoundError{}
	}

	return output.Policy, nil
}
