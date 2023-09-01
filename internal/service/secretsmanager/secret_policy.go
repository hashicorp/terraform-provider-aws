// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
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

// @SDKResource("aws_secretsmanager_secret_policy")
func ResourceSecretPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecretPolicyCreate,
		ReadWithoutTimeout:   resourceSecretPolicyRead,
		UpdateWithoutTimeout: resourceSecretPolicyUpdate,
		DeleteWithoutTimeout: resourceSecretPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"secret_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
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
			"block_public_policy": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceSecretPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerConn(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", d.Get("policy").(string), err)
	}

	input := &secretsmanager.PutResourcePolicyInput{
		ResourcePolicy: aws.String(policy),
		SecretId:       aws.String(d.Get("secret_arn").(string)),
	}

	if v, ok := d.GetOk("block_public_policy"); ok {
		input.BlockPublicPolicy = aws.Bool(v.(bool))
	}

	log.Printf("[DEBUG] Setting Secrets Manager Secret resource policy; %#v", input)
	var output *secretsmanager.PutResourcePolicyOutput

	err = retry.RetryContext(ctx, PropagationTimeout, func() *retry.RetryError {
		var err error
		output, err = conn.PutResourcePolicyWithContext(ctx, input)
		if tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeMalformedPolicyDocumentException,
			"This resource policy contains an unsupported principal") {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		output, err = conn.PutResourcePolicyWithContext(ctx, input)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Secrets Manager Secret %q policy: %s", d.Id(), err)
	}

	d.SetId(aws.StringValue(output.ARN))

	return append(diags, resourceSecretPolicyRead(ctx, d, meta)...)
}

func resourceSecretPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerConn(ctx)

	input := &secretsmanager.GetResourcePolicyInput{
		SecretId: aws.String(d.Id()),
	}

	outputRaw, err := tfresource.RetryWhenNotFound(ctx, PropagationTimeout, func() (interface{}, error) {
		return conn.GetResourcePolicyWithContext(ctx, input)
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, secretsmanager.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Secrets Manager Secret Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret Policy (%s): %s", d.Id(), err)
	}

	output := outputRaw.(*secretsmanager.GetResourcePolicyOutput)

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret Policy (%s): empty response", d.Id())
	}

	if output.ResourcePolicy != nil {
		policyToSet, err := verify.PolicyToSet(d.Get("policy").(string), aws.StringValue(output.ResourcePolicy))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret Policy (%s): %s", d.Id(), err)
		}

		d.Set("policy", policyToSet)
	} else {
		d.Set("policy", "")
	}
	d.Set("secret_arn", d.Id())

	return diags
}

func resourceSecretPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerConn(ctx)

	if d.HasChanges("policy", "block_public_policy") {
		policy, err := structure.NormalizeJsonString(d.Get("policy").(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "policy contains an invalid JSON: %s", err)
		}
		input := &secretsmanager.PutResourcePolicyInput{
			ResourcePolicy:    aws.String(policy),
			SecretId:          aws.String(d.Id()),
			BlockPublicPolicy: aws.Bool(d.Get("block_public_policy").(bool)),
		}

		log.Printf("[DEBUG] Setting Secrets Manager Secret resource policy; %#v", input)
		err = retry.RetryContext(ctx, PropagationTimeout, func() *retry.RetryError {
			_, err := conn.PutResourcePolicyWithContext(ctx, input)
			if tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeMalformedPolicyDocumentException,
				"This resource policy contains an unsupported principal") {
				return retry.RetryableError(err)
			}
			if err != nil {
				return retry.NonRetryableError(err)
			}
			return nil
		})
		if tfresource.TimedOut(err) {
			_, err = conn.PutResourcePolicyWithContext(ctx, input)
		}
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting Secrets Manager Secret %q policy: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSecretPolicyRead(ctx, d, meta)...)
}

func resourceSecretPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerConn(ctx)

	input := &secretsmanager.DeleteResourcePolicyInput{
		SecretId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Removing Secrets Manager Secret policy: %#v", input)
	_, err := conn.DeleteResourcePolicyWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, secretsmanager.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "removing Secrets Manager Secret %q policy: %s", d.Id(), err)
	}

	return diags
}
