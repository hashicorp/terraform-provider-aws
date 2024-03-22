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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_ecr_repository_policy")
func ResourceRepositoryPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRepositoryPolicyPut,
		ReadWithoutTimeout:   resourceRepositoryPolicyRead,
		UpdateWithoutTimeout: resourceRepositoryPolicyPut,
		DeleteWithoutTimeout: resourceRepositoryPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"repository": {
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
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRepositoryPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
	}

	input := ecr.SetRepositoryPolicyInput{
		RepositoryName: aws.String(d.Get("repository").(string)),
		PolicyText:     aws.String(policy),
	}

	log.Printf("[DEBUG] Creating ECR repository policy: %#v", input)

	// Retry due to IAM eventual consistency
	var out *ecr.SetRepositoryPolicyOutput
	err = retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		out, err = conn.SetRepositoryPolicy(ctx, &input)

		if errs.IsAErrorMessageContains[*awstypes.InvalidParameterException](err, "Invalid repository policy provided") {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		out, err = conn.SetRepositoryPolicy(ctx, &input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ECR Repository Policy: %s", err)
	}

	log.Printf("[DEBUG] ECR repository policy created: %s", aws.ToString(out.RepositoryName))

	d.SetId(aws.ToString(out.RepositoryName))

	return append(diags, resourceRepositoryPolicyRead(ctx, d, meta)...)
}

func resourceRepositoryPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	input := &ecr.GetRepositoryPolicyInput{
		RepositoryName: aws.String(d.Id()),
	}

	var out *ecr.GetRepositoryPolicyOutput

	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error

		out, err = conn.GetRepositoryPolicy(ctx, input)

		if d.IsNewResource() && (errs.IsA[*awstypes.RepositoryNotFoundException](err) || errs.IsA[*awstypes.RepositoryPolicyNotFoundException](err)) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		out, err = conn.GetRepositoryPolicy(ctx, input)
	}

	if !d.IsNewResource() && errs.IsA[*awstypes.RepositoryNotFoundException](err) {
		log.Printf("[WARN] ECR Repository Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if !d.IsNewResource() && errs.IsA[*awstypes.RepositoryPolicyNotFoundException](err) {
		log.Printf("[WARN] ECR Repository Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Repository Policy (%s): %s", d.Id(), err)
	}

	if out == nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Repository Policy (%s): empty response", d.Id())
	}

	log.Printf("[DEBUG] Received repository policy %s", *out.PolicyText)

	d.Set("repository", out.RepositoryName)
	d.Set("registry_id", out.RegistryId)

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.ToString(out.PolicyText))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "while setting policy (%s), encountered: %s", policyToSet, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policyToSet, err)
	}

	d.Set("policy", policyToSet)

	return diags
}

func resourceRepositoryPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	_, err := conn.DeleteRepositoryPolicy(ctx, &ecr.DeleteRepositoryPolicyInput{
		RepositoryName: aws.String(d.Id()),
		RegistryId:     aws.String(d.Get("registry_id").(string)),
	})
	if err != nil {
		if errs.IsA[*awstypes.RepositoryNotFoundException](err) ||
			errs.IsA[*awstypes.RepositoryPolicyNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting ECR Repository Policy (%s): %s", d.Id(), err)
	}

	return diags
}
