// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_resource_policy")
func ResourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourcePolicyPut(awstypes.ExistConditionNotExist),
		ReadWithoutTimeout:   resourceResourcePolicyRead,
		UpdateWithoutTimeout: resourceResourcePolicyPut(awstypes.ExistConditionMustExist),
		DeleteWithoutTimeout: resourceResourcePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
			"enable_hybrid": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.EnableHybridValues](),
			},
		},
	}
}

func resourceResourcePolicyPut(condition awstypes.ExistCondition) func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		var diags diag.Diagnostics
		conn := meta.(*conns.AWSClient).GlueClient(ctx)

		policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "policy is invalid JSON: %s", err)
		}

		input := &glue.PutResourcePolicyInput{
			PolicyInJson:          aws.String(policy),
			PolicyExistsCondition: condition,
		}

		if v, ok := d.GetOk("enable_hybrid"); ok {
			input.EnableHybrid = awstypes.EnableHybridValues(v.(string))
		}

		_, err = conn.PutResourcePolicy(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting policy request: %s", err)
		}
		d.SetId(meta.(*conns.AWSClient).Region)

		return append(diags, resourceResourcePolicyRead(ctx, d, meta)...)
	}
}

func resourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	resourcePolicy, err := conn.GetResourcePolicy(ctx, &glue.GetResourcePolicyInput{})
	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		log.Printf("[WARN] Glue Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Resource Policy (%s): %s", d.Id(), err)
	}

	if aws.ToString(resourcePolicy.PolicyInJson) == "" {
		//Since the glue resource policy is global we expect it to be deleted when the policy is empty
		d.SetId("")
	} else {
		policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(resourcePolicy.PolicyInJson))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Glue Resource Policy (%s): %s", d.Id(), err)
		}

		d.Set(names.AttrPolicy, policyToSet)
	}
	return diags
}

func resourceResourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	_, err := conn.DeleteResourcePolicy(ctx, &glue.DeleteResourcePolicyInput{})
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting policy request: %s", err)
	}

	return diags
}
