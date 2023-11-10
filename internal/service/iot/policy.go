// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
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

const (
	// Maximum amount of time for a just-removed policy attachment to propagate.
	policyAttachmentTimeout = 5 * time.Minute
)

// @SDKResource("aws_iot_policy")
func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyCreate,
		ReadWithoutTimeout:   resourcePolicyRead,
		UpdateWithoutTimeout: resourcePolicyUpdate,
		DeleteWithoutTimeout: resourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	name := d.Get("name").(string)
	input := &iot.CreatePolicyInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(name),
	}

	output, err := conn.CreatePolicyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Policy (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.PolicyName))

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	out, err := conn.GetPolicyWithContext(ctx, &iot.GetPolicyInput{
		PolicyName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] IoT Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Policy (%s): %s", d.Id(), err)
	}

	d.Set("arn", out.PolicyArn)
	d.Set("default_version_id", out.DefaultVersionId)
	d.Set("name", out.PolicyName)

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.StringValue(out.PolicyDocument))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IoT Policy (%s): %s", d.Id(), err)
	}

	d.Set("policy", policyToSet)

	return diags
}

func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	input := &iot.CreatePolicyVersionInput{
		PolicyDocument: aws.String(policy),
		PolicyName:     aws.String(d.Id()),
		SetAsDefault:   aws.Bool(true),
	}

	_, err = conn.CreatePolicyVersionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating IoT Policy (%s): %s", d.Id(), err)
	}

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IoTConn(ctx)

	out, err := conn.ListPolicyVersionsWithContext(ctx, &iot.ListPolicyVersionsInput{
		PolicyName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing IoT Policy (%s) versions: %s", d.Id(), err)
	}

	// Delete all non-default versions of the policy
	for _, ver := range out.PolicyVersions {
		if !aws.BoolValue(ver.IsDefaultVersion) {
			err = retry.RetryContext(ctx, policyAttachmentTimeout, func() *retry.RetryError {
				_, err := conn.DeletePolicyVersionWithContext(ctx, &iot.DeletePolicyVersionInput{
					PolicyName:      aws.String(d.Id()),
					PolicyVersionId: ver.VersionId,
				})

				if tfawserr.ErrCodeEquals(err, iot.ErrCodeDeleteConflictException) {
					return retry.RetryableError(err)
				}

				if err != nil {
					return retry.NonRetryableError(err)
				}

				return nil
			})

			if tfresource.TimedOut(err) {
				_, err = conn.DeletePolicyVersionWithContext(ctx, &iot.DeletePolicyVersionInput{
					PolicyName:      aws.String(d.Id()),
					PolicyVersionId: ver.VersionId,
				})
			}

			if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting IoT Policy (%s) version (%s): %s", d.Id(), aws.StringValue(ver.VersionId), err)
			}
		}
	}

	//Delete default policy version
	err = retry.RetryContext(ctx, policyAttachmentTimeout, func() *retry.RetryError {
		_, err := conn.DeletePolicyWithContext(ctx, &iot.DeletePolicyInput{
			PolicyName: aws.String(d.Id()),
		})

		if tfawserr.ErrCodeEquals(err, iot.ErrCodeDeleteConflictException) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeletePolicyWithContext(ctx, &iot.DeletePolicyInput{
			PolicyName: aws.String(d.Id()),
		})
	}

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Policy (%s): %s", d.Id(), err)
	}

	return diags
}
