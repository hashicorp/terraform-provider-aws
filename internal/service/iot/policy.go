// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_version_id": {
				Type:     schema.TypeString,
				Computed: true,
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

	out, err := conn.CreatePolicyWithContext(ctx, &iot.CreatePolicyInput{
		PolicyName:     aws.String(d.Get("name").(string)),
		PolicyDocument: aws.String(policy),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IoT Policy: %s", err)
	}

	d.SetId(aws.StringValue(out.PolicyName))

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

	if d.HasChange("policy") {
		policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
		}

		_, err = conn.CreatePolicyVersionWithContext(ctx, &iot.CreatePolicyVersionInput{
			PolicyName:     aws.String(d.Id()),
			PolicyDocument: aws.String(policy),
			SetAsDefault:   aws.Bool(true),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IoT Policy (%s): %s", d.Id(), err)
		}
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
			_, err = conn.DeletePolicyVersionWithContext(ctx, &iot.DeletePolicyVersionInput{
				PolicyName:      aws.String(d.Id()),
				PolicyVersionId: ver.VersionId,
			})

			if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting IoT Policy (%s) version (%s): %s", d.Id(), aws.StringValue(ver.VersionId), err)
			}
		}
	}

	//Delete default policy version
	_, err = conn.DeletePolicyWithContext(ctx, &iot.DeletePolicyInput{
		PolicyName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, iot.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IoT Policy (%s): %s", d.Id(), err)
	}

	return diags
}
