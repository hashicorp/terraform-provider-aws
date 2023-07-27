// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_acmpca_policy")
func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyPut,
		ReadWithoutTimeout:   resourcePolicyRead,
		UpdateWithoutTimeout: resourcePolicyPut,
		DeleteWithoutTimeout: resourcePolicyDelete,

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
			"resource_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourcePolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAConn(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", d.Get("policy").(string), err)
	}

	resourceARN := d.Get("resource_arn").(string)
	input := &acmpca.PutPolicyInput{
		Policy:      aws.String(policy),
		ResourceArn: aws.String(resourceARN),
	}

	log.Printf("[DEBUG] Putting ACM PCA Policy: %s", input)
	_, err = conn.PutPolicyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting ACM PCA Policy (%s): %s", resourceARN, err)
	}

	d.SetId(resourceARN)

	return append(diags, resourcePolicyRead(ctx, d, meta)...)
}

func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAConn(ctx)

	policy, err := FindPolicyByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ACM PCA Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ACM PCA Policy (%s): %s", d.Id(), err)
	}

	d.Set("policy", policy)
	d.Set("resource_arn", d.Id())

	return diags
}

func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ACMPCAConn(ctx)

	log.Printf("[DEBUG] Deleting ACM PCA Policy: %s", d.Id())
	_, err := conn.DeletePolicyWithContext(ctx, &acmpca.DeletePolicyInput{
		ResourceArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) ||
		tfawserr.ErrCodeEquals(err, acmpca.ErrCodeRequestAlreadyProcessedException) ||
		tfawserr.ErrCodeEquals(err, acmpca.ErrCodeRequestInProgressException) ||
		tfawserr.ErrMessageContains(err, acmpca.ErrCodeInvalidRequestException, "Self-signed policy can not be revoked") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ACM PCA Policy (%s): %s", d.Id(), err)
	}

	return diags
}
