// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
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

// @SDKResource("aws_ecr_repository_policy", name="Repsitory Policy")
func resourceRepositoryPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRepositoryPolicyPut,
		ReadWithoutTimeout:   resourceRepositoryPolicyRead,
		UpdateWithoutTimeout: resourceRepositoryPolicyPut,
		DeleteWithoutTimeout: resourceRepositoryPolicyDelete,

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
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRepositoryPolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	policy, err := structure.NormalizeJsonString(d.Get(names.AttrPolicy).(string))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	repositoryName := d.Get("repository").(string)
	input := &ecr.SetRepositoryPolicyInput{
		PolicyText:     aws.String(policy),
		RepositoryName: aws.String(repositoryName),
	}

	_, err = tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidParameterException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.SetRepositoryPolicy(ctx, input)
	}, "Invalid repository policy provided")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting ECR Repository Policy (%s): %s", repositoryName, err)
	}

	if d.IsNewResource() {
		d.SetId(repositoryName)
	}

	return append(diags, resourceRepositoryPolicyRead(ctx, d, meta)...)
}

func resourceRepositoryPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findRepositoryPolicyByRepositoryName(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ECR Repository Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Repository Policy (%s): %s", d.Id(), err)
	}

	output := outputRaw.(*ecr.GetRepositoryPolicyOutput)

	policyToSet, err := verify.SecondJSONUnlessEquivalent(d.Get(names.AttrPolicy).(string), aws.ToString(output.PolicyText))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set(names.AttrPolicy, policyToSet)
	d.Set("registry_id", output.RegistryId)
	d.Set("repository", output.RepositoryName)

	return diags
}

func resourceRepositoryPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	log.Printf("[DEBUG] Deleting ECR Repository Policy: %s", d.Id())
	_, err := conn.DeleteRepositoryPolicy(ctx, &ecr.DeleteRepositoryPolicyInput{
		RegistryId:     aws.String(d.Get("registry_id").(string)),
		RepositoryName: aws.String(d.Id()),
	})

	if errs.IsA[*types.RepositoryNotFoundException](err) || errs.IsA[*types.RepositoryPolicyNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ECR Repository Policy (%s): %s", d.Id(), err)
	}

	return diags
}

func findRepositoryPolicyByRepositoryName(ctx context.Context, conn *ecr.Client, repositoryName string) (*ecr.GetRepositoryPolicyOutput, error) {
	input := &ecr.GetRepositoryPolicyInput{
		RepositoryName: aws.String(repositoryName),
	}

	output, err := conn.GetRepositoryPolicy(ctx, input)

	if errs.IsA[*types.RepositoryNotFoundException](err) || errs.IsA[*types.RepositoryPolicyNotFoundException](err) {
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
