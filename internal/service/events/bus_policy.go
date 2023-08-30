// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"fmt"
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
		CreateWithoutTimeout: resourceBusPolicyCreate,
		ReadWithoutTimeout:   resourceBusPolicyRead,
		UpdateWithoutTimeout: resourceBusPolicyUpdate,
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

func resourceBusPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	eventBusName := d.Get("event_bus_name").(string)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	input := eventbridge.PutPermissionInput{
		EventBusName: aws.String(eventBusName),
		Policy:       aws.String(policy),
	}

	log.Printf("[DEBUG] Creating EventBridge policy: %s", input)
	_, err = conn.PutPermissionWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Creating EventBridge policy failed: %s", err)
	}

	d.SetId(eventBusName)

	return append(diags, resourceBusPolicyRead(ctx, d, meta)...)
}

// See also: https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_DescribeEventBus.html
func resourceBusPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	eventBusName := d.Id()

	input := eventbridge.DescribeEventBusInput{
		Name: aws.String(eventBusName),
	}
	var output *eventbridge.DescribeEventBusOutput
	var err error
	var policy *string

	// Especially with concurrent PutPermission calls there can be a slight delay
	err = retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		log.Printf("[DEBUG] Reading EventBridge bus: %s", input)
		output, err = conn.DescribeEventBusWithContext(ctx, &input)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("reading EventBridge permission (%s) failed: %w", d.Id(), err))
		}

		policy, err = getEventBusPolicy(output)
		if err != nil {
			return retry.RetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.DescribeEventBusWithContext(ctx, &input)
		if output != nil {
			policy, err = getEventBusPolicy(output)
		}
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Policy on EventBridge Bus (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading policy from EventBridge Bus (%s): %s", d.Id(), err)
	}

	busName := aws.StringValue(output.Name)
	if busName == "" {
		busName = DefaultEventBusName
	}
	d.Set("event_bus_name", busName)

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.StringValue(policy))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading policy from EventBridge Bus (%s): %s", d.Id(), err)
	}

	d.Set("policy", policyToSet)

	return diags
}

func getEventBusPolicy(output *eventbridge.DescribeEventBusOutput) (*string, error) {
	if output == nil || output.Policy == nil {
		return nil, &retry.NotFoundError{
			Message:      fmt.Sprintf("Policy for EventBridge Bus (%s) not found", *output.Name),
			LastResponse: output,
		}
	}

	return output.Policy, nil
}

func resourceBusPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	eventBusName := d.Id()

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	input := eventbridge.PutPermissionInput{
		EventBusName: aws.String(eventBusName),
		Policy:       aws.String(policy),
	}

	log.Printf("[DEBUG] Update EventBridge Bus policy: %s", input)
	_, err = conn.PutPermissionWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating policy for EventBridge Bus (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBusPolicyRead(ctx, d, meta)...)
}

func resourceBusPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	eventBusName := d.Id()
	removeAllPermissions := true

	input := eventbridge.RemovePermissionInput{
		EventBusName:         aws.String(eventBusName),
		RemoveAllPermissions: &removeAllPermissions,
	}

	log.Printf("[DEBUG] Delete EventBridge Bus Policy: %s", input)
	_, err := conn.RemovePermissionWithContext(ctx, &input)
	if tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting policy for EventBridge Bus (%s): %s", d.Id(), err)
	}
	return diags
}
