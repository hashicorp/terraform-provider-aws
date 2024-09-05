// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_secretsmanager_secret_policy", name="Secret Policy")
func resourceSecretPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecretPolicyCreate,
		ReadWithoutTimeout:   resourceSecretPolicyRead,
		UpdateWithoutTimeout: resourceSecretPolicyUpdate,
		DeleteWithoutTimeout: resourceSecretPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"block_public_policy": {
				Type:     schema.TypeBool,
				Optional: true,
			},
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
			"secret_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceSecretPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &secretsmanager.PutResourcePolicyInput{
		ResourcePolicy: aws.String(policy),
		SecretId:       aws.String(d.Get("secret_arn").(string)),
	}

	if v, ok := d.GetOk("block_public_policy"); ok {
		input.BlockPublicPolicy = aws.Bool(v.(bool))
	}

	output, err := putSecretPolicy(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(aws.ToString(output.ARN))

	_, err = tfresource.RetryWhenNotFound(ctx, PropagationTimeout, func() (interface{}, error) {
		return findSecretPolicyByID(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Secrets Manager Secret Policy (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceSecretPolicyRead(ctx, d, meta)...)
}

func resourceSecretPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	output, err := findSecretPolicyByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Secrets Manager Secret Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Secrets Manager Secret Policy (%s): %s", d.Id(), err)
	}

	// Empty (nil or "") policy indicates that the policy has been deleted.
	// For backwards compatibility we don't check that.

	if output.ResourcePolicy != nil {
		policyToSet, err := verify.PolicyToSet(d.Get(names.AttrPolicy).(string), aws.ToString(output.ResourcePolicy))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set(names.AttrPolicy, policyToSet)
	} else {
		d.Set(names.AttrPolicy, "")
	}
	d.Set("secret_arn", d.Id())

	return diags
}

func resourceSecretPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &secretsmanager.PutResourcePolicyInput{
		ResourcePolicy:    aws.String(policy),
		SecretId:          aws.String(d.Id()),
		BlockPublicPolicy: aws.Bool(d.Get("block_public_policy").(bool)),
	}

	if _, err := putSecretPolicy(ctx, conn, input); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceSecretPolicyRead(ctx, d, meta)...)
}

func resourceSecretPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecretsManagerClient(ctx)

	log.Printf("[DEBUG] Deleting Secrets Manager Secret Policy: %s", d.Id())
	err := deleteSecretPolicy(ctx, conn, d.Id())

	if errs.IsA[*types.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "You can't perform this operation on the secret because it was marked for deletion") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, PropagationTimeout, func() (interface{}, error) {
		output, err := findSecretPolicyByID(ctx, conn, d.Id())

		if err != nil {
			return nil, err
		}

		if aws.ToString(output.ResourcePolicy) == "" {
			return nil, &retry.NotFoundError{}
		}

		return output, nil
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Secrets Manager Secret Policy (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findSecretPolicyByID(ctx context.Context, conn *secretsmanager.Client, id string) (*secretsmanager.GetResourcePolicyOutput, error) {
	input := &secretsmanager.GetResourcePolicyInput{
		SecretId: aws.String(id),
	}

	output, err := conn.GetResourcePolicy(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*types.InvalidRequestException](err, "You can't perform this operation on the secret because it was marked for deletion") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
