// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
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

// @SDKResource("aws_sagemaker_model_package_group_policy")
func ResourceModelPackageGroupPolicy() *schema.Resource {
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
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceModelPackageGroupPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	policy, err := structure.NormalizeJsonString(d.Get("resource_policy").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", d.Get("resource_policy").(string), err)
	}

	name := d.Get("model_package_group_name").(string)
	input := &sagemaker.PutModelPackageGroupPolicyInput{
		ModelPackageGroupName: aws.String(name),
		ResourcePolicy:        aws.String(policy),
	}

	_, err = conn.PutModelPackageGroupPolicyWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Model Package Group Policy %s: %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceModelPackageGroupPolicyRead(ctx, d, meta)...)
}

func resourceModelPackageGroupPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	mpg, err := FindModelPackageGroupPolicyByName(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Unable to find SageMaker Model Package Group Policy (%s); removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Model Package Group Policy (%s): %s", d.Id(), err)
	}

	d.Set("model_package_group_name", d.Id())

	policyToSet, err := verify.PolicyToSet(d.Get("resource_policy").(string), aws.StringValue(mpg.ResourcePolicy))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Model Package Group Policy (%s): %s", d.Id(), err)
	}

	d.Set("resource_policy", policyToSet)

	return diags
}

func resourceModelPackageGroupPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	input := &sagemaker.DeleteModelPackageGroupPolicyInput{
		ModelPackageGroupName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteModelPackageGroupPolicyWithContext(ctx, input); err != nil {
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Cannot find Model Package Group") ||
			tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Cannot find resource policy") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Model Package Group Policy (%s): %s", d.Id(), err)
	}

	return diags
}
