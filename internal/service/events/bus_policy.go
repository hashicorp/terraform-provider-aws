// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

// @SDKResource("aws_cloudwatch_event_bus_policy", name="Event Bus Policy")
func resourceBusPolicy() *schema.Resource {
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

func resourceBusPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
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

	_, err = conn.PutPermission(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge Event Bus (%s) Policy: %s", eventBusName, err)
	}

	if d.IsNewResource() {
		d.SetId(eventBusName)
	}

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findEventBusPolicyByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EventBridge Event Bus (%s) Policy create: %s", d.Id(), err)
	}

	return append(diags, resourceBusPolicyRead(ctx, d, meta)...)
}

func resourceBusPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	policy, err := findEventBusPolicyByName(ctx, conn, d.Id())

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

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(policy))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	return diags
}

func resourceBusPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	log.Printf("[DEBUG] Deleting EventBridge Event Bus Policy: %s", d.Id())
	_, err := conn.RemovePermission(ctx, &eventbridge.RemovePermissionInput{
		EventBusName:         aws.String(d.Id()),
		RemoveAllPermissions: true,
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge Event Bus (%s) Policy: %s", d.Id(), err)
	}

	return diags
}

func findEventBusPolicyByName(ctx context.Context, conn *eventbridge.Client, name string) (*string, error) {
	output, err := findEventBusByName(ctx, conn, name)

	if err != nil {
		return nil, err
	}

	if aws.ToString(output.Policy) == "" {
		return nil, &retry.NotFoundError{}
	}

	return output.Policy, nil
}
