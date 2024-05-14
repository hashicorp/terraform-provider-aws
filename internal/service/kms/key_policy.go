// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_kms_key_policy", name="Key Policy")
func resourceKeyPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKeyPolicyCreate,
		ReadWithoutTimeout:   resourceKeyPolicyRead,
		UpdateWithoutTimeout: resourceKeyPolicyUpdate,
		DeleteWithoutTimeout: resourceKeyPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bypass_policy_lockout_safety_check": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrKeyID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 2048),
			},
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Required:              true,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				ValidateFunc:          validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceKeyPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	keyID := d.Get(names.AttrKeyID).(string)

	if err := updateKeyPolicy(ctx, conn, "KMS Key Policy", keyID, d.Get(names.AttrPolicy).(string), d.Get("bypass_policy_lockout_safety_check").(bool)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(keyID)

	return append(diags, resourceKeyPolicyRead(ctx, d, meta)...)
}

func resourceKeyPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	key, err := findKeyInfo(ctx, conn, d.Id(), d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Key (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrKeyID, key.metadata.KeyId)

	policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), key.policy)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)

	return diags
}

func resourceKeyPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	if d.HasChange(names.AttrPolicy) {
		if err := updateKeyPolicy(ctx, conn, "KMS Key Policy", d.Id(), d.Get(names.AttrPolicy).(string), d.Get("bypass_policy_lockout_safety_check").(bool)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceKeyPolicyRead(ctx, d, meta)...)
}

func resourceKeyPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSClient(ctx)

	if !d.Get("bypass_policy_lockout_safety_check").(bool) {
		if err := updateKeyPolicy(ctx, conn, "KMS Key Policy", d.Get(names.AttrKeyID).(string), meta.(*conns.AWSClient).DefaultKMSKeyPolicy(ctx), d.Get("bypass_policy_lockout_safety_check").(bool)); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		} else {
			log.Printf("[WARN] KMS Key Policy for Key (%s) does not allow PutKeyPolicy. Default Policy cannot be restored. Removing from state", d.Id())
		}
	}

	return diags
}
