// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_ecr_registry_policy")
func ResourceRegistryPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegistryPolicyPut,
		ReadWithoutTimeout:   resourceRegistryPolicyRead,
		UpdateWithoutTimeout: resourceRegistryPolicyPut,
		DeleteWithoutTimeout: resourceRegistryPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRegistryPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	input := ecr.PutRegistryPolicyInput{
		PolicyText: aws.String(policy),
	}

	out, err := conn.PutRegistryPolicy(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECR Registry Policy: %s", err)
	}

	regID := aws.ToString(out.RegistryId)

	d.SetId(regID)

	return append(diags, resourceRegistryPolicyRead(ctx, d, meta)...)
}

func resourceRegistryPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	log.Printf("[DEBUG] Reading registry policy %s", d.Id())
	out, err := conn.GetRegistryPolicy(ctx, &ecr.GetRegistryPolicyInput{})
	if err != nil {
		if !d.IsNewResource() && errs.IsA[*awstypes.RegistryPolicyNotFoundException](err) {
			log.Printf("[WARN] ECR Registry (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading ECR Registry Policy (%s): %s", d.Id(), err)
	}

	d.Set("registry_id", out.RegistryId)

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.ToString(out.PolicyText))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Registry Policy (%s): setting policy: %s", d.Id(), err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Registry Policy (%s): setting policy: %s", d.Id(), err)
	}

	d.Set("policy", policyToSet)

	return diags
}

func resourceRegistryPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	_, err := conn.DeleteRegistryPolicy(ctx, &ecr.DeleteRegistryPolicyInput{})
	if err != nil {
		if errs.IsA[*awstypes.RegistryPolicyNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting ECR Registry Policy (%s): %s", d.Id(), err)
	}

	return diags
}
