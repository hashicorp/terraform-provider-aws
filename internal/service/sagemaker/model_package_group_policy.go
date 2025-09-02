// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
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

// @SDKResource("aws_sagemaker_model_package_group_policy", name="Model Package Group Policy")
func resourceModelPackageGroupPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceModelPackageGroupPolicyPut,
		ReadWithoutTimeout:   resourceModelPackageGroupPolicyRead,
		UpdateWithoutTimeout: resourceModelPackageGroupPolicyPut,
		DeleteWithoutTimeout: resourceModelPackageGroupPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"model_package_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_policy": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceModelPackageGroupPolicyPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("resource_policy").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", d.Get("resource_policy").(string), err)
	}

	name := d.Get("model_package_group_name").(string)
	input := &sagemaker.PutModelPackageGroupPolicyInput{
		ModelPackageGroupName: aws.String(name),
		ResourcePolicy:        aws.String(policy),
	}

	_, err = conn.PutModelPackageGroupPolicy(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Model Package Group Policy %s: %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceModelPackageGroupPolicyRead(ctx, d, meta)...)
}

func resourceModelPackageGroupPolicyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	mpg, err := findModelPackageGroupPolicyByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Unable to find SageMaker AI Model Package Group Policy (%s); removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Model Package Group Policy (%s): %s", d.Id(), err)
	}

	d.Set("model_package_group_name", d.Id())

	policyToSet, err := verify.PolicyToSet(d.Get("resource_policy").(string), aws.ToString(mpg.ResourcePolicy))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Model Package Group Policy (%s): %s", d.Id(), err)
	}

	d.Set("resource_policy", policyToSet)

	return diags
}

func resourceModelPackageGroupPolicyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	input := &sagemaker.DeleteModelPackageGroupPolicyInput{
		ModelPackageGroupName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteModelPackageGroupPolicy(ctx, input); err != nil {
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Cannot find Model Package Group") ||
			tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Cannot find resource policy") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Model Package Group Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findModelPackageGroupPolicyByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.GetModelPackageGroupPolicyOutput, error) {
	input := &sagemaker.GetModelPackageGroupPolicyInput{
		ModelPackageGroupName: aws.String(name),
	}

	output, err := conn.GetModelPackageGroupPolicy(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Cannot find Model Package Group") ||
		tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Cannot find resource policy") {
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
