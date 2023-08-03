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
)

// @SDKResource("aws_kms_key_policy")
func ResourceKeyPolicy() *schema.Resource {
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
			"key_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 2048),
			},
			"policy": {
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
	conn := meta.(*conns.AWSClient).KMSConn(ctx)

	keyID := d.Get("key_id").(string)

	if err := updateKeyPolicy(ctx, conn, keyID, d.Get("policy").(string), d.Get("bypass_policy_lockout_safety_check").(bool)); err != nil {
		return sdkdiag.AppendErrorf(diags, "attaching KMS Key policy (%s): %s", keyID, err)
	}

	d.SetId(keyID)

	return append(diags, resourceKeyPolicyRead(ctx, d, meta)...)
}

func resourceKeyPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn(ctx)

	key, err := findKey(ctx, conn, d.Id(), d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading KMS Key (%s): %s", d.Id(), err)
	}

	d.Set("key_id", key.metadata.KeyId)

	policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), key.policy)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while setting policy (%s), encountered: %s", key.policy, err)
	}

	d.Set("policy", policyToSet)

	return diags
}

func resourceKeyPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn(ctx)

	if d.HasChange("policy") {
		if err := updateKeyPolicy(ctx, conn, d.Id(), d.Get("policy").(string), d.Get("bypass_policy_lockout_safety_check").(bool)); err != nil {
			return sdkdiag.AppendErrorf(diags, "attaching KMS Key policy (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceKeyPolicyRead(ctx, d, meta)...)
}

func resourceKeyPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KMSConn(ctx)

	if !d.Get("bypass_policy_lockout_safety_check").(bool) {
		if err := updateKeyPolicy(ctx, conn, d.Get("key_id").(string), meta.(*conns.AWSClient).DefaultKMSKeyPolicy(), d.Get("bypass_policy_lockout_safety_check").(bool)); err != nil {
			return sdkdiag.AppendErrorf(diags, "attaching KMS Key policy (%s): %s", d.Id(), err)
		} else {
			log.Printf("[WARN] KMS Key Policy for Key (%s) does not allow PutKeyPolicy. Default Policy cannot be restored. Removing from state", d.Id())
		}
	}

	return diags
}
