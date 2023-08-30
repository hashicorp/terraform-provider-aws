// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_glue_resource_policy")
func ResourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourcePolicyPut(glue.ExistConditionNotExist),
		ReadWithoutTimeout:   resourceResourcePolicyRead,
		UpdateWithoutTimeout: resourceResourcePolicyPut(glue.ExistConditionMustExist),
		DeleteWithoutTimeout: resourceResourcePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
			"enable_hybrid": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(glue.EnableHybridValues_Values(), false),
			},
		},
	}
}

func resourceResourcePolicyPut(condition string) func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		var diags diag.Diagnostics
		conn := meta.(*conns.AWSClient).GlueConn(ctx)

		policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "policy is invalid JSON: %s", err)
		}

		input := &glue.PutResourcePolicyInput{
			PolicyInJson:          aws.String(policy),
			PolicyExistsCondition: aws.String(condition),
		}

		if v, ok := d.GetOk("enable_hybrid"); ok {
			input.EnableHybrid = aws.String(v.(string))
		}

		_, err = conn.PutResourcePolicyWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting policy request: %s", err)
		}
		d.SetId(meta.(*conns.AWSClient).Region)

		return append(diags, resourceResourcePolicyRead(ctx, d, meta)...)
	}
}

func resourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

	resourcePolicy, err := conn.GetResourcePolicyWithContext(ctx, &glue.GetResourcePolicyInput{})
	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		log.Printf("[WARN] Glue Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Resource Policy (%s): %s", d.Id(), err)
	}

	if aws.StringValue(resourcePolicy.PolicyInJson) == "" {
		//Since the glue resource policy is global we expect it to be deleted when the policy is empty
		d.SetId("")
	} else {
		policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.StringValue(resourcePolicy.PolicyInJson))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Glue Resource Policy (%s): %s", d.Id(), err)
		}

		d.Set("policy", policyToSet)
	}
	return diags
}

func resourceResourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn(ctx)

	_, err := conn.DeleteResourcePolicyWithContext(ctx, &glue.DeleteResourcePolicyInput{})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting policy request: %s", err)
	}

	return diags
}
